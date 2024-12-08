package viper

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"log"
)

const (
	DefaultConfigFile = "config.yaml"
)

var (
	fileSystem          = afero.NewOsFs()
	config              = entity.Config{}
	configFilePath      string
	ErrorConfigNotFound = fmt.Errorf("the config at '%s' was not found", DefaultConfigFile)
	DefaultConfig       = Config{
		NutServers: []NutServer{
			{
				Name:     "NUT Server 1",
				Host:     "192.168.1.13",
				Port:     entity.DefaultNUTServerPort,
				Username: "",
				Password: "",
				Targets: []TargetServer{
					{
						Name:      "NAS 1",
						MAC:       "00:00:00:00:00:00",
						Broadcast: "192.168.1.255",
						Port:      entity.DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"80percentOn.rego",
						},
					},
				},
			},
		},
	}
)

func init() {
	viper.SetEnvPrefix("UPSWAKE")
	viper.AutomaticEnv() // read in environment variables that match
	configFilePath = DefaultConfigFile
	if viper.GetString("CONFIG_FILE") != "" {
		log.Printf("Loading config file from environment")
		configFilePath = viper.GetString("CONFIG_FILE")
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		if _, err := load(fileSystem, configFilePath); err != nil {
			log.Fatal(err)
		}
	})
}

func Load() (*entity.Config, error) {
	return load(fileSystem, configFilePath)
}

func load(fs afero.Fs, configFile string) (*entity.Config, error) {
	viper.SetFs(fs)
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return &entity.Config{}, ErrorConfigNotFound
		}
		// Config file was found but another error was produced
		return &entity.Config{}, err
	}

	loadConfig := Config{}
	if err := viper.Unmarshal(&loadConfig); err != nil {
		return &entity.Config{}, err
	}
	entityConfig := *fromFileConfig(&loadConfig)

	if err := entityConfig.Validate(); err != nil {
		return &entity.Config{}, err
	}
	config = entityConfig
	return &config, nil
}

func CreateDefaultConfig() (*entity.Config, error) {
	return fromFileConfig(&DefaultConfig), nil
}
