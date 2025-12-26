package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
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

type serveCMD struct {
	logger *zap.SugaredLogger
	fs     afero.Fs
	regoFs afero.Fs
}

func NewServeCommand(ctx context.Context, logger *zap.SugaredLogger, fs, regoFs afero.Fs) *cobra.Command {
	sc := &serveCMD{
		logger: logger,
		fs:     fs,
		regoFs: regoFs,
	}

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
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()
	cmd.SetContext(ctx)

	cfgPath, _ := cmd.Flags().GetString("config")
	certFile, _ := cmd.Flags().GetString("certFile")
	keyFile, _ := cmd.Flags().GetString("keyFile")
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetString("port")
	useSSL, err := cmd.Flags().GetBool("ssl")
	if err != nil {
		return err
	}

	cliArgs, err := config.NewCLIArgs(j.fs, cfgPath, useSSL, certFile, keyFile, host, port)
	if err != nil {
		return err
	}

	viper.InitConfig(j.fs, cliArgs.ConfigFile)

	cfg, err := viper.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("error validating config: %w", err)
	}

	server := api.NewServer(cmd.Context(), j.logger)

	rootHandler := handlers.NewRootHandler(cfg, j.regoFs)
	rootHandler.Register(server.Root())

	serverHandler := handlers.NewServerHandler()
	serverHandler.Register(server.API().Group("/servers"))

	upsWakeHandler := handlers.NewUPSWakeHandler(cfg, j.regoFs)
	upsWakeHandler.Register(server.API().Group("/upswake"))

	var wg sync.WaitGroup

	for _, mapping := range cfg.NutServers {
		for _, target := range mapping.Targets {
			wg.Add(1)
			go j.processTarget(cmd.Context(), &wg, target, cliArgs.URL()+"/api/upswake", cliArgs.TLSConfig)
		}
	}
	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		<-ctx.Done()
		j.logger.Info("Shutting down server")
		if err1 := server.Stop(); err1 != nil {
			j.logger.Warnf("Error stopping server: %v", err1)
		}
	}(ctx, &wg)

	err = server.Start(
		cliArgs.ListenAddress(),
		cliArgs.UseSSL,
		cliArgs.CertFile,
		cliArgs.KeyFile,
	)

	cancel()
	wg.Wait()
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

	client := &http.Client{
		Timeout:   time.Duration(30) * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	for {
		select {
		case <-ctx.Done():
			j.logger.Infof("[%s] Gracefully stopping worker", target.Name)
			return
		case <-ticker.C:
			j.sendWakeRequest(ctx, target, endpoint, client)
		}
	}
}

func (j *serveCMD) sendWakeRequest(ctx context.Context, target config.TargetServer, address string, client *http.Client) {
	body, err := json.Marshal(map[string]string{"mac": target.MAC})
	if err != nil {
		j.logger.Errorf("[%s] Error marshalling JSON: %s", target.Name, err)
		return
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewBuffer(body))
	if err != nil {
		j.logger.Errorf("[%s] Error creating post request: %s", target.Name, err)
		return
	}
	r.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(r)
	if errors.Is(err, context.Canceled) {
		j.logger.Infof("[%s] Gracefully stopping", target.Name)
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		j.logger.Warnf("[%s] Timeout sending post request: %s", target.Name, err)
		return
	}
	if err != nil {
		j.logger.Errorf("[%s] Error sending post request: %s", target.Name, err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			j.logger.Errorf("[%s] Error closing response body: %s", target.Name, err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		j.logger.Errorf("[%s] Error sending post request: %s", target.Name, resp.Status)
		return
	}
}
