package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
)

const DefaultNUTPort = 3493

type Host struct {
	Host        string        `yaml:"host" validate:"required,ip|hostname"`
	PortIn      int           `yaml:"port" validate:"omitempty,gte=1,lte=65535"`
	Name        string        `yaml:"name" validate:"required"`
	Credentials []Credentials `yaml:"credentials" validate:"required,dive"`
}

type Credentials struct {
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type WakeHosts struct {
	Name      string   `yaml:"name" validate:"required"`
	Mac       string   `yaml:"mac" validate:"required,mac"`
	Broadcast string   `yaml:"broadcast" validate:"required,ip"`
	Port      int      `yaml:"port" validate:"omitempty,gte=1,lte=65535" default:"9"`
	NutHost   NutHost  `yaml:"nutHost" validate:"required"`
	Rules     []string `yaml:"rules" validate:"required,gt=0,dive,required"`
}

type NutHost struct {
	Name     string `yaml:"name" validate:"required"`
	Username string `yaml:"username" validate:"required"`
}

type Config struct {
	NutHosts  []Host      `yaml:"nutHosts"`
	WakeHosts []WakeHosts `yaml:"wakeHosts"`
}

func (cfg *Config) getHostConfig(name string) (Host, error) {
	for _, host := range cfg.NutHosts {
		if host.Name == name {
			return host, nil
		}
	}
	return Host{}, fmt.Errorf("could not find host '%s' in config", name)
}

// GetHostConfig Get the host config for a given wakehost name
// We're assuming that the config.IsValid has been run before this
func (cfg *Config) GetHostConfig(name string) Host {
	host, err := cfg.getHostConfig(name)
	if err != nil {
		panic(err)
	}
	return host
}

// IsValid Validate the config
// ensure all 'wakeHosts' are valid and have a corresponding 'nutHost' that is valid
// nutHosts that are not used are not used by a wakeHost are not validated
func (cfg *Config) IsValid() error {
	validate := validator.New()

	for _, wakeHost := range cfg.WakeHosts {
		log.Println("Validating config")

		if err := validate.Struct(wakeHost); err != nil {
			return fmt.Errorf("invalid wakeHost: %s", err)
		}

		if err := validate.Struct(wakeHost.NutHost); err != nil {
			return fmt.Errorf("invalid nutHost for %s: %s", wakeHost.Name, err)
		}

		host, err := cfg.getHostConfig(wakeHost.NutHost.Name)
		if err != nil {
			return fmt.Errorf("could not find corresponding NUT host for wakehost %s", wakeHost.Name)
		}

		if err = validate.Struct(host); err != nil {
			return fmt.Errorf("invalid host: %s", err)
		}

		for _, cred := range host.Credentials {
			if err = validate.Struct(cred); err != nil {
				return fmt.Errorf("invalid host credentials: %s", err)
			}
		}

	}
	return nil
}

func (host *Host) GetCredentials(username string) Credentials {
	for _, credentials := range host.Credentials {
		if credentials.Username == username {
			return credentials
		}
	}
	return Credentials{}
}

func (host *Host) Port() int {
	if host.PortIn == 0 {
		return DefaultNUTPort
	}
	return host.PortIn
}

func CreateDefaultConfig() Config {
	return Config{
		NutHosts: []Host{
			{
				Host:   "192.168.1.133",
				PortIn: DefaultNUTPort,
				Name:   "ups1",
				Credentials: []Credentials{
					{
						Username: "upsmon",
						Password: "bigsecret",
					},
				},
			},
		},
		WakeHosts: []WakeHosts{
			{
				Name:      "server1",
				Mac:       "00:00:00:00:00:00",
				Broadcast: "192.168.1.255",
				Port:      9,
				NutHost: NutHost{
					Name:     "ups1",
					Username: "upsmon",
				},
				Rules: []string{
					"80percentOn.rego",
				},
			},
		},
	}
}
