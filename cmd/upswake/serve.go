package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/rules"
	directups "github.com/TheDarthMole/UPSWake/internal/infrastructure/ups/direct"
	"github.com/TheDarthMole/UPSWake/internal/worker"
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
	ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
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

	configRepo := viper.NewConfigLoader(j.fs, cliArgs.ConfigFile)

	cfg, err := configRepo.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	ruleRepo, err := rules.NewPreparedRepository(j.regoFs)
	if err != nil {
		return fmt.Errorf("error compiling rego rules: %w", err)
	}

	upsRepo := directups.NewDirectRepository()

	server := api.NewServer(cmd.Context(), j.logger)

	rootHandler := handlers.NewRootHandler(cfg, j.regoFs)
	rootHandler.Register(server.Root())

	serverHandler := handlers.NewServerHandler()
	serverHandler.Register(server.API().Group("/servers"))

	upsWakeHandler := handlers.NewUPSWakeHandler(cfg, upsRepo, ruleRepo)
	upsWakeHandler.Register(server.API().Group("/upswake"))

	workerPool, err := worker.NewWorkerPool(ctx, cfg, cliArgs.TLSConfig, j.logger, fmt.Sprintf("%s/api/upswake", cliArgs.URL()))
	if err != nil {
		return fmt.Errorf("error creating worker pool: %w", err)
	}
	workerPool.Start()

	err = server.Start(
		j.fs,
		cliArgs.ListenAddress(),
		cliArgs.UseSSL,
		cliArgs.CertFile,
		cliArgs.KeyFile,
	)

	j.logger.Info("Server stopped, waiting for workers to finish")
	cancel()
	workerPool.Wait()
	j.logger.Info("All workers stopped, exiting")

	if err != nil {
		j.logger.Error("Server exited with error", slog.Any("error", err))
		return err
	}
	return nil
}
