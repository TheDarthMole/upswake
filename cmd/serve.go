package cmd

import (
	"github.com/TheDarthMole/UPSWake/rego"
	"github.com/TheDarthMole/UPSWake/ups"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/spf13/cobra"
	"io/fs"
	"log"
	"os"
)

var regoFiles fs.FS

func init() {
	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")
	initConfig()

}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the UPSWake server",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, woLTarget := range cfg.WoLTargets {
			ns := woLTarget.NutServer
			log.Printf("Connecting to NUT server %s as %s\n", ns.Host, ns.Credentials.Username)
			client, err := ups.Connect(ns.Host, ns.GetPort(), ns.Credentials.Username, ns.Credentials.Password)
			if err != nil {
				log.Fatalf("could not connect to NUT server: %s", err)
			}

			log.Println("Getting JSON from NUT server")

			inputJson, err := client.ToJson()
			if err != nil {
				log.Fatalf("could not get UPS list: %s", err)
			}

			for _, ruleName := range woLTarget.Rules {

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
