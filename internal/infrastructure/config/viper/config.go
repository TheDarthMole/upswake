package viper

import (
	"errors"
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	DefaultConfigFile = "config.yaml"
)

var (
	config                 = &entity.Config{}
	configFilePath         = DefaultConfigFile
	ErrReadingConfigFile   = errors.New("error reading config file")
	ErrUnmarshallingConfig = errors.New("error unmarshaling config")
	DefaultConfig          = Config{
		NutServers: []NutServer{
			{
				Name:     "NUT Server 1",
				Host:     "192.168.1.13",
				Port:     entity.DefaultNUTServerPort,
				Username: "",
				Password: "",
				Targets: []*TargetServer{
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

func InitConfig(fs afero.Fs, cfgPath string) {
	configFilePath = DefaultConfigFile
	if cfgPath != "" {
		configFilePath = cfgPath
	}
	viper.SetFs(fs)
	viper.SetConfigFile(configFilePath)
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("UPSWAKE")
	viper.AutomaticEnv() // read in environment variables that match
	//  viper.OnConfigChange(func(in fsnotify.Event) { # TODO: investigate how to mock this in tests
	//	  fmt.Println("Config file changed:", in.Name)
	//	  err := viper.Unmarshal(&config)
	//	  if err != nil {
	//  		return
	// 	  }
	//  })
	if ok, _ := afero.Exists(fs, configFilePath); ok {
		viper.WatchConfig()
	}
}

func Load() (*entity.Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		// Return on any read error (including file not found or decode errors)
		return &entity.Config{}, fmt.Errorf("%w: %w", ErrReadingConfigFile, err)
	}

	loadConfig := &Config{}
	if err := viper.Unmarshal(loadConfig); err != nil {
		return &entity.Config{}, fmt.Errorf("%w: %w", ErrUnmarshallingConfig, err)
	}
	entityConfig := fromFileConfig(loadConfig)

	if err := entityConfig.Validate(); err != nil {
		return &entity.Config{}, err
	}
	config = entityConfig
	return config, nil
}

func CreateDefaultConfig() *entity.Config {
	return fromFileConfig(&DefaultConfig)
}
