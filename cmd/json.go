package cmd

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/ups"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Retrieve JSON from a NUT server",
	Long: `Retrieve JSON from a NUT server and print it to stdout

This is useful for testing the connection to a NUT server
and for creating rego rules for a WoL target`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.Fatalf("could not get port: %s", err)
			return
		}
		nutServer := config.NutServer{
			Name: "test",
			Host: cmd.Flag("host").Value.String(),
			Port: port,
			Credentials: config.NutCredentials{
				Username: cmd.Flag("username").Value.String(),
				Password: cmd.Flag("password").Value.String(),
			},
		}

		ups, err := ups.GetJSON(&nutServer)
		if err != nil {
			log.Fatalf("failed to get JSON: %s", err)
			return
		}
		fmt.Println(ups)
	},
}

func init() {
	rootCmd.AddCommand(jsonCmd)
	jsonCmd.Flags().StringP("username", "u", "", "Username for the NUT server")
	jsonCmd.Flags().StringP("password", "p", "", "Password for the NUT server")
	jsonCmd.Flags().StringP("host", "H", "", "Host address of the NUT server")
	jsonCmd.Flags().IntP("port", "P", 9, "Port number of the NUT server")
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
