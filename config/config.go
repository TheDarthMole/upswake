package config

import "fmt"
import "github.com/go-playground/validator/v10"

type Host struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Name        string        `yaml:"name"`
	Credentials []Credentials `yaml:"credentials"`
}

type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type WakeHosts struct {
	Name      string   `yaml:"name"`
	Mac       string   `yaml:"mac"`
	Broadcast string   `yaml:"broadcast"`
	Port      int      `yaml:"port"`
	NutHost   NutHost  `yaml:"nutHost"`
	Rules     []string `yaml:"rules"`
}

type NutHost struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
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

func (cfg *Config) IsValid() error {
	for _, wakeHost := range cfg.WakeHosts {
		validate := validator.New()
		err := validate.Struct(wakeHost)
		if err != nil {
			return fmt.Errorf("invalid wakeHost: %s", err)
		}
		// TODO: add more validation here
		_, err = cfg.getHostConfig(wakeHost.NutHost.Name)
		if err != nil {
			return fmt.Errorf("could not find corresponding NUT host for wakehost %s", wakeHost.Name)
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
