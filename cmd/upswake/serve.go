package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	defaultListenHost = "0.0.0.0"
	defaultListenPort = "8080"
)

var (
	fileSystem = afero.NewOsFs()
	regoFiles  = afero.NewBasePathFs(fileSystem, "rules")
)

type jsonCMD struct {
	logger *zap.SugaredLogger
}

func NewServeCommand(ctx context.Context, logger *zap.SugaredLogger) *cobra.Command {
	serve := &jsonCMD{logger: logger}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `Run the UPSWake server and API on the specified port`,
		RunE:  serve.serveCmdRunE,
	}
	serveCmd.SetContext(ctx)
	serveCmd.Flags().StringP("port", "p", defaultListenPort, "Port to listen on")
	serveCmd.Flags().StringP("host", "H", defaultListenHost, "Interface to listen on")
	serveCmd.Flags().BoolP("ssl", "s", false, "Enable SSL (HTTPS)")
	serveCmd.Flags().StringP("certFile", "c", "", "SSL Certificate file (required if SSL is enabled)")
	serveCmd.Flags().StringP("keyFile", "k", "", "SSL Key file (required if SSL is enabled)")
	serveCmd.PersistentFlags().String(
		"config",
		"./config.yaml",
		"The location of config file",
	)
	return serveCmd
}

func (serve *jsonCMD) serveCmdRunE(cmd *cobra.Command, _ []string) error {
	cliArgs, err := config.NewCLIArgs(
		fileSystem,
		cmd.Flag("config").Value.String(),
		cmd.Flag("ssl").Value.String() == "true",
		cmd.Flag("certFile").Value.String(),
		cmd.Flag("keyFile").Value.String(),
		cmd.Flag("host").Value.String(),
		cmd.Flag("port").Value.String(),
	)
	if err != nil {
		return err
	}

	viper.InitConfig(fileSystem, cliArgs.ConfigFile)

	cfg, err := viper.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("error validating config: %w", err)
	}

	server := api.NewServer(cmd.Context(), serve.logger)

	rootHandler := handlers.NewRootHandler(cfg, regoFiles)
	rootHandler.Register(server.Root())

	serverHandler := handlers.NewServerHandler()
	serverHandler.Register(server.API().Group("/servers"))

	upsWakeHandler := handlers.NewUPSWakeHandler(cfg, regoFiles)
	upsWakeHandler.Register(server.API().Group("/upswake"))

	var workerWG sync.WaitGroup

	for _, mapping := range cfg.NutServers {
		for _, target := range mapping.Targets {
			workerWG.Add(1)
			go serve.processTarget(cmd.Context(), &workerWG, target, cliArgs.URL()+"/api/upswake", cliArgs.TLSConfig)
		}
	}
	var shutdownWG sync.WaitGroup
	go func() {
		<-cmd.Context().Done()
		workerWG.Wait()
		serve.logger.Info("Shutting down server")
		_ = server.Stop()
	}()
	shutdownWG.Add(1)
	err = server.Start(
		cliArgs.ListenAddress(),
		cliArgs.UseSSL,
		cliArgs.CertFile,
		cliArgs.KeyFile,
	)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	shutdownWG.Wait()
	return err
}

func (serve *jsonCMD) processTarget(ctx context.Context, wg *sync.WaitGroup, target config.TargetServer, endpoint string, tlsConfig *tls.Config) {
	serve.logger.Infof("[%s] Starting worker", target.Name)
	interval, err := time.ParseDuration(target.Interval)
	if err != nil {
		serve.logger.Fatalf("[%s] Stopping Worker. Could not parse interval: %s", target.Name, err)
		return
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			serve.logger.Infof("[%s] Gracefully stopping worker", target.Name)
			return
		case <-ticker.C:
			serve.sendWakeRequest(ctx, target, endpoint, tlsConfig)
			ticker.Reset(interval)
		}
	}
}

func (serve *jsonCMD) sendWakeRequest(ctx context.Context, target config.TargetServer, address string, tlsConfig *tls.Config) {
	body := []byte(`{"mac":"` + target.MAC + `"}`) // target.Mac is validated in the config
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewBuffer(body))
	if err != nil {
		serve.logger.Errorf("Error creating post request: %s", err)
	}
	r.Header.Set("Content-Type", "application/json")

	if err != nil {
		serve.logger.Fatalf("Error creating TLS configuration: %v", err)
	}
	client := &http.Client{
		Timeout:   time.Duration(30) * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	resp, err := client.Do(r)
	if err != nil {
		serve.logger.Errorf("Error sending post request: %s", err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			serve.logger.Errorf("Error closing response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		serve.logger.Errorf("Error sending post request: %s", resp.Status)
		return
	}
}
