package main

import (
	"bytes"
	"context"
	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"
)

const (
	listenHost        = "localhost"
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
			baseURL := listenScheme + listenHost + ":" + cmd.Flag("port").Value.String()
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

			//server.PrintRoutes()
			sugar.Fatal(server.Start(":" + cmd.Flag("port").Value.String()))
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")
	serveCmd.Flags().StringP("port", "p", defaultListenPort, "Port to listen on, default: "+defaultListenPort)
	serveCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"./config.yaml",
		"config file (default is ./config.yaml)")
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
