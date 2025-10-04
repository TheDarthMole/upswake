package main

import (
	"github.com/spf13/cobra"
)

var (
	Version    string
	Commit     string
	Date       string
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version information",
		Long:  `Shows the version, commit hash, and build date of the application`,
		Run: func(_ *cobra.Command, _ []string) {
			sugar.Infof("UPSWake version %s, commit %s, built at %s", Version, Commit, Date)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
