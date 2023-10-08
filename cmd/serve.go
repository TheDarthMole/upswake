package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"upsWake/ups"
)

var (
	host       string
	username   string
	password   string
	broadcasts []string
)

func init() {
	serveCmd.Flags().StringVarP(&host, "host", "H", "", "Host of the UPS to connect to")
	serveCmd.Flags().StringVarP(&username, "username", "u", "", "The NUT username to use to connect to the UPS")
	serveCmd.Flags().StringVarP(&password, "password", "p", "", "The NUT password to use to connect to the UPS")
	serveCmd.MarkFlagRequired("host")
	serveCmd.MarkFlagRequired("username")
	serveCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the UPSWake server",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ups.Connect(host, username, password)
		if err != nil {
			log.Panicf("could not connect to UPS: %s", err)
		}
		fmt.Println("we didn't error out!")
		client.Help()

	},
}
