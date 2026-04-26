package viper

import (
	"errors"
	"fmt"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	DefaultConfigFile = "config.yaml"
)

var (
	configFilePath           = DefaultConfigFile
	ErrReadingConfigFile     = errors.New("error reading config file")
	ErrUnmarshallingConfig   = errors.New("error unmarshalling config")
	ErrFailedReadingRegoFile = errors.New("failed to read rego rule file")
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

func Load(configFS, rulesFS afero.Fs) (*entity.Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		// Return on any read error (including file not found or decode errors)
		return &entity.Config{}, fmt.Errorf("%w: %w", ErrReadingConfigFile, err)
	}

	loadConfig := &Config{}
	if err := viper.Unmarshal(loadConfig); err != nil {
		return &entity.Config{}, fmt.Errorf("%w: %w", ErrUnmarshallingConfig, err)
	}
	config, err := FromFileConfig(loadConfig)
	if err != nil {
		return &entity.Config{}, err
	}

	// TODO: This Load function should be refactored to take in inputs of a config loader.
	// This would allow us to separate the concerns of loading the config and reading the rego files, and would make it easier to test.
	// As well as move towards repository pattern

	for _, nutServer := range config.NutServers {
		for _, target := range nutServer.Targets {
			rulesContent := make([]string, len(target.Rules))
			for index, rule := range target.Rules {
				ruleContent, err := afero.ReadFile(rulesFS, rule)
				if err != nil {
					return &entity.Config{}, fmt.Errorf("%w: '%s': %w", ErrFailedReadingRegoFile, rule, err)
				}
				rulesContent[index] = string(ruleContent)
			}
			target.RulesContent = rulesContent
		}
	}

	if err := config.Validate(); err != nil {
		return &entity.Config{}, err
	}

	return config, nil
}

func CreateDefaultConfig() *entity.Config {
	return &entity.Config{
		NutServers: []*entity.NutServer{
			{
				Name:     "NUT Server 1",
				Host:     "192.168.1.13",
				Port:     entity.DefaultNUTServerPort,
				Username: "",
				Password: "",
				Targets: []*entity.TargetServer{
					{
						Name:      "NAS 1",
						MAC:       "00:00:00:00:00:00",
						Broadcast: "192.168.1.255",
						Port:      entity.DefaultWoLPort,
						Interval:  15 * time.Minute,
						Rules: []string{
							"80percentOn.rego",
						},
					},
				},
			},
		},
	}
}
