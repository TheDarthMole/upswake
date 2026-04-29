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
	ErrReadingConfigFile   = errors.New("error reading config file")
	ErrUnmarshallingConfig = errors.New("error unmarshalling config")
)

type ConfigLoader struct {
	*viper.Viper
}

func NewConfigLoader(fs afero.Fs, cfgPath string) *ConfigLoader {
	configFilePath := DefaultConfigFile
	if cfgPath != "" {
		configFilePath = cfgPath
	}

	viperConfig := viper.New()
	viperConfig.SetFs(fs)
	viperConfig.SetConfigFile(configFilePath)
	viperConfig.AddConfigPath(".")
	viperConfig.SetEnvPrefix("UPSWAKE")
	viperConfig.AutomaticEnv() // read in environment variables that match
	if ok, _ := afero.Exists(fs, configFilePath); ok {
		viperConfig.WatchConfig()
	}
	return &ConfigLoader{
		Viper: viperConfig,
	}
}

func (c *ConfigLoader) Load() (*entity.Config, error) {
	if err := c.ReadInConfig(); err != nil {
		// Return on any read error (including file not found or decode errors)
		return &entity.Config{}, fmt.Errorf("%w: %w", ErrReadingConfigFile, err)
	}

	loadConfig := &Config{}
	if err := c.Unmarshal(loadConfig); err != nil {
		return &entity.Config{}, fmt.Errorf("%w: %w", ErrUnmarshallingConfig, err)
	}
	entityConfig, err := FromFileConfig(loadConfig)
	if err != nil {
		return &entity.Config{}, err
	}

	if err := entityConfig.Validate(); err != nil {
		return &entity.Config{}, err
	}
	return entityConfig, nil
}
