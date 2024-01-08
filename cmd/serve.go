package cmd

import (
	"bytes"
	"context"
	"github.com/TheDarthMole/UPSWake/api"
	"github.com/TheDarthMole/UPSWake/api/handlers"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	listenHost   = "localhost"
	listenPort   = ":8080"
	listenScheme = "http://"
	localURL     = listenScheme + listenHost + listenPort
)

var (
	regoFiles fs.FS
	serveCmd  = &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `All software has versions. This is Hugo's`,
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			ctx := context.Background()
			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can't initialize zap logger: %v", err)
			}

			sugar := logger.Sugar()
			server := api.NewServer(ctx, sugar)

			rootHandler := handlers.NewRootHandler()
			rootHandler.Register(server.Root())

			serverHandler := handlers.NewServerHandler()
			serverHandler.Register(server.API().Group("/servers"))

			upsWakeHandler := handlers.NewUPSWakeHandler(&cfg, regoFiles)
			upsWakeHandler.Register(server.API().Group("/upswake"))

			for _, mapping := range cfg.NutServerMappings {
				for _, target := range mapping.Targets {
					go processTarget(ctx, sugar, target)
				}
			}

			//server.PrintRoutes()
			sugar.Fatal(server.Start(listenPort))
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")
}

func processTarget(ctx context.Context, sugar *zap.SugaredLogger, target config.TargetServer) {
	sugar.Infof("[%s] Starting worker", target.Name)
	interval, err := time.ParseDuration(target.Config.Interval)
	if err != nil {
		sugar.Fatalf("[%s] Stopping Worker. Could not parse interval: %s", target.Name, err)
		return
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			sugar.Infof("[%s] Gracefully stopping worker\n", target.Name)
			return
		case <-ticker.C:
			sendWakeRequest(ctx, target)
			ticker.Reset(interval)
		}
	}
}

func sendWakeRequest(ctx context.Context, target config.TargetServer) {
	body := []byte(`{"mac":"` + target.Mac + `"}`)
	r, err := http.NewRequestWithContext(ctx, "POST", localURL+"/api/upswake", bytes.NewBuffer(body))
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
