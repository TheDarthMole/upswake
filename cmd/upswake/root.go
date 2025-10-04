package main

import (
	"log"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	sugar *zap.SugaredLogger
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "upswake",
		Short: "UPSWake sends Wake on LAN packets based on a UPS's status",
		Long: `UPSWake sends Wake on LAN packets to target servers

It uses the status of a UPS to determine which servers to wake
using a set of Rego rules defined and the servers in the config file`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialise zap logger: %v", err)
	}
	sugar = logger.Sugar()
	err = rootCmd.Execute()
	if err != nil {
		log.Fatalf("Error executing root command: %s", err)
	}
}
