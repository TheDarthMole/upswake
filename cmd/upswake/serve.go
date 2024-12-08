package main

import (
	"bytes"
	"context"
	"github.com/TheDarthMole/UPSWake/internal/api"
	"github.com/TheDarthMole/UPSWake/internal/api/handlers"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io/fs"
	"log"
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
			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can't initialize zap logger: %v", err)
			}

			cfg, err := viper.Load()
			if err != nil {
				logger.Fatal("Error loading config", zap.Error(err))
			}
			baseURL := listenScheme + listenHost + ":" + cmd.Flag("port").Value.String()
			ctx := context.Background()

			sugar := logger.Sugar()
			server := api.NewServer(ctx, sugar)

			rootHandler := handlers.NewRootHandler()
			rootHandler.Register(server.Root())

			serverHandler := handlers.NewServerHandler()
			serverHandler.Register(server.API().Group("/servers"))

			upsWakeHandler := handlers.NewUPSWakeHandler(cfg, regoFiles)
			upsWakeHandler.Register(server.API().Group("/upswake"))

			for _, mapping := range cfg.NutServers {
				for _, target := range mapping.Targets {
					go processTarget(ctx, sugar, target, baseURL+"/api/upswake")
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

func processTarget(ctx context.Context, sugar *zap.SugaredLogger, target config.TargetServer, endpoint string) {
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
		log.Fatalf("Error creating post request: %s", err)
	}
	r.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Fatalf("Error sending post request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error sending post request: %s", resp.Status)
	}
}
