package cmd

import (
	"github.com/TheDarthMole/UPSWake/config"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "upsWake",
	Short: "UPSWake sends WoL packets based on a UPS's status",
	Long:  `TODO: Add a long description here`, // TODO: Add a long description here
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".test" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Reading config file from %s", viper.ConfigFileUsed())
	} else {
		log.Fatalf("No config file found: %s", err)
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Println("Unable to unmarshal config:", err)
		os.Exit(1)
	}
	if err = cfg.IsValid(); err != nil {
		log.Println("Invalid config:", err)
		os.Exit(1)
	}
}
