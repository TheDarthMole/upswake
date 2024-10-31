package file

import (
	"errors"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
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
	fileSystem    = afero.NewOsFs()
	config        = entity.Config{}
	DefaultConfig = Config{
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
}

func Load() (*entity.Config, error) {
	viper.OnConfigChange(func(in fsnotify.Event) {
		if _, err := load(fileSystem); err != nil {
			log.Fatal(err)
		}
	})
	return load(fileSystem)
}

func load(fs afero.Fs) (*entity.Config, error) {

	viper.SetFs(fs)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			if err1 := createDefaultConfig(); err1 != nil {
				return &entity.Config{}, errors.Join(err, err1)
			}
		}
		// Config file was found but another error was produced
		return &entity.Config{}, err
	}

	loadConfig := Config{}
	unmarshalOptions := viper.DecoderConfigOption(func(decoderConfig *mapstructure.DecoderConfig) {
		// This is needed because the decoder defaults to being 'mapstructure'
		decoderConfig.TagName = DefaultConfigType
	})
	if err := viper.Unmarshal(&loadConfig, unmarshalOptions); err != nil {
		return &entity.Config{}, err
	}

	config = *fromFileConfig(&loadConfig)
	return &config, nil
}

func createDefaultConfig() error {
	viper.SetConfigName(DefaultConfigName)
	viper.SetConfigType(DefaultConfigType)
	viper.AddConfigPath(DefaultConfigPath)

	if err := viper.Unmarshal(DefaultConfig); err != nil {
		return err
	}

	if err := viper.SafeWriteConfigAs(DefaultConfigName + "." + DefaultConfigType); err != nil {
		log.Fatalf("Error creating default config file, %s", err)
	}

	return nil
}
