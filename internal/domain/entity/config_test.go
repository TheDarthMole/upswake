package entity

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		NutServers []*NutServer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "empty NutServers",
			fields: fields{
				NutServers: nil,
			},
			wantErr: nil,
		},
		{
			name: "one valid NutServer",
			fields: fields{
				NutServers: []*NutServer{
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
			wantErr: nil,
		},
		{
			name: "one valid one invalid NutServer",
			fields: fields{
				NutServers: []*NutServer{
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
			wantErr: ErrInvalidHost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				NutServers: tt.fields.NutServers,
			}
			err := c.Validate()

			assert.ErrorIs(t, err, tt.wantErr)
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
		name    string
		args    args
		want    *TargetServer
		wantErr error
	}{
		{
			name: "valid NewTargetServer",
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
			wantErr: nil,
		},
		{
			name: "invalid NewTargetServer",
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
			want:    nil,
			wantErr: ErrInvalidBroadcast,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTargetServer(tt.args.name, tt.args.mac, tt.args.broadcast, tt.args.interval, tt.args.port, tt.args.rules)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
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
		Targets  []*TargetServer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid NutServer",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr: nil,
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
			wantErr: ErrInvalidHost,
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
			wantErr: ErrInvalidPort,
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
			wantErr: ErrInvalidPort,
		},
		{
			name: "valid NutServer with single valid target",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     DefaultNUTServerPort,
				Username: "test",
				Password: "test",
				Targets: []*TargetServer{
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
			wantErr: nil,
		},
		{
			name: "valid NutServer with multiple valid targets",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     DefaultNUTServerPort,
				Username: "test",
				Password: "test",
				Targets: []*TargetServer{
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
			wantErr: nil,
		},
		{
			name: "valid NutServer with one valid and one invalid TargetServer",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     DefaultNUTServerPort,
				Username: "test",
				Password: "test",
				Targets: []*TargetServer{
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
						MAC:       "xx:22:33:44:55:yy", // invalid mac address for target server
						Broadcast: "192.168.1.255",
						Port:      DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"test.rego",
						},
					},
				},
			},
			wantErr: ErrInvalidMac, // a target server has invalid characters in MAC
		},
		{
			name: "NutServer no name",
			fields: fields{
				Name:     "",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "NutServer no host",
			fields: fields{
				Name:     "test",
				Host:     "",
				Port:     3493,
				Username: "test",
				Password: "test",
				Targets:  nil,
			},
			wantErr: ErrHostRequired,
		},
		{
			name: "NutServer no username",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "",
				Password: "test",
				Targets:  nil,
			},
			wantErr: ErrUsernameRequired,
		},
		{
			name: "NutServer no password",
			fields: fields{
				Name:     "test",
				Host:     "192.168.1.133",
				Port:     3493,
				Username: "test",
				Password: "",
				Targets:  nil,
			},
			wantErr: ErrPasswordRequired,
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
			assert.ErrorIs(t, err, tt.wantErr)
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
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid TargetServer",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: nil,
		},
		{
			name: "valid TargetServer multiple rules",
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
			wantErr: nil,
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
			wantErr: ErrInvalidMac,
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
			wantErr: ErrInvalidBroadcast,
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
			wantErr: ErrInvalidPort,
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
			wantErr: ErrInvalidPort,
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
			wantErr: ErrInvalidInterval,
		},
		{
			name: "TargetServer no name",
			fields: fields{
				Name:      "",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "TargetServer no mac",
			fields: fields{
				Name:      "test",
				MAC:       "",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: ErrMACRequired,
		},
		{
			name: "TargetServer no broadcast",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: ErrBroadcastRequired,
		},
		{
			name: "TargetServer no interval",
			fields: fields{
				Name:      "test",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
				Port:      9,
				Interval:  "",
				Rules:     []string{},
			},
			wantErr: ErrIntervalRequired,
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
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func Test_duration(t *testing.T) {
	type durationTest struct {
		Duration string `validate:"duration"`
	}

	type nonStringDurationTest struct {
		Duration float64 `validate:"duration"`
	}

	type args struct {
		fl any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "15m",
			args: args{
				fl: durationTest{Duration: "15m"},
			},
			wantErr: false,
		},
		{
			name: "1s",
			args: args{
				fl: durationTest{Duration: "1s"},
			},
			wantErr: false,
		},
		{
			name: "twenty minutes",
			args: args{
				fl: durationTest{Duration: "twenty minutes"},
			},
			wantErr: true,
		},
		{
			name: "non-string duration",
			args: args{
				fl: nonStringDurationTest{Duration: 1},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate = validator.New()
			err := validate.RegisterValidation("duration", duration, true)
			require.NoError(t, err)

			err = validate.Struct(tt.args.fl)

			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
