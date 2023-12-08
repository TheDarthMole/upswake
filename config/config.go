package config

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/rego"
	"github.com/TheDarthMole/UPSWake/util"
	"github.com/go-playground/validator/v10"
	"log"
	"os"
	"reflect"
	"time"
)

const DefaultNUTPort = 3493

var regoFiles = os.DirFS("rules")

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
	Interval  string    `yaml:"interval" validate:"duration" default:"15m"`
	NutServer NutServer `yaml:"nutServer" validate:"required"`
	Rules     []string  `yaml:"rules" validate:"required,dive,regofile,required"`
}

type Config struct {
	WoLTargets []WoLTarget `yaml:"wolTargets"`
}

func Duration(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		if _, err := time.ParseDuration(fl.Field().String()); err != nil {
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

		err = rego.IsValidRego(string(regoFile))
		if err != nil {
			log.Printf("File %s is not a valid rego file: %s", field.String(), err)
			return false
		}
		return true
	}

	panic(fmt.Sprintf("Bad field type %T", field.Interface()))
}

func (wol *WoLTarget) Validate() error {
	validate := validator.New()
	err := validate.RegisterValidation("duration", Duration, true)
	if err != nil {
		return fmt.Errorf("could not register Duration validator: %s", err)
	}
	err = validate.RegisterValidation("regofile", IsRegoFile, true)
	if err != nil {
		return fmt.Errorf("could not register IsRegoFile validator: %s", err)
	}
	if err = validate.Struct(wol); err != nil {
		return fmt.Errorf("invalid woLTarget: %s", err)
	}
	return nil
}

func (cred *Credentials) Validate() error {
	validate := validator.New()
	if err := validate.Struct(cred); err != nil {
		return fmt.Errorf("invalid credentials: %s", err)
	}
	return nil
}

func (ns *NutServer) Validate() error {
	validate := validator.New()
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
// ensure all 'WoLTargets' are valid and have a corresponding 'NutServers' that is valid
// 'NutServers' that are not used are not used by a 'WoLTargets' are not validated
func (cfg *Config) IsValid() error {
	for _, woLTarget := range cfg.WoLTargets {
		log.Printf("Validating config for %s\n", woLTarget.Name)

		if err := woLTarget.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func CreateDefaultConfig() Config {
	return Config{
		WoLTargets: []WoLTarget{
			{
				Name:      "server1",
				Mac:       "00:00:00:00:00:00",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				NutServer: NutServer{
					Host: "192.168.1.13",
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
