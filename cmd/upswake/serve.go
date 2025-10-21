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

type serveCMD struct {
	logger *zap.SugaredLogger
}

func NewServeCommand(ctx context.Context, logger *zap.SugaredLogger) *cobra.Command {
	sc := &serveCMD{logger: logger}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `Run the UPSWake server and API on the specified port`,
		Example: `  upswake serve --port 8080
  upswake serve -p 8080 -H 192.168.1.10
  upswake serve --port 8443 --ssl --certFile /path/to/cert.pem --keyFile /path/to/key.pem
  upswake serve -p 8443 -s -c /path/to/cert.pem -k /path/to/key.pem`,
		RunE: sc.serveCmdRunE,
	}
	serveCmd.SetContext(ctx)
	serveCmd.Flags().StringP("port", "p", defaultListenPort, "Port to listen on")
	serveCmd.Flags().StringP("host", "H", defaultListenHost, "Interface to listen on")
	serveCmd.Flags().BoolP("ssl", "s", false, "Enable SSL (HTTPS)")
	serveCmd.Flags().StringP("certFile", "c", "", "SSL Certificate file (required if SSL is enabled)")
	serveCmd.Flags().StringP("keyFile", "k", "", "SSL Key file (required if SSL is enabled)")
	serveCmd.Flags().String(
		"config",
		"./config.yaml",
		"The location of config file",
	)
	return serveCmd
}

func (j *serveCMD) serveCmdRunE(cmd *cobra.Command, _ []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	certFile, _ := cmd.Flags().GetString("certFile")
	keyFile, _ := cmd.Flags().GetString("keyFile")
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetString("port")
	useSSL, err := cmd.Flags().GetBool("ssl")
	if err != nil {
		return err
	}

	cliArgs, err := config.NewCLIArgs(fileSystem, cfgPath, useSSL, certFile, keyFile, host, port)
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

	server := api.NewServer(cmd.Context(), j.logger)

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
			go j.processTarget(cmd.Context(), &workerWG, target, cliArgs.URL()+"/api/upswake", cliArgs.TLSConfig)
		}
	}
	var shutdownWG sync.WaitGroup
	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		<-cmd.Context().Done()
		workerWG.Wait()
		j.logger.Info("Shutting down server")
		_ = server.Stop()
	}()

	err = server.Start(
		cliArgs.ListenAddress(),
		cliArgs.UseSSL,
		cliArgs.CertFile,
		cliArgs.KeyFile,
	)
	shutdownWG.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (j *serveCMD) processTarget(ctx context.Context, wg *sync.WaitGroup, target config.TargetServer, endpoint string, tlsConfig *tls.Config) {
	defer wg.Done()
	j.logger.Infof("[%s] Starting worker", target.Name)
	interval, err := time.ParseDuration(target.Interval)
	if err != nil {
		j.logger.Errorf("[%s] Stopping Worker. Could not parse interval: %s", target.Name, err)
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			j.logger.Infof("[%s] Gracefully stopping worker", target.Name)
			return
		case <-ticker.C:
			j.sendWakeRequest(ctx, target, endpoint, tlsConfig)
			ticker.Reset(interval)
		}
	}
}

func (j *serveCMD) sendWakeRequest(ctx context.Context, target config.TargetServer, address string, tlsConfig *tls.Config) {
	body := []byte(`{"mac":"` + target.MAC + `"}`) // target.Mac is validated in the config
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewBuffer(body))
	if err != nil {
		j.logger.Errorf("Error creating post request: %s", err)
		return
	}
	r.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout:   time.Duration(30) * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	resp, err := client.Do(r)

	if errors.Is(err, context.Canceled) {
		j.logger.Infof("[%s] Stopping wake request, context cancelled", target.Name)
		return
	}
	if err != nil {
		j.logger.Errorf("Error sending post request: %s", err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			j.logger.Errorf("Error closing response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		j.logger.Errorf("Error sending post request: %s", resp.Status)
		return
	}
}
