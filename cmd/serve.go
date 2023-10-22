package cmd

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/rego"
	"github.com/TheDarthMole/UPSWake/ups"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/TheDarthMole/UPSWake/wol"
	"github.com/spf13/cobra"
	"io/fs"
	"log"
	"os"
	"time"
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
		for {
			for _, woLTarget := range cfg.WoLTargets {
				err := processWoLTarget(&woLTarget)
				if err != nil {
					log.Printf("Error processing WoL target %s: %s\n", woLTarget.Name, err)
				}
			}
			// TODO: Make this configurable
			time.Sleep(15 * time.Second)
		}
	},
}

func getJSON(woLTarget *config.WoLTarget) (string, error) {
	ns := woLTarget.NutServer
	log.Printf("Connecting to NUT server %s as %s\n", ns.Host, ns.Credentials.Username)
	client, err := ups.Connect(ns.Host, ns.GetPort(), ns.Credentials.Username, ns.Credentials.Password)
	if err != nil {
		return "", fmt.Errorf("could not connect to NUT server: %s", err)
	}
	defer client.Disconnect()
	log.Println("Getting JSON from NUT server")

	inputJson, err := client.ToJson()
	if err != nil {
		return "", fmt.Errorf("could not get UPS list: %s", err)
	}
	return inputJson, nil
}

func processWoLTarget(woLTarget *config.WoLTarget) error {
	inputJson, err := getJSON(woLTarget)
	if err != nil {
		return err
	}
	for _, ruleName := range woLTarget.Rules {
		log.Printf("Evaluating rule %s\n", ruleName)

		regoRule, err := util.GetFile(regoFiles, ruleName)
		if err != nil {
			return fmt.Errorf("could not get file: %s", err)
		}

		allowed, err := rego.EvaluateExpression(inputJson, string(regoRule))
		if err != nil {
			return fmt.Errorf("could not evaluate expression: %s", err)
		}
		log.Printf("Rule %s evaluated to %t\n", ruleName, allowed)

		if allowed {
			wolClient := wol.NewWoLClient(*woLTarget)

			if err = wolClient.Wake(); err != nil {
				return fmt.Errorf("could not send WoL packet: %s", err)
			}
			log.Printf("Sent WoL packet to %s (%s)\n", woLTarget.Name, woLTarget.Mac)
		}
	}
	return nil
}
