package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "UPSWake",
	Short: "UPSWake sends WoL packets based on a UPS's status",
	Long:  `TODO: Add a long description here`, // TODO: Add a long description here
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello World!")
		//err := wol.Wake("a4:ae:11:1e:7d:1c")
		//if err != nil {
		//	fmt.Printf("Error: %v", err)
		//	return
		//}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
