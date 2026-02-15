package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/TheDarthMole/UPSWake/internal/network"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	shortAppDesc = "UPSWake sends Wake on LAN packets based on a UPS's status"
	longAppDesc  = `UPSWake sends Wake on LAN packets to target servers

It uses the status of a UPS to determine which servers to wake
using a set of Rego rules defined and the servers in the config file`
)

var Version string

func NewRootCommand() *cobra.Command {
	// represents the base command when called without any subcommands
	return &cobra.Command{
		Use:     "upswake",
		Short:   shortAppDesc,
		Long:    longAppDesc,
		Version: Version,
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context, fs, regoFs afero.Fs, logDestination io.Writer) int {
	handler := slog.NewJSONHandler(logDestination, nil)
	logger := slog.New(handler)

	bc, err := network.GetAllBroadcastAddresses()
	if err != nil {
		logger.Error(
			"error getting broadcast addresses",
			slog.String("cmd", "root"),
			slog.Any("error", err),
		)
		return 1
	}
	rootCmd := NewRootCommand()

	wakeCmd := NewWakeCmd(logger, bc)
	rootCmd.AddCommand(wakeCmd)

	jsonCmd := NewJSONCommand(logger)
	rootCmd.AddCommand(jsonCmd)

	serveCmd := NewServeCommand(ctx, logger, fs, regoFs)
	rootCmd.AddCommand(serveCmd)

	healthCheckCmd := NewHealthCheckCommand(logger)
	serveCmd.AddCommand(healthCheckCmd)

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		logger.Error(
			"Error executing root command",
			slog.String("cmd", "root"),
			slog.Any("error", err),
		)
		return 1
	}
	return 0
}

func main() {
	fs := afero.NewOsFs()
	regoFs := afero.NewBasePathFs(fs, "rules")
	os.Exit(Execute(context.Background(), fs, regoFs, os.Stdout))
}
