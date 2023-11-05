/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/spf13/cobra"
	"log"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		wolTarget := config.WoLTarget{
			Name: "test",
			NutServer: config.NutServer{
				Host: cmd.Flag("host").Value.String(),
				Credentials: config.Credentials{
					Username: cmd.Flag("username").Value.String(),
					Password: cmd.Flag("password").Value.String(),
				},
			},
		}

		ups, err := util.GetJSON(&wolTarget)
		if err != nil {
			log.Fatalf("failed to get JSON: %s", err)
			return
		}
		fmt.Println(ups)
	},
}

func init() {
	rootCmd.AddCommand(jsonCmd)
	jsonCmd.Flags().StringP("username", "u", "", "MAC address of the computer to wake")
	jsonCmd.Flags().StringP("password", "p", "", "MAC address of the computer to wake")
	jsonCmd.Flags().StringP("host", "H", "", "MAC address of the computer to wake")
	jsonCmd.Flags().StringP("port", "P", "9", "MAC address of the computer to wake")
	jsonCmd.MarkFlagRequired("username")
	jsonCmd.MarkFlagRequired("password")
	jsonCmd.MarkFlagRequired("host")
}
