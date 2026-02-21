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

type serveJob struct {
	ctx         context.Context
	wg          *sync.WaitGroup
	logger      *slog.Logger
	interval    time.Duration
	client      *http.Client
	requestBody []byte
	url         string
}

func newServeJob(ctx context.Context, targetServer *config.TargetServer, tlsConfig *tls.Config, wg *sync.WaitGroup, logger *slog.Logger, endpoint string) (*serveJob, error) {
	jobLogger := logger.With(
		slog.String("type", "serveJob"),
		slog.String("worker_name", targetServer.Name),
	)

	interval, err := time.ParseDuration(targetServer.Interval)
	if err != nil {
		jobLogger.Error("Stopping Worker. Could not parse interval",
			slog.Any("error", err))
		return &serveJob{}, err
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	body, err := json.Marshal(map[string]string{"mac": targetServer.MAC})
	if err != nil {
		jobLogger.Error("Error marshalling JSON",
			slog.Any("error", err))
		return &serveJob{}, err
	}

	url := endpoint + "/api/upswake"

	return &serveJob{
		ctx:         ctx,
		client:      client,
		wg:          wg,
		logger:      jobLogger,
		interval:    interval,
		requestBody: body,
		url:         url,
	}, nil
}

func (j *serveJob) run() {
	j.logger.Info("Starting worker")

	go func() {
		defer j.wg.Done()
		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()

		for {
			select {
			case <-j.ctx.Done():
				j.logger.Info("Gracefully stopping worker")
				return
			case <-ticker.C:
				j.sendWakeRequest()
			}
		}
	}()
}

func (j *serveJob) sendWakeRequest() {
	resp, err := j.client.Post(j.url, "application/json", bytes.NewBuffer(j.requestBody))

	if errors.Is(err, context.Canceled) {
		j.logger.Warn("Context canceled when making request",
			slog.Any("error", err))
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		j.logger.Warn("Timeout sending post request",
			slog.Any("error", err))
		return
	}
	if err != nil {
		j.logger.Error("Error sending post request",
			slog.Any("error", err))
		return
	}

	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body) // Drain body to enable connection reuse
		err := Body.Close()
		if err != nil {
			j.logger.Error("Error closing response body",
				slog.Any("error", err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		j.logger.Error("Error sending post request",
			slog.String("status_code", resp.Status))
		return
	}
}

func NewServeCommand(ctx context.Context, logger *slog.Logger, fs, regoFs afero.Fs) *cobra.Command {
	childLogger := logger.With(
		slog.String("cmd", "serve"),
	)

	sc := &serveCMD{
		logger: childLogger,
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
			job, jobErr := newServeJob(ctx, target, cliArgs.TLSConfig, &wg, j.logger, cliArgs.URL())
			if jobErr != nil {
				cancel()
				wg.Wait()
				return jobErr
			}
			wg.Add(1)
			job.run()
		}
	}

	err = server.Start(
		j.fs,
		cliArgs.ListenAddress(),
		cliArgs.UseSSL,
		cliArgs.CertFile,
		cliArgs.KeyFile,
	)

	j.logger.Info("Server stopped, waiting for workers to finish")
	cancel()
	wg.Wait()
	j.logger.Info("All workers stopped, exiting")

	if err != nil {
		j.logger.Error("Server exited with error", slog.Any("error", err))
		return err
	}
	return nil
}
