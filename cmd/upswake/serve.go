package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/spf13/cobra"
)

const (
	defaultListenHost = "0.0.0.0"
	defaultListenPort = "8080"
	listenScheme      = "http://"
)

var (
	cfgFile   string
	regoFiles fs.FS
	serveCmd  = &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `Run the UPSWake server and API on the specified port`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := viper.Load()
			if err != nil {
				sugar.Fatal("Error loading config", err)
			}

			listenHost := net.ParseIP(cmd.Flag("host").Value.String())
			if listenHost == nil {
				sugar.Fatalf("Invalid listen host IP address: %s", cmd.Flag("host").Value.String())
			}
			listenPort, err := strconv.Atoi(cmd.Flag("port").Value.String())
			if err != nil || listenPort <= 0 || listenPort > 65535 {
				sugar.Fatalf("Invalid listen port %s", cmd.Flag("port").Value.String())
			}

			baseURL := fmt.Sprintf("%s%s:%d", listenScheme, listenHost.String(), listenPort)
			if listenHost.IsUnspecified() {
				baseURL = fmt.Sprintf("%s127.0.0.1:%d", listenScheme, listenPort)
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
					go processTarget(ctx, target, baseURL+"/api/upswake")
				}
			}

			sugar.Fatal(server.Start(fmt.Sprintf("%s:%d", listenHost.String(), listenPort)))
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")
	serveCmd.Flags().StringP("port", "p", defaultListenPort, "Port to listen on")
	serveCmd.Flags().StringP("host", "H", defaultListenHost, "Interface to listen on")
	serveCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"./config.yaml",
		"location of config file")
}

func processTarget(ctx context.Context, target config.TargetServer, endpoint string) {
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
			sendWakeRequest(ctx, target, endpoint)
			ticker.Reset(interval)
		}
	}
}

func sendWakeRequest(ctx context.Context, target config.TargetServer, address string) {
	body := []byte(`{"mac":"` + target.MAC + `"}`) // target.Mac is validated in the config
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewBuffer(body))
	if err != nil {
		sugar.Errorf("Error creating post request: %s", err)
	}
	r.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Duration(30) * time.Second}
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
