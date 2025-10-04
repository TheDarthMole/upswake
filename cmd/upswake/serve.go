package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	defaultListenHost = "0.0.0.0"
	defaultListenPort = "8080"
)

var (
	cfgFile    string
	regoFiles  afero.Fs
	fileSystem afero.Fs
	serveCmd   = &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `Run the UPSWake server and API on the specified port`,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			viper.InitConfig(cliArgs.ConfigFile)

			cfg, err := viper.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("error validating config: %w", err)
			}

			ctx := context.Background()

			server := api.NewServer(ctx, sugar)

			rootHandler := handlers.NewRootHandler(cfg, regoFiles)
			rootHandler.Register(server.Root())

			serverHandler := handlers.NewServerHandler()
			serverHandler.Register(server.API().Group("/servers"))

			upsWakeHandler := handlers.NewUPSWakeHandler(cfg, regoFiles)
			upsWakeHandler.Register(server.API().Group("/upswake"))

			for _, mapping := range cfg.NutServers {
				for _, target := range mapping.Targets {
					go processTarget(ctx, target, cliArgs.Address()+"/api/upswake", cliArgs.TLSConfig)
				}
			}

			return server.Start(
				fmt.Sprintf("%s:%s", cliArgs.Host.String(), cliArgs.Port),
				cmd.Flag("ssl").Value.String() == "true",
				cmd.Flag("certFile").Value.String(),
				cmd.Flag("keyFile").Value.String(),
			)
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	fileSystem = afero.NewOsFs()
	regoFiles = afero.NewBasePathFs(fileSystem, "rules")
	serveCmd.Flags().StringP("port", "p", defaultListenPort, "Port to listen on")
	serveCmd.Flags().StringP("host", "H", defaultListenHost, "Interface to listen on")
	serveCmd.Flags().BoolP("ssl", "s", false, "Enable SSL (HTTPS)")
	serveCmd.Flags().StringP("certFile", "c", "", "SSL Certificate file (required if SSL is enabled)")
	serveCmd.Flags().StringP("keyFile", "k", "", "SSL Key file (required if SSL is enabled)")
	serveCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"./config.yaml",
		"location of config file")
}

func processTarget(ctx context.Context, target config.TargetServer, endpoint string, tlsConfig *tls.Config) {
	sugar.Infof("[%s] Starting worker", target.Name)
	interval, err := time.ParseDuration(target.Interval)
	if err != nil {
		sugar.Fatalf("[%s] Stopping Worker. Could not parse interval: %s", target.Name, err)
		return
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			sugar.Infof("[%s] Gracefully stopping worker", target.Name)
			return
		case <-ticker.C:
			sendWakeRequest(ctx, target, endpoint, tlsConfig)
			ticker.Reset(interval)
		}
	}
}

func sendWakeRequest(ctx context.Context, target config.TargetServer, address string, tlsConfig *tls.Config) {
	body := []byte(`{"mac":"` + target.MAC + `"}`) // target.Mac is validated in the config
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewBuffer(body))
	if err != nil {
		sugar.Errorf("Error creating post request: %s", err)
	}
	r.Header.Set("Content-Type", "application/json")

	if err != nil {
		sugar.Fatalf("Error creating TLS configuration: %v", err)
	}
	client := &http.Client{
		Timeout:   time.Duration(30) * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	resp, err := client.Do(r)
	if err != nil {
		sugar.Errorf("Error sending post request: %s", err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			sugar.Errorf("Error closing response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		sugar.Errorf("Error sending post request: %s", resp.Status)
		return
	}
}
