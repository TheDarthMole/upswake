package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	_ "golang.org/x/crypto/x509roots/fallback" // Embeds x509root certificates into the binary
)

const (
	defaultListenHost = "0.0.0.0"
	defaultListenPort = "8080"
)

type serveCMD struct {
	logger *slog.Logger
	fs     afero.Fs
	regoFs afero.Fs
}

func NewServeCommand(ctx context.Context, logger *slog.Logger, fs, regoFs afero.Fs) *cobra.Command {
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
			j.logger.Warn("Error stopping server", slog.Any("error", err1))
		}
	}(ctx, &wg)

	err = server.Start(
		j.fs,
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
	j.logger.Info("Starting worker",
		slog.String("workerName", target.Name))
	interval, err := time.ParseDuration(target.Interval)
	if err != nil {
		j.logger.Error("Stopping Worker. Could not parse interval",
			slog.String("workerName", target.Name),
			slog.Any("error", err))
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
			j.logger.Info("Gracefully stopping worker",
				slog.String("workerName", target.Name))
			return
		case <-ticker.C:
			j.sendWakeRequest(ctx, target, endpoint, client)
		}
	}
}

func (j *serveCMD) sendWakeRequest(ctx context.Context, target config.TargetServer, address string, client *http.Client) {
	body, err := json.Marshal(map[string]string{"mac": target.MAC})
	if err != nil {
		j.logger.Error("Error marshalling JSON",
			slog.String("workerName", target.Name),
			slog.Any("error", err))
		return
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewBuffer(body))
	if err != nil {
		j.logger.Error("Error creating post request",
			slog.String("workerName", target.Name),
			slog.Any("error", err))
		return
	}
	r.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(r)
	if errors.Is(err, context.Canceled) {
		j.logger.Info("Gracefully stopping worker",
			slog.String("workerName", target.Name))
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		j.logger.Warn("Timeout sending post request",
			slog.String("workerName", target.Name),
			slog.Any("error", err))
		return
	}
	if err != nil {
		j.logger.Error("Error sending post request",
			slog.String("workerName", target.Name),
			slog.Any("error", err))
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			j.logger.Error("Error closing response body",
				slog.String("workerName", target.Name),
				slog.Any("error", err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		j.logger.Error("Error sending post request",
			slog.String("workerName", target.Name),
			slog.String("statusCode", resp.Status))
		return
	}
}
