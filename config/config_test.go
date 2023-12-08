package config

import (
	"github.com/go-playground/validator/v10"
	"testing"
)

var (
	validCredentials = Credentials{
		Username: "test",
		Password: "test",
	}
	validNutServer = NutServer{
		Name:        "test",
		Host:        "127.0.0.1",
		Port:        3493,
		Credentials: validCredentials,
	}
	validWoLTarget = WoLTarget{
		Name:      "test",
		Mac:       "12:34:56:78:90:ab",
		Broadcast: "127.0.0.255",
		Port:      9,
		Interval:  "15m",
		NutServer: validNutServer,
		Rules:     []string{},
	}
)

func TestCredentials_Validate(t *testing.T) {
	type fields struct {
		Username string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				Username: "test",
				Password: "test",
			},
			wantErr: false,
		},
		{
			name: "Missing Username",
			fields: fields{
				Username: "",
				Password: "test",
			},
			wantErr: true,
		},
		{
			name: "Missing Password",
			fields: fields{
				Username: "test",
				Password: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := &Credentials{
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			if err := cred.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNutServer_Validate(t *testing.T) {
	type fields struct {
		Name        string
		Host        string
		Port        int
		Credentials Credentials
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				Name:        "test",
				Host:        "127.0.0.1",
				Port:        3493,
				Credentials: validCredentials,
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			fields: fields{
				Name:        "",
				Host:        "127.0.0.1",
				Port:        3493,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Missing Host",
			fields: fields{
				Name:        "test",
				Host:        "",
				Port:        3493,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Invalid Hostname",
			fields: fields{
				Name:        "test",
				Host:        "invalid!host",
				Port:        3493,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Invalid IP",
			fields: fields{
				Name:        "test",
				Host:        "127.0.0.256",
				Port:        3493,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Negative Port Number",
			fields: fields{
				Name:        "test",
				Host:        "127.0.0.1",
				Port:        -1,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Port Number Too Large",
			fields: fields{
				Name:        "test",
				Host:        "127.0.0.1",
				Port:        65536,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Missing Credentials",
			fields: fields{
				Name:        "test",
				Host:        "127.0.0.1",
				Port:        3493,
				Credentials: Credentials{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NutServer{
				Name:        tt.fields.Name,
				Host:        tt.fields.Host,
				Port:        tt.fields.Port,
				Credentials: tt.fields.Credentials,
			}
			if err := ns.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNutServer_GetPort(t *testing.T) {
	type fields struct {
		Port int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "Default Port",
			fields: fields{
				Port: 0,
			},
			want: DefaultNUTPort,
		},
		{
			name: "Custom Port",
			fields: fields{
				Port: 3400,
			},
			want: 3400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NutServer{
				Port: tt.fields.Port,
			}
			if got := ns.GetPort(); got != tt.want {
				t.Errorf("GetPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWoLTarget_Validate(t *testing.T) {
	type fields struct {
		Name      string
		Mac       string
		Broadcast string
		Port      int
		Interval  string
		NutServer NutServer
		Rules     []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			// TODO: Add tests for rules
		},
		{
			name: "Missing Name",
			fields: fields{
				Name:      "",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Missing MAC",
			fields: fields{
				Name:      "test",
				Mac:       "",
				Broadcast: "127.0.0.255",
				Port:      9,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid MAC",
			fields: fields{
				Name:      "test",
				Mac:       "invalid!mac",
				Broadcast: "127.0.0.255",
				Port:      9,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Missing Broadcast",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "",
				Port:      9,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid Broadcast",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "invalid!broadcast",
				Port:      9,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Negative Port Number",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      -1,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Port Number Too Large",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      65536,
				Interval:  "15m",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Missing Interval",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Interval:  "",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid Interval",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Interval:  "invalid!interval",
				NutServer: validNutServer,
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Missing NutServer",
			fields: fields{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Interval:  "15m",
				NutServer: NutServer{},
				Rules:     []string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wol := &WoLTarget{
				Name:      tt.fields.Name,
				Mac:       tt.fields.Mac,
				Broadcast: tt.fields.Broadcast,
				Port:      tt.fields.Port,
				Interval:  tt.fields.Interval,
				NutServer: tt.fields.NutServer,
				Rules:     tt.fields.Rules,
			}
			if err := wol.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_IsValid(t *testing.T) {
	type fields struct {
		WoLTargets []WoLTarget
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				WoLTargets: []WoLTarget{
					validWoLTarget,
				},
			},
			wantErr: false,
		},
		{
			name: "Empty WoLTargets",
			fields: fields{
				WoLTargets: []WoLTarget{},
			},
			wantErr: false,
		},
		{
			name: "Invalid WoLTarget",
			fields: fields{
				WoLTargets: []WoLTarget{
					{
						Name:      "test",
						Mac:       "12:34:56:78:90:ab",
						Broadcast: "127.0.0.255",
						Port:      9,
						Interval:  "15m",
						NutServer: NutServer{},
						Rules:     []string{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid NutServer",
			fields: fields{
				WoLTargets: []WoLTarget{
					{
						Name:      "test",
						Mac:       "12:34:56:78:90:ab",
						Broadcast: "127.0.0.255",
						Port:      9,
						Interval:  "15m",
						NutServer: NutServer{
							Name:        "test",
							Host:        "invalid!hostname",
							Port:        3493,
							Credentials: validCredentials,
						},
						Rules: []string{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid Credentials",
			fields: fields{
				WoLTargets: []WoLTarget{
					{
						Name:      "test",
						Mac:       "12:34:56:78:90:ab",
						Broadcast: "127.0.0.255",
						Port:      9,
						Interval:  "15m",
						NutServer: NutServer{
							Name:        "test",
							Host:        "127.0.0.1",
							Port:        3493,
							Credentials: Credentials{},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				WoLTargets: tt.fields.WoLTargets,
			}
			if err := cfg.IsValid(); (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	type args struct {
		Duration string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid 15m",
			args: args{
				Duration: "15m",
			},
			wantErr: false,
		},
		{
			name: "Valid 1h",
			args: args{
				Duration: "15h",
			},
			wantErr: false,
		},
		{
			name: "Valid 1s",
			args: args{
				Duration: "1s",
			},
			wantErr: false,
		},
		{
			name: "Invalid",
			args: args{
				Duration: "invalid!duration",
			},
			wantErr: true,
		},
		{
			name: "Empty",
			args: args{
				Duration: "",
			},
			wantErr: true,
		},
		{
			name: "Negative",
			args: args{
				Duration: "-15m",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type test struct {
				Duration string `validate:"duration"`
			}
			v := validator.New()
			if err := v.RegisterValidation("duration", Duration); err != nil {
				t.Errorf("Duration() = %v, want error %v", err, tt.wantErr)
			}
			if err := v.Struct(test{tt.args.Duration}); (err != nil) != tt.wantErr {
				t.Errorf("Duration() = %v, want error %v", err, tt.wantErr)
			}
		})
	}
}
