package cmd

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/fs"
	"log"
	"os"
)

var (
	cfgFile    string
	cfg        config.Config
	fileSystem fs.FS

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "upswake",
		Short: "UPSWake sends Wake on LAN packets based on a UPS's status",
		Long: `UPSWake sends Wake on LAN packets to target servers

It uses the status of a UPS to determine which servers to wake
using a set of Rego rules defined and the servers in the config file`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Error executing root command: %s", err)
	}
}

func init() {
	fileSystem = os.DirFS(".")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		fmt.Sprintf("config file (default is ./%s%s)", config.DefaultConfigFile, config.DefaultConfigExt))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if !util.FileExists(fileSystem, cfgFile) {
			log.Fatalf("config file %s does not exist", cfgFile)
		}

	} else {
		// Search config in home directory with name ".test" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigType(config.DefaultConfigExt)
		viper.SetConfigName(config.DefaultConfigFile)
	}
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.

	if err := config.CheckCreateConfigFile(fileSystem, "./"+config.DefaultConfigFile+"."+config.DefaultConfigExt); err != nil {
		log.Fatal(err)
	}

	if err := parseConfig(); err != nil {
		log.Fatal(err)
	}
}

func parseConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}

	unmarshalOptions := viper.DecoderConfigOption(func(decoderConfig *mapstructure.DecoderConfig) {
		// This is needed because the decoder defaults to being 'mapstructure' and causes an error
		decoderConfig.TagName = config.DefaultConfigExt
	})

	if err := viper.Unmarshal(&cfg, unmarshalOptions); err != nil {
		return fmt.Errorf("unable to unmarshal config: %s", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	return nil
}
