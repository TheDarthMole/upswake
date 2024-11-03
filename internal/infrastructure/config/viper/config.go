package viper

import (
	"errors"
	"fmt"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"log"
)

const (
	DefaultConfigName = "config"
	DefaultConfigType = "yaml"
	DefaultConfigPath = "."
	DefaultNUTPort    = 3493
	DefaultWoLPort    = 9
)

var (
	fileSystem          = afero.NewOsFs()
	config              = entity.Config{}
	ErrorConfigNotFound = errors.New(fmt.Sprintf("the config at %s/%s.%s was not found", DefaultConfigPath, DefaultConfigName, DefaultConfigType))
	DefaultConfig       = Config{
		NutServers: []NutServer{
			{
				Name:     "NUT Server 1",
				Host:     "192.168.1.13",
				Port:     DefaultNUTPort,
				Username: "",
				Password: "",
				Targets: []TargetServer{
					{
						Name:      "NAS 1",
						MAC:       "00:00:00:00:00:00",
						Broadcast: "192.168.1.255",
						Port:      DefaultWoLPort,
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
	viper.SetConfigName(DefaultConfigName)
	viper.SetConfigType(DefaultConfigType)
	viper.AddConfigPath(DefaultConfigPath)
	viper.SetEnvPrefix("UPSWAKE")
	viper.OnConfigChange(func(in fsnotify.Event) {
		if _, err := load(fileSystem); err != nil {
			log.Fatal(err)
		}
	})
}

func Load() (*entity.Config, error) {
	return load(fileSystem)
}

func load(fs afero.Fs) (*entity.Config, error) {
	viper.SetFs(fs)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return &entity.Config{}, ErrorConfigNotFound
		}
		// Config file was found but another error was produced
		return &entity.Config{}, err
	}

	viper.AutomaticEnv() // read in environment variables that match

	loadConfig := Config{}
	if err := viper.Unmarshal(&loadConfig); err != nil {
		return &entity.Config{}, err
	}

	config = *fromFileConfig(&loadConfig)
	return &config, nil
}

func CreateDefaultConfig() (*entity.Config, error) {
	return fromFileConfig(&DefaultConfig), nil
}
