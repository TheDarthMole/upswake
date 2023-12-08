package cmd

import (
	"context"
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

var (
	regoFiles fs.FS
	serveCmd  = &cobra.Command{
		Use:   "serve",
		Short: "Run the UPSWake server",
		Long:  `All software has versions. This is Hugo's`,
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			ctx := context.Background()

			for _, woLTarget := range cfg.WoLTargets {
				log.Printf("Starting worker for %s\n", woLTarget.Name)
				go runWorker(ctx, &woLTarget)
			}

			select {}
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")
}

func runWorker(ctx context.Context, woLTarget *config.WoLTarget) {
	for {
		interval, _ := time.ParseDuration(woLTarget.Interval)

		ticker := time.NewTicker(interval)
		select {
		case <-ctx.Done():
			// TODO: this may not be the best way to stop a goroutine
			log.Printf("[%s] Stopping worker\n", woLTarget.Name)
			return
		case <-ticker.C:
			err := processWoLTarget(woLTarget)
			if err != nil {
				// TODO: this may cause a race condition
				log.Printf("[%s] Error processing WoL target: %s\n", woLTarget.Name, err)
			}
		}
	}
}

func processWoLTarget(woLTarget *config.WoLTarget) error {
	inputJson, err := ups.GetJSON(woLTarget)
	if err != nil {
		return err
	}
	for _, ruleName := range woLTarget.Rules {
		log.Printf("[%s] Evaluating rule %s\n", woLTarget.Name, ruleName)

		regoRule, err := util.GetFile(regoFiles, ruleName)
		if err != nil {
			return fmt.Errorf("could not get file: %s", err)
		}

		allowed, err := rego.EvaluateExpression(inputJson, string(regoRule))
		if err != nil {
			return fmt.Errorf("could not evaluate expression: %s", err)
		}
		log.Printf("[%s] Rule %s evaluated to %t\n", woLTarget.Name, ruleName, allowed)

		if allowed {
			wolClient := wol.NewWoLClient(*woLTarget)

			if err = wolClient.Wake(); err != nil {
				return fmt.Errorf("could not send WoL packet: %s", err)
			}
			log.Printf("[%s] Sent WoL packet to %s (%s)\n", woLTarget.Name, woLTarget.Name, woLTarget.Mac)
		}
	}
	return nil
}
