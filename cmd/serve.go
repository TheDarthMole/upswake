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

var regoFiles fs.FS

func init() {
	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")

	cobra.OnInitialize(initConfig)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the UPSWake server",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {

		for _, wakeHost := range cfg.WakeHosts {
			nutHost := cfg.GetHostConfig(wakeHost.NutHost.Name)
			nutCreds := nutHost.GetCredentials(wakeHost.NutHost.Username)

			log.Printf("Connecting to NUT server %s as %s\n", nutHost.Host, nutCreds.Username)

			client, err := ups.Connect(nutHost.Address(), nutCreds.Username, nutCreds.Password)
			if err != nil {
				log.Fatalf("could not connect to UPS: %s", err)
			}

			log.Println("Connected to UPS")

			inputJson, err := client.ToJson()
			if err != nil {
				log.Fatalf("could not get UPS list: %s", err)
			}

			for _, ruleName := range wakeHost.Rules {

				log.Printf("Evaluating rule %s\n", ruleName)

				regoRule, err := util.GetFile(regoFiles, ruleName)
				if err != nil {
					log.Fatalf("could not get file: %s", err)
				}

				allowed, err := rego.EvaluateExpression(inputJson, string(regoRule))
				if err != nil {
					log.Fatalf("could not evaluate expression: %s", err)
				}

				log.Printf("Rule %s evaluated to %t\n", ruleName, allowed)
			}
		}
	},
}
