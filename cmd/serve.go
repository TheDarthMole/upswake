package cmd

import (
	"context"
	"github.com/TheDarthMole/UPSWake/api"
	"github.com/TheDarthMole/UPSWake/api/handlers"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io/fs"
	"log"
	"os"
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

			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("can't initialize zap logger: %v", err)
			}
			sugar := logger.Sugar()

			server := api.NewServer(ctx, sugar)

			rootHandler := handlers.NewRootHandler()
			rootHandler.Register(server.Root())

			serverHandler := handlers.NewServerHandler()
			serverHandler.Register(server.API().Group("/servers"))

			// TODO: Add UPS handler that uses the new api rather than go routines
			//for _, woLTarget := range cfg.NutServerMappings {
			//	sugar.Infof("Starting worker for %s with interval %s\n", woLTarget.Name, woLTarget.Interval)
			//	go runWorker(ctx, &woLTarget)
			//}

			server.PrintRoutes()
			sugar.Fatal(server.Start(":8080"))
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	regoFiles = os.DirFS("rules")
}

//
//func runWorker(ctx context.Context, nutServerMapping *config.NutServerMapping) {
//	interval, err := time.ParseDuration(nutServerMapping.Targets)
//
//	if err != nil {
//		log.Printf("[%s] Stopping Worker. Could not parse interval: %s\n", nutServerMapping.Name, err)
//		return
//	}
//	ticker := time.NewTicker(interval)
//
//	for {
//		select {
//		case <-ctx.Done():
//			log.Printf("[%s] Stopping worker\n", nutServerMapping.Name)
//			return
//		case <-ticker.C:
//			if err = processWoLTarget(nutServerMapping); err != nil {
//				log.Printf("[%s] Error processing WoL target: %s\n", nutServerMapping.Name, err)
//			}
//			ticker.Reset(interval)
//		}
//	}
//}
//
//func processWoLTarget(woLTarget *config.TargetServer) error {
//	inputJson, err := ups.GetJSON(woLTarget)
//	if err != nil {
//		return err
//	}
//	for _, ruleName := range woLTarget.Rules {
//		log.Printf("[%s] Evaluating rule %s\n", woLTarget.Name, ruleName)
//
//		regoRule, err := util.GetFile(regoFiles, ruleName)
//		if err != nil {
//			return fmt.Errorf("could not get file: %s", err)
//		}
//
//		allowed, err := rego.EvaluateExpression(inputJson, string(regoRule))
//		if err != nil {
//			return fmt.Errorf("could not evaluate expression: %s", err)
//		}
//		log.Printf("[%s] Rule %s evaluated to %t\n", woLTarget.Name, ruleName, allowed)
//
//		if allowed {
//			wolClient := wol.NewWoLClient(*woLTarget)
//
//			if err = wolClient.Wake(); err != nil {
//				return fmt.Errorf("could not send WoL packet: %s", err)
//			}
//			log.Printf("[%s] Sent WoL packet to %s (%s)\n", woLTarget.Name, woLTarget.Name, woLTarget.Mac)
//		}
//	}
//	return nil
//}
