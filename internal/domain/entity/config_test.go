package entity

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		NutServers []NutServer
	}
	tests := []struct {
		name           string
		fields         fields
		wantErr        bool
		wantErrMessage error
	}{
		{
			name: "empty nutservers",
			fields: fields{
				NutServers: nil,
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "one valid nutserver",
			fields: fields{
				NutServers: []NutServer{
					{
						Name:     "test",
						Host:     "192.168.1.133",
						Port:     DefaultNUTServerPort,
						Username: "test",
						Password: "test",
						Targets:  nil,
					},
				},
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "one valid one invalid nutserver",
			fields: fields{
				NutServers: []NutServer{
					{
						Name:     "test1",
						Host:     "192.168.1.133",
						Port:     DefaultNUTServerPort,
						Username: "test",
						Password: "test",
						Targets:  nil,
					},
					{
						Name:     "test2",
						Host:     "192.168.1.555",
						Port:     DefaultNUTServerPort,
						Username: "test",
						Password: "test",
						Targets:  nil,
					},
				},
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidHost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				NutServers: tt.fields.NutServers,
			}
			err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !errors.Is(err, tt.wantErrMessage) {
				t.Errorf("load() error = %v, want error message %v", err, tt.wantErrMessage)
			}
		})
	}
}

func TestNewTargetServer(t *testing.T) {
	type args struct {
		name      string
		mac       string
		broadcast string
		interval  string
		port      int
		rules     []string
	}
	tests := []struct {
		name           string
		args           args
		want           *TargetServer
		wantErr        bool
		wantErrMessage error
	}{
		{
			name: "valid newtargetserver",
			args: args{
				name:      "test",
				mac:       "11:22:33:44:55:66",
				broadcast: "192.168.1.255",
				interval:  "15m",
				port:      DefaultWoLPort,
				rules: []string{
					"test1.rego",
					"test2.rego",
				},
			},
			want: &TargetServer{
				Name:      "test",
				MAC:       "11:22:33:44:55:66",
				Broadcast: "192.168.1.255",
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules: []string{
					"test1.rego",
					"test2.rego",
				},
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "invalid newtargetserver",
			args: args{
				name:      "test",
				mac:       "11:22:33:44:55:66",
				broadcast: "192.168.1.555",
				interval:  "15m",
				port:      DefaultWoLPort,
				rules: []string{
					"test1.rego",
					"test2.rego",
				},
			},
			want:           nil,
			wantErr:        true,
			wantErrMessage: ErrorInvalidBroadcast,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTargetServer(tt.args.name, tt.args.mac, tt.args.broadcast, tt.args.interval, tt.args.port, tt.args.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTargetServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !errors.Is(err, tt.wantErrMessage) {
				t.Errorf("load() error = %v, want error message %v", err, tt.wantErrMessage)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTargetServer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNutServer_Validate(t *testing.T) {
	type fields struct {
		Name     string
		Host     string
		Port     int
		Username string
		Password string
		Targets  []TargetServer
	}
	tests := []struct {
		name           string
		fields         fields
		wantErr        bool
		wantErrMessage error
	}{
		{
			name: "valid nutserver",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "invalid host",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.555",
				Port:     DefaultNUTServerPort,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidHost,
		},
		{
			name: "invalid port too large",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     1234567890,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidPort,
		},
		{
			name: "invalid port too small",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     -1,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidPort,
		},
		{
			name: "valid nutserver with single valid target",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     DefaultNUTServerPort,
				Username: "test",
				Password: "test",
				Targets: []TargetServer{
					{
						Name:      "test1",
						MAC:       "00:11:22:33:44:55",
						Broadcast: "192.168.1.255",
						Port:      DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"test.rego",
						},
					},
				},
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "valid nutserver with multiple valid targets",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     DefaultNUTServerPort,
				Username: "test",
				Password: "test",
				Targets: []TargetServer{
					{
						Name:      "test1",
						MAC:       "00:11:22:33:44:55",
						Broadcast: "192.168.1.255",
						Port:      DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"test.rego",
						},
					},
					{
						Name:      "test2",
						MAC:       "11:22:33:44:55:66",
						Broadcast: "192.168.1.255",
						Port:      DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"test.rego",
						},
					},
				},
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "valid nutserver with one valid and one invalid targetserver",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     DefaultNUTServerPort,
				Username: "test",
				Password: "test",
				Targets: []TargetServer{
					{
						Name:      "test1",
						MAC:       "00:11:22:33:44:55",
						Broadcast: "192.168.1.255",
						Port:      DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"test.rego",
						},
					},
					{
						Name:      "test2",
						MAC:       "xx:22:33:44:55:yy", // invalid mac address for targetserver
						Broadcast: "192.168.1.255",
						Port:      DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"test.rego",
						},
					},
				},
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidMac, // a targetserver has invalid characters in mac
		},
		{
			name: "nutserver no name",
			fields: fields{
				Name:     "",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr:        true,
			wantErrMessage: ErrorNameRequired,
		},
		{
			name: "nutserver no host",
			fields: fields{
				Name:     "test",
				Host:     "",
				Port:     3493,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr:        true,
			wantErrMessage: ErrorHostRequired,
		},
		{
			name: "nutserver no username",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "",
				Password: "test",
				Targets:  nil,
			},
			wantErr:        true,
			wantErrMessage: ErrorUsernameRequired,
		},
		{
			name: "nutserver no password",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "test",
				Password: "",
				Targets:  nil,
			},
			wantErr:        true,
			wantErrMessage: ErrorPasswordRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NutServer{
				Name:     tt.fields.Name,
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
				Targets:  tt.fields.Targets,
			}
			err := ns.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !errors.Is(err, tt.wantErrMessage) {
				t.Errorf("load() error = %v, want error message %v", err, tt.wantErrMessage)
			}
		})
	}
}

