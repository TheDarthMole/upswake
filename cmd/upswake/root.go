package main

import (
	"log"
	"os"

	"github.com/TheDarthMole/UPSWake/internal/network"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	shortAppDesc = "UPSWake sends Wake on LAN packets based on a UPS's status"
	longAppDesc  = `UPSWake sends Wake on LAN packets to target servers

It uses the status of a UPS to determine which servers to wake
using a set of Rego rules defined and the servers in the config file`
)

var (
	Version string
	sugar   *zap.SugaredLogger
	rootCmd = NewRootCommand()
)

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
func Execute() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialise zap logger: %v", err)
	}
	sugar = logger.Sugar()

	bc, err := network.GetAllBroadcastAddresses()

	if err != nil {
		sugar.Panic(err)
		return
	}
	stringBroadcasts := network.IPsToStrings(bc)

	wakeCmd := NewWakeCmd(stringBroadcasts)
	rootCmd.AddCommand(wakeCmd)

	jsonCmd := NewJSONCommand()
	rootCmd.AddCommand(jsonCmd)

	serveCmd := NewServeCommand()
	rootCmd.AddCommand(serveCmd)

	err = rootCmd.Execute()
	if err != nil {
		logger.Debug("Error executing root command: " + err.Error())
		os.Exit(1)
	}
}

func main() {
	Execute()
}
