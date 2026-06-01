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
	case reflect.Int64:
		// true if the time is greater than 1ms
		return time.Duration(field.Int()) >= 1*time.Millisecond
	default:
		slog.Warn("could not parse duration",
			slog.String("field_name", fl.FieldName()),
			slog.Any("value", fl.Field()),
			slog.String("kind", field.Kind().String()))
		return false
	}
}

type Config struct {
	Profiler   *Profiler
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

type Profiler struct {
	Enabled bool `json:"enabled" default:"false"`
}

type NutServer struct {
	Name     string          `json:"name"`
	Host     string          `json:"host"`
	Username string          `json:"username"`
	Password string          `json:"password"`
	Targets  []*TargetServer `json:"targets"`
	Port     int             `json:"port"`
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

func NewMacAddress(mac string) (*MacAddress, error) {
	macAddress := &MacAddress{mac}

	if err := macAddress.Validate(); err != nil {
		return nil, err
	}

	return macAddress, nil
}

type MacAddress struct {
	MAC string `json:"mac" example:"00:11:22:33:44:55"`
}

func (m *MacAddress) Validate() error {
	if m.MAC == "" {
		return ErrMACRequired
	}
	if validate.Var(m.MAC, "mac") != nil {
		return ErrInvalidMac
	}
	return nil
}

type TargetServer struct {
	*MacAddress
	Name      string        `json:"name"`
	Broadcast string        `json:"broadcast"`
	Rules     []string      `json:"rules"`
	Interval  time.Duration `json:"interval" default:"900000000000"`
	Port      int           `json:"port" default:"9"`
}

func (ts *TargetServer) Validate() error {
	if ts.Name == "" {
		return ErrNameRequired
	}
	if ts.MacAddress == nil {
		return ErrMACRequired
	}
	if err := ts.MacAddress.Validate(); err != nil {
		return err
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
	if ts.Interval == 0 {
		return ErrIntervalRequired
	}
	if validate.Var(ts.Interval, "duration") != nil {
		return ErrInvalidInterval
	}

	return nil
}

func NewTargetServer(name, mac, broadcast string, interval time.Duration, port int, rules []string) (*TargetServer, error) {
	ts := &TargetServer{
		Name:       name,
		MacAddress: &MacAddress{mac},
		Broadcast:  broadcast,
		Port:       port,
		Interval:   interval,
		Rules:      rules,
	}
	if err := ts.Validate(); err != nil {
		return nil, err
	}
	return ts, nil
}

func CreateDefaultConfig() *Config {
	return &Config{
		Profiler: &Profiler{
			Enabled: false,
		},
		NutServers: []*NutServer{
			{
				Name:     "NUT Server 1",
				Host:     "192.168.1.13",
				Port:     DefaultNUTServerPort,
				Username: "username",
				Password: "password",
				Targets: []*TargetServer{
					{
						Name:       "NAS 1",
						MacAddress: &MacAddress{"00:00:00:00:00:00"},
						Broadcast:  "192.168.1.255",
						Port:       DefaultWoLPort,
						Interval:   15 * time.Minute,
						Rules: []string{
							"80percentOn.rego",
						},
					},
				},
			},
		},
	}
}
