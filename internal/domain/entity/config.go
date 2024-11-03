package entity

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"reflect"
	"time"
)

var (
	ErrorUsernameRequired  = errors.New("username is required")
	ErrorPasswordRequired  = errors.New("password is required")
	ErrorInvalidPort       = errors.New("port must be between 1 and 65535")
	ErrorNameRequired      = errors.New("name is required")
	ErrorHostRequired      = errors.New("host is required")
	ErrorMACRequired       = errors.New("MAC is required")
	ErrorInvalidMac        = errors.New("MAC address is invalid")
	ErrorInvalidHost       = errors.New("host is invalid, must be an IP address or hostname")
	ErrorBroadcastRequired = errors.New("broadcast is required")
	ErrorInvalidBroadcast  = errors.New("broadcast is invalid, must be an IP address")
	ErrorIntervalRequired  = errors.New("interval is required")
	ErrorInvalidInterval   = errors.New("interval is invalid, must be a duration")
	validate               *validator.Validate
)

const (
	DefaultWoLPort       = 9
	DefaultNUTServerPort = 3493
)

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("duration", duration, true); err != nil {
		log.Fatalf("could not register Duration validator: %s", err)
	}
}

func duration(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		dur, err := time.ParseDuration(fl.Field().String())
		if err != nil || dur < 1*time.Millisecond {
			return false
		}
		return true
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}
}

type Config struct {
	NutServers []NutServer
}

type NutServer struct {
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	Targets  []TargetServer
}

func (ns *NutServer) validate() error {
	if ns.Name == "" {
		return ErrorNameRequired
	}
	if ns.Host == "" {
		return ErrorHostRequired
	}
	if validate.Var(ns.Host, "ip|hostname") != nil {
		return ErrorInvalidHost
	}
	if ns.Port < 1 || ns.Port > 65535 {
		return ErrorInvalidPort
	}
	if ns.Username == "" {
		return ErrorUsernameRequired
	}
	if ns.Password == "" {
		return ErrorPasswordRequired
	}
	for _, target := range ns.Targets {
		if err := target.validate(); err != nil {
			return err
		}
	}
	return nil
}

type TargetServer struct {
	Name      string
	MAC       string
	Broadcast string
	Port      int    `default:"9"`
	Interval  string `default:"15m"`
	Rules     []string
}

func (ts *TargetServer) validate() error {
	if ts.Name == "" {
		return ErrorNameRequired
	}
	if ts.MAC == "" {
		return ErrorMACRequired
	}
	if validate.Var(ts.MAC, "mac") != nil {
		return ErrorInvalidMac
	}
	if ts.Broadcast == "" {
		return ErrorBroadcastRequired
	}
	if validate.Var(ts.Broadcast, "ip") != nil {
		return ErrorInvalidBroadcast
	}
	if ts.Port < 1 || ts.Port > 65535 {
		return ErrorInvalidPort
	}
	if ts.Interval == "" {
		return ErrorIntervalRequired
	}
	if validate.Var(ts.Interval, "duration") != nil {
		return ErrorInvalidInterval
	}

	return nil
}

func NewTargetServer(name, mac, broadcast, interval string, port int, rules []string) (*TargetServer, error) {
	ts := &TargetServer{
		Name:      name,
		MAC:       mac,
		Broadcast: broadcast,
		Port:      port,
		Interval:  interval,
		Rules:     rules,
	}
	err := ts.validate()
	if err != nil {
		return nil, err
	}
	return ts, nil
}

type ConfigLoader interface {
	Load() (*Config, error)
}
