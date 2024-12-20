package main

import (
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/ups"
	"github.com/spf13/cobra"
	"os"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Retrieve JSON from a NUT server",
	Long: `Retrieve JSON from a NUT server and print it to stdout

This is useful for testing the connection to a NUT server
and for creating rego rules for waking a target`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			sugar.Fatalf("could not get port: %s", err)
			return
		}
		nutServer := entity.NutServer{
			Name:     "test",
			Host:     cmd.Flag("host").Value.String(),
			Port:     port,
			Username: cmd.Flag("username").Value.String(),
			Password: cmd.Flag("password").Value.String(),
		}

		ups, err := ups.GetJSON(&nutServer)
		if err != nil {
			sugar.Fatalf("failed to get JSON: %s", err)
			return
		}
		sugar.Info(ups)
	},
}

func init() {
	rootCmd.AddCommand(jsonCmd)
	jsonCmd.Flags().StringP("username", "u", "", "Username for the NUT server")
	jsonCmd.Flags().StringP("password", "p", "", "Password for the NUT server")
	jsonCmd.Flags().StringP("host", "H", "", "Host address of the NUT server")
	jsonCmd.Flags().IntP("port", "P", entity.DefaultWoLPort, "Port number of the NUT server")
	if err := jsonCmd.MarkFlagRequired("username"); err != nil {
		_ = jsonCmd.Usage()
		os.Exit(1)
	}
	if err := jsonCmd.MarkFlagRequired("password"); err != nil {
		_ = jsonCmd.Usage()
		os.Exit(1)
	}
	if err := jsonCmd.MarkFlagRequired("host"); err != nil {
		_ = jsonCmd.Usage()
		os.Exit(1)
	}
}
