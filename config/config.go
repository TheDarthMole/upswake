package config

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/rego"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/go-playground/validator/v10"
	"github.com/hack-pad/hackpadfs"
	"log"
	"reflect"
	"time"
)

const (
	DefaultNUTPort    = 3493
	DefaultConfigFile = "config.yaml"
)

var (
	regoFiles hackpadfs.FS
	validate  *validator.Validate
)

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

type TargetServerConfig struct {
	Interval string   `yaml:"interval" validate:"required,duration" default:"15m"`
	Rules    []string `yaml:"rules" validate:"omitempty,dive,regofile"`
}

type TargetServer struct {
	Name      string             `yaml:"name" validate:"required"`
	Mac       string             `yaml:"mac" validate:"required,mac"`
	Broadcast string             `yaml:"broadcast" validate:"required,ip"`
	Port      int                `yaml:"port" validate:"omitempty,gte=1,lte=65535" default:"9"`
	Config    TargetServerConfig `yaml:"config" validate:"omitempty"`
}

type NutServerMapping struct {
	NutServer NutServer      `yaml:"nutServer"`
	Targets   []TargetServer `yaml:"targets"`
}

type Config struct {
	NutServerMappings []NutServerMapping `yaml:"wolTargets"`
}

func init() {
	localFS, err := util.GetLocalFS()
	if err != nil {
		log.Fatalf("could not get local filesystem: %s", err)
	}
	rules, err := hackpadfs.Sub(localFS, "rules")
	if err != nil {
		log.Fatalf("could not get subdirectory 'rules': %s", err)
	}
	regoFiles = rules

	validate = validator.New()
	if err = validate.RegisterValidation("duration", Duration, true); err != nil {
		log.Fatalf("could not register Duration validator: %s", err)
	}

	if err := validate.RegisterValidation("regofile", IsRegoFile, true); err != nil {
		log.Fatalf("could not register IsRegoFile validator: %s", err)
	}

}

func Duration(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		dur, err := time.ParseDuration(fl.Field().String())
		if err != nil || dur < 1*time.Millisecond {
			return false
		}
		return true
	}

	panic(fmt.Sprintf("Bad field type %T", field.Interface()))
}

func IsRegoFile(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		exists := util.FileExists(regoFiles, field.String())
		if !exists {
			log.Printf("File %s does not exist", field.String())
			return false
		}

		regoFile, err := util.GetFile(regoFiles, field.String())
		if err != nil {
			log.Printf("Could not get file: %s", err)
			return false
		}

		if err = rego.IsValidRego(string(regoFile)); err != nil {
			log.Printf("File %s is not a valid rego file: %s", field.String(), err)
			return false
		}
		return true
	}

	panic(fmt.Sprintf("Bad field type %T", field.Interface()))
}

func (nsm *NutServerMapping) Validate() error {
	validate := validator.New()
	if err := validate.Struct(nsm); err != nil {
		return fmt.Errorf("invalid nutServerMapping: %s", err)
	}
	return nil
}

func (wol *TargetServer) Validate() error {
	if err := validate.Struct(wol); err != nil {
		return fmt.Errorf("invalid woLTarget: %s", err)
	}

	return nil
}

func (cfg *TargetServerConfig) Validate() error {
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("invalid TargetServerConfig: %s", err)
	}
	return nil
}

func (cred *Credentials) Validate() error {
	if err := validate.Struct(cred); err != nil {
		return fmt.Errorf("invalid credentials: %s", err)
	}
	return nil
}

func (ns *NutServer) Validate() error {
	if err := validate.Struct(ns); err != nil {
		return fmt.Errorf("invalid nutServer: %s", err)
	}
	return nil
}

func (ns *NutServer) GetPort() int {
	if ns.Port == 0 {
		return DefaultNUTPort
	}
	return ns.Port
}

// IsValid Validate the config
// ensure all 'NutServerMappings' are valid and have a corresponding 'NutServers' that is valid
// 'NutServers' that are not used are not used by a 'NutServerMappings' are not validated
func (cfg *Config) IsValid() error {
	for _, nutServerMapping := range cfg.NutServerMappings {
		log.Printf("Validating config for %s\n", nutServerMapping.NutServer.Name)

		if err := nutServerMapping.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func CreateDefaultConfig() Config {
	return Config{
		NutServerMappings: []NutServerMapping{
			{
				NutServer: NutServer{
					Name: "ups1",
					Host: "192.168.1.13",
					Port: DefaultNUTPort,
					Credentials: Credentials{
						Username: "upsmon",
						Password: "bigsecret",
					},
				},
				Targets: []TargetServer{
					{
						Name:      "server1",
						Mac:       "00:00:00:00:00:00",
						Broadcast: "192.168.1.255",
						Port:      DefaultNUTPort,
						Config: TargetServerConfig{
							Interval: "15m",
							Rules: []string{
								"80percentOn.rego",
							},
						},
					},
				},
			},
		},
	}
}
