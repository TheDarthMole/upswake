package main

import (
	"context"
	"log"
	"os"

	"github.com/TheDarthMole/UPSWake/internal/network"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
func Execute(ctx context.Context, fs, regoFs afero.Fs) int {
	logger, err := zap.NewProduction(zap.WithCaller(false))
	if err != nil {
		log.Fatalf("can't initialise zap logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)
	sugar := logger.Sugar()

	bc, err := network.GetAllBroadcastAddresses()
	if err != nil {
		sugar.Error(err)
		return 1
	}
	rootCmd := NewRootCommand()

	wakeCmd := NewWakeCmd(sugar, bc)
	rootCmd.AddCommand(wakeCmd)

	jsonCmd := NewJSONCommand(sugar)
	rootCmd.AddCommand(jsonCmd)

	serveCmd := NewServeCommand(ctx, sugar, fs, regoFs)
	rootCmd.AddCommand(serveCmd)

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		logger.Error("Error executing root command: " + err.Error())
		return 1
	}
	return 0
}

func main() {
	fs := afero.NewOsFs()
	regoFs := afero.NewBasePathFs(fs, "rules")
	os.Exit(Execute(context.Background(), fs, regoFs))
}
