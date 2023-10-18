package cmd

import (
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"log"
	"os"
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
		cwd := util.GetCurrentDirectory()
		cfg := config.CreateDefaultConfig()
		marshalledConfig, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("Unable to marshal config: %s", err)
		}

		err = util.CreateFile(cwd+"/config.yaml", marshalledConfig)
		if err != nil {
			log.Fatalf("Unable to create new config file: %s", err)
		}
		log.Printf("Created new config file at %s", cwd+"/config.yaml")
		os.Exit(0)
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Unable to unmarshal config: %s", err)
	}
	if err = cfg.IsValid(); err != nil {
		log.Fatalf("Invalid config: %s", err)
	}
}
