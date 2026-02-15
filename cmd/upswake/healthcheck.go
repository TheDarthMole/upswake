package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var (
	ErrHealthCheckFailed = errors.New("healthcheck failed")
	ErrMakingRequest     = errors.New("error making request")
)

type healthCheck struct {
	logger *slog.Logger
}

func NewHealthCheckCommand(logger *slog.Logger) *cobra.Command {
	childLogger := logger.With(
		slog.String("cmd", "healthcheck"),
	)

	hc := healthCheck{logger: childLogger}

	healthcheckCmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "A health check for upswake server",
		Long:  "Queries the /health endpoint of the upswake API",
		RunE:  hc.HealthCheckRunE,
	}

	healthcheckCmd.Flags().StringP("host", "H", "localhost", "Host address of the UPSWake server")
	healthcheckCmd.Flags().StringP("port", "p", defaultListenPort, "Port the UPSWake server is listening on")
	healthcheckCmd.Flags().BoolP("ssl", "s", false, "Enable SSL (HTTPS)")

	return healthcheckCmd
}

func (h *healthCheck) HealthCheckRunE(cmd *cobra.Command, _ []string) error {
	h.logger.Info("Checking health")
	protocol := "http"

	ssl, _ := cmd.Flags().GetBool("ssl")
	if ssl {
		protocol = "https"
	}

	healthURL := fmt.Sprintf("%s://%s:%s/health", protocol, cmd.Flag("host").Value.String(), cmd.Flag("port").Value.String())
	h.logger.Debug("Checking url", slog.String("url", healthURL))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // TODO: This could be changed to allow a trusted cert, but this is fine for now
		},
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		h.logger.Error(
			"error making healthcheck request",
			slog.String("url", healthURL),
			slog.Any("error", err))
		return errors.Join(ErrHealthCheckFailed, ErrMakingRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("healthcheck status not ok",
			slog.String("url", healthURL),
			slog.String("status", resp.Status))
		return fmt.Errorf("%w: %s", ErrHealthCheckFailed, resp.Status)
	}
	h.logger.Info("health check passed")
	return nil
}
