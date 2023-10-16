package cmd

import (
	"github.com/spf13/cobra"
	"io/fs"
	"log"
	"os"
	"upsWake/rego"
	"upsWake/ups"
	"upsWake/util"
)

var (
	host, username, password string
	broadcasts               []string
	regoFiles                fs.FS
)

func init() {
	serveCmd.Flags().StringVarP(&host, "host", "H", "", "Host of the UPS to connect to")
	serveCmd.Flags().StringVarP(&username, "username", "u", "", "The NUT username to use to connect to the UPS")
	serveCmd.Flags().StringVarP(&password, "password", "p", "", "The NUT password to use to connect to the UPS")
	serveCmd.MarkFlagRequired("host")
	serveCmd.MarkFlagRequired("username")
	serveCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")
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

		inputJson, err := client.ToJson()
		if err != nil {
			log.Panicf("could not get UPS list: %s", err)
		}

		files, err := util.ListFiles(regoFiles, ".")
		if err != nil {
			log.Panicf("could not list files: %s", err)
		}
		log.Println(files)
		regoRule, err := util.GetFile(regoFiles, "80percentOn.rego")
		if err != nil {
			log.Fatalf("could not get file: %s", err)
		}

		allowed, err := rego.EvaluateExpression(inputJson, string(regoRule))
		if err != nil {
			log.Panicf("could not evaluate expression: %s", err)
		}
		log.Printf("Allowed: %t", allowed)

		client.Help()

	},
}
