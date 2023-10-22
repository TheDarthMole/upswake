package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
)

const DefaultNUTPort = 3493

type NutServer struct {
	Name        string      `yaml:"name" validate:"required"`
	Host        string      `yaml:"host" validate:"required,ip|hostname"`
	Port        int         `yaml:"port" validate:"omitempty,gte=1,lte=65535"`
	Credentials Credentials `yaml:"credentials" validate:"required"`
}

type Credentials struct {
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type WoLTarget struct {
	Name      string    `yaml:"name" validate:"required"`
	Mac       string    `yaml:"mac" validate:"required,mac"`
	Broadcast string    `yaml:"broadcast" validate:"required,ip"`
	Port      int       `yaml:"port" validate:"omitempty,gte=1,lte=65535" default:"9"`
	NutServer NutServer `yaml:"nutServer" validate:"required"`
	Rules     []string  `yaml:"rules" validate:"required,gt=0,dive,required"`
}

type Config struct {
	WoLTargets []WoLTarget `yaml:"wolTargets"`
}

// IsValid Validate the config
// ensure all 'WoLTargets' are valid and have a corresponding 'NutServers' that is valid
// 'NutServers' that are not used are not used by a 'WoLTargets' are not validated
func (cfg *Config) IsValid() error {
	validate := validator.New()

	for _, woLTarget := range cfg.WoLTargets {
		log.Println("Validating config")

		if err := validate.Struct(woLTarget); err != nil {
			return fmt.Errorf("invalid woLTarget: %s", err)
		}

		if err := validate.Struct(woLTarget.NutServer); err != nil {
			return fmt.Errorf("invalid nutServer for %s: %s", woLTarget.Name, err)
		}

		if err := validate.Struct(woLTarget.NutServer.Credentials); err != nil {
			return fmt.Errorf("invalid nutServer credentials: %s", err)
		}
	}
	return nil
}

func (host *NutServer) GetPort() int {
	if host.Port == 0 {
		return DefaultNUTPort
	}
	return host.Port
}

func CreateDefaultConfig() Config {
	return Config{
		WoLTargets: []WoLTarget{
			{
				Name:      "server1",
				Mac:       "00:00:00:00:00:00",
				Broadcast: "192.168.1.255",
				Port:      9,
				NutServer: NutServer{
					Host: "",
					Port: DefaultNUTPort,
					Name: "ups1",
					Credentials: Credentials{
						Username: "upsmon",
						Password: "bigsecret",
					},
				},
				Rules: []string{
					"80percentOn.rego",
				},
			},
		},
	}
}
