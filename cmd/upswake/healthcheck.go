package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ErrHealthCheckFailed = errors.New("healthcheck failed")
	ErrMakingRequest     = errors.New("error making request")
)

type healthCheck struct {
	logger *zap.SugaredLogger
}

func NewHealthCheckCommand(logger *zap.SugaredLogger) *cobra.Command {
	hc := healthCheck{logger: logger}

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
	h.logger.Infoln("Checking health")
	protocol := "http"

	ssl, _ := cmd.Flags().GetBool("ssl")
	if ssl {
		protocol = "https"
	}

	healthURL := fmt.Sprintf("%s://%s:%s/health", protocol, cmd.Flag("host").Value.String(), cmd.Flag("port").Value.String())
	h.logger.Debugf("Checking %s", healthURL)

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
		h.logger.Errorw(ErrHealthCheckFailed.Error(), "url", healthURL, "err", err)
		return errors.Join(ErrHealthCheckFailed, ErrMakingRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorw(ErrHealthCheckFailed.Error(), "url", healthURL, "status", resp.Status)
		return fmt.Errorf("%w: %s", ErrHealthCheckFailed, resp.Status)
	}
	h.logger.Infof("health check passed")
	return nil
}
