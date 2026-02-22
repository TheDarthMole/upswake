package entity

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	ErrUsernameRequired  = errors.New("username is required")
	ErrPasswordRequired  = errors.New("password is required")
	ErrInvalidPort       = errors.New("port must be a number between 1 and 65535")
	ErrNameRequired      = errors.New("name is required")
	ErrHostRequired      = errors.New("host is required")
	ErrMACRequired       = errors.New("MAC is required")
	ErrInvalidMac        = errors.New("MAC address is invalid")
	ErrInvalidHost       = errors.New("host is invalid, must be an IP address or hostname")
	ErrBroadcastRequired = errors.New("broadcast is required")
	ErrInvalidBroadcast  = errors.New("broadcast is invalid, must be an IP address")
	ErrIntervalRequired  = errors.New("interval is required")
	ErrInvalidInterval   = errors.New("interval is invalid, must be a duration")
	ErrIntervalTooShort  = errors.New("interval must be greater than 0")
	validate             *validator.Validate
)

const (
	DefaultWoLPort       = 9
	DefaultNUTServerPort = 3493
)

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("duration", duration, true); err != nil {
		slog.Error("could not register Duration validator", slog.Any("error", err))
		os.Exit(1)
	}
}

func duration(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		dur, err := time.ParseDuration(fl.Field().String())
		// true if there is no error and the time is greater than 1ms, else false
		return err == nil && dur >= 1*time.Millisecond
	default:
		slog.Warn("could not parse duration",
			slog.String("field_name", fl.FieldName()),
			slog.Any("value", fl.Field()),
			slog.String("kind", field.Kind().String()))
		return false
	}
}

type Config struct {
	NutServers []*NutServer
}

func (c *Config) Validate() error {
	for _, target := range c.NutServers {
		if err := target.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type NutServer struct {
	Name     string          `json:"name"`
	Host     string          `json:"host"`
	Port     int             `json:"port"`
	Username string          `json:"username"`
	Password string          `json:"password"` //nolint:gosec // G117: TODO investigate secure ways to handle this password, such as using environment variables
	Targets  []*TargetServer `json:"targets"`
}

func (ns *NutServer) Validate() error {
	if ns.Name == "" {
		return ErrNameRequired
	}
	if ns.Host == "" {
		return ErrHostRequired
	}
	if validate.Var(ns.Host, "ip|hostname") != nil {
		return ErrInvalidHost
	}
	if ns.Port < 1 || ns.Port > 65535 {
		return ErrInvalidPort
	}
	if ns.Username == "" {
		return ErrUsernameRequired
	}
	if ns.Password == "" {
		return ErrPasswordRequired
	}
	for _, target := range ns.Targets {
		if err := target.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type TargetServer struct {
	Name      string   `json:"name"`
	MAC       string   `json:"mac"`
	Broadcast string   `json:"broadcast"`
	Port      int      `json:"port" default:"9"`
	Interval  string   `json:"interval" default:"15m"`
	Rules     []string `json:"rules"`
}

func (ts *TargetServer) Validate() error {
	if ts.Name == "" {
		return ErrNameRequired
	}
	if ts.MAC == "" {
		return ErrMACRequired
	}
	if validate.Var(ts.MAC, "mac") != nil {
		return ErrInvalidMac
	}
	if ts.Broadcast == "" {
		return ErrBroadcastRequired
	}
	if validate.Var(ts.Broadcast, "ip") != nil {
		return ErrInvalidBroadcast
	}
	if ts.Port < 1 || ts.Port > 65535 {
		return ErrInvalidPort
	}
	if ts.Interval == "" {
		return ErrIntervalRequired
	}
	if validate.Var(ts.Interval, "duration") != nil {
		return ErrInvalidInterval
	}
	interval, err := time.ParseDuration(ts.Interval)
	if err != nil {
		return ErrInvalidInterval
	}
	if interval <= 0 {
		return ErrIntervalTooShort
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
	if err := ts.Validate(); err != nil {
		return nil, err
	}
	return ts, nil
}

type ConfigLoader interface {
	Load() (*Config, error)
}

type NutServerInterface interface {
	Validate() error
	GetJSON() (string, error)
	NutServer
}
