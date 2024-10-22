package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/TheDarthMole/UPSWake/api"
	"github.com/TheDarthMole/UPSWake/api/handlers"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	cfgFile    string
	cfg        config.Config
	regoFiles  fs.FS
	fileSystem fs.FS
	serveCmd   = &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `Run the UPSWake server and API on the specified port`,
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			baseURL := listenScheme + listenHost + ":" + cmd.Flag("port").Value.String()
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
	fileSystem = os.DirFS(".")
	regoFiles = os.DirFS("rules")
	serveCmd.Flags().StringP("port", "p", defaultListenPort, "Port to listen on, default: "+defaultListenPort)
	serveCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		fmt.Sprintf("%s.%s", config.DefaultConfigName, config.DefaultConfigExt),
		fmt.Sprintf("config file (default is ./%s%s)", config.DefaultConfigName, config.DefaultConfigExt))
}

func processTarget(ctx context.Context, sugar *zap.SugaredLogger, target config.TargetServer, endpoint string) {
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
			sugar.Infof("[%s] Gracefully stopping worker", target.Name)
			return
		case <-ticker.C:
			sendWakeRequest(ctx, target, endpoint)
			ticker.Reset(interval)
		}
	}
}

func sendWakeRequest(ctx context.Context, target config.TargetServer, address string) {
	body := []byte(`{"mac":"` + target.Mac + `"}`) // target.Mac is validated in the config
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

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType(config.DefaultConfigExt)

	if cfgFile != config.DefaultConfigFile {
		// Use config file from the flag.
		if !util.FileExists(fileSystem, cfgFile) {
			log.Fatalf("config file %s does not exist", cfgFile)
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Ensure a config file exists, if not then create one
	if err := config.CheckCreateConfigFile(fileSystem, viper.ConfigFileUsed()); err != nil {
		log.Fatal(err)
	}

	if err := parseConfig(); err != nil {
		log.Fatal(err)
	}

	viper.OnConfigChange(func(in fsnotify.Event) {
		if err := parseConfig(); err != nil {
			log.Fatal(err)
		}
	})
}

func parseConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}

	unmarshalOptions := viper.DecoderConfigOption(func(decoderConfig *mapstructure.DecoderConfig) {
		// This is needed because the decoder defaults to being 'mapstructure' and causes an error
		decoderConfig.TagName = config.DefaultConfigExt
	})

	if err := viper.Unmarshal(&cfg, unmarshalOptions); err != nil {
		return fmt.Errorf("unable to unmarshal config: %s", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	return nil
}
