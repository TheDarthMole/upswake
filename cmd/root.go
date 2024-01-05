package cmd

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
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
		Use:   "upsWake",
		Short: "UPSWake sends WoL packets based on a UPS's status",
		Long:  `TODO: Add a long description here`, // TODO: Add a long description here
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
	} else {
		// Search config in home directory with name ".test" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigType(config.DefaultConfigExt)
		viper.SetConfigName(config.DefaultConfigFile)
	}
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.

	if !util.FileExists(fileSystem, fmt.Sprintf("%s.%s", config.DefaultConfigFile, config.DefaultConfigExt)) {
		defaultConfig := config.CreateDefaultConfig()
		marshalledConfig, err := yaml.Marshal(defaultConfig)
		if err != nil {
			log.Fatalf("Unable to marshal config: %s", err)
		}

		localFS, err := util.GetLocalFS()
		if err != nil {
			log.Fatalf("Unable to get local filesystem: %s", err)
		}
		configFile := fmt.Sprintf("%s.%s", config.DefaultConfigFile, config.DefaultConfigExt)
		if err = util.CreateFile(localFS, configFile, marshalledConfig); err != nil {
			log.Fatalf("Unable to create new config file: %s", err)
		}

		log.Printf("Created new config file at %s", config.DefaultConfigFile)
		os.Exit(0)
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	unmarshalOptions := viper.DecoderConfigOption(func(decoderConfig *mapstructure.DecoderConfig) {
		// This is needed because the decoder defaults to being 'mapstructure' and causes an error
		decoderConfig.TagName = config.DefaultConfigExt
	})

	if err = viper.Unmarshal(&cfg, unmarshalOptions); err != nil {
		log.Fatalf("Unable to unmarshal config: %s", err)
	}
	if err = cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %s", err)
	}
}
