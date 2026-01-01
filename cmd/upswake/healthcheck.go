package main

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

	if cmd.Flag("ssl").Changed {
		protocol = "https"
	}

	healthURL := fmt.Sprintf("%s://%s:%s/health", protocol, cmd.Flag("host").Value.String(), cmd.Flag("port").Value.String())
	h.logger.Debugf("Checking %s", healthURL)
	resp, err := http.Get(healthURL) //nolint:gosec // G107: Potential HTTP request made with variable url
	if err != nil {
		h.logger.Errorw("health check failed", "url", healthURL, "err", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorw("health check failed", "url", healthURL, "status", resp.Status)
		return fmt.Errorf("health check failed: %s", resp.Status)
	}
	h.logger.Infof("health check passed")
	return nil
}