func TestTargetServer_Validate(t *testing.T) {
	type fields struct {
		Name      string
		MAC       string
		Broadcast string
		Port      int
		Interval  string
		Rules     []string
	}
	tests := []struct {
		name           string
		fields         fields
		wantErr        bool
		wantErrMessage error
	}{
		{
			name: "valid targetserver",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "valid targetserver multiple rules",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules: []string{
					"test1.rego",
					"test2.rego",
				},
			},
			wantErr:        false,
			wantErrMessage: nil,
		},
		{
			name: "invalid mac",
			fields: fields{
				Name:      "test",
				MAC:       "xx:11:22:33:44:zz",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidMac,
		},
		{
			name: "invalid broadcast",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.555",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidBroadcast,
		},
		{
			name: "invalid port too high",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      1234567890,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidPort,
		},
		{
			name: "invalid port too low",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      -1,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidPort,
		},
		{
			name: "invalid interval",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15beans",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorInvalidInterval,
		},
		{
			name: "targetserver no name",
			fields: fields{
				Name:      "",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorNameRequired,
		},
		{
			name: "targetserver no mac",
			fields: fields{
				Name:      "test",
				MAC:       "",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorMACRequired,
		},
		{
			name: "targetserver no broadcast",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorBroadcastRequired,
		},
		{
			name: "targetserver no interval",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "",
				Rules:     []string{},
			},
			wantErr:        true,
			wantErrMessage: ErrorIntervalRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TargetServer{
				Name:      tt.fields.Name,
				MAC:       tt.fields.MAC,
				Broadcast: tt.fields.Broadcast,
				Port:      tt.fields.Port,
				Interval:  tt.fields.Interval,
				Rules:     tt.fields.Rules,
			}
			err := ts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !errors.Is(err, tt.wantErrMessage) {
				t.Errorf("load() error = %v, want error message %v", err, tt.wantErrMessage)
			}
		})
	}
}

func Test_duration(t *testing.T) {
	type args struct {
		fl validator.FieldLevel
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := duration(tt.args.fl); got != tt.want {
				t.Errorf("duration() = %v, want %v", got, tt.want)
			}
		})
	}
}
