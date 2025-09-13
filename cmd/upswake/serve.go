package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
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
	cfgFile, baseURL string
	tlsConfig        *tls.Config
	listenHost       net.IP
	listenPort       int
	listenScheme     = "http://"
	regoFiles afero.Fs
	serveCmd  = &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `Run the UPSWake server and API on the specified port`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag("ssl").Value.String() == "true" {
				listenScheme = "https://"
				if cmd.Flag("certFile").Value.String() == "" || cmd.Flag("keyFile").Value.String() == "" {
					return fmt.Errorf("SSL is enabled but certFile or keyFile is not set")
				}
				err := errors.New("")
				tlsConfig, err = x509Cert(cmd.Flag("certFile").Value.String())
				if err != nil {
					return fmt.Errorf("error loading SSL certificate: %s", err)
				}
			}

			listenHost = net.ParseIP(cmd.Flag("host").Value.String())
			if listenHost == nil {
				return fmt.Errorf("invalid listen host IP address: %s", cmd.Flag("host").Value.String())
			}
			err := error(nil)
			listenPort, err = strconv.Atoi(cmd.Flag("port").Value.String())
			if err != nil || listenPort <= 0 || listenPort > 65535 {
				return fmt.Errorf("invalid listen port %s", cmd.Flag("port").Value.String())
			}

			baseURL = fmt.Sprintf("%s%s:%d", listenScheme, listenHost.String(), listenPort)
			if listenHost.IsUnspecified() {
				baseURL = fmt.Sprintf("%s127.0.0.1:%d", listenScheme, listenPort)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := viper.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %s", err)
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("error validating config: %s", err)
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

			return server.Start(
				fmt.Sprintf("%s:%d", listenHost.String(), listenPort),
				cmd.Flag("ssl").Value.String() == "true",
				cmd.Flag("certFile").Value.String(),
				cmd.Flag("keyFile").Value.String(),
			)
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	fileSystem := afero.NewOsFs()
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

func x509Cert(certPath string) (*tls.Config, error) {
	certFile, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	// Decode the PEM certificate
	data, _ := pem.Decode(certFile)
	if data == nil {
		return nil, errors.New("failed to parse PEM certificate")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(data.Bytes)
	if err != nil {
		return nil, err
	}

	conf := &tls.Config{}
	conf.RootCAs = x509.NewCertPool()
	conf.RootCAs.AddCert(cert)

	return conf, nil
}

func sendWakeRequest(ctx context.Context, target config.TargetServer, address string) {
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
		Transport: &http.Transport{TLSClientConfig: tlsConfig}}
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
