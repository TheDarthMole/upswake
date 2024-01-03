package config

import (
	"embed"
	"github.com/go-playground/validator/v10"

	//"github.com/go-playground/validator/v10"
	"io/fs"
	"testing"
)

var (
	validCredentials = Credentials{
		Username: "test",
		Password: "test",
	}
	//go:embed "testing/*"
	fakedRegoFiles embed.FS
)

func init() {
	regoFiles, _ = fs.Sub(fakedRegoFiles, "testing")
}

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

func TestTargetServer_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fields  TargetServer
		wantErr bool
	}{
		{
			name: "Valid With No Rules",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: false,
			// TODO: Add tests for rules
		},
		{
			name: "Invalid Rule Location",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.1",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{"fileDoesNotExist.rego"},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid With Rule",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.1",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{"80percentOn.rego"},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Rego File",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.1",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{"regoWithSyntaxError.rego"},
				},
			},
			wantErr: true,
		},
		{
			name: "Multiple Valid Rules",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.1",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules: []string{
						"80percentOn.rego",
						"alwaysPasses.rego",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "One Valid One Invalid Rules",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.1",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules: []string{
						"alwaysPasses.rego",
						"regoWithSyntaxError.rego",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Name",
			fields: TargetServer{
				Name:      "",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing MAC",
			fields: TargetServer{
				Name:      "test",
				Mac:       "",
				Broadcast: "127.0.0.255",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid MAC",
			fields: TargetServer{
				Name:      "test",
				Mac:       "invalid!mac",
				Broadcast: "127.0.0.255",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Broadcast",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid Broadcast",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "invalid!broadcast",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Negative Port Number",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      -1,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Port Number Too Large",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      65536,
				Config: TargetServerConfig{
					Interval: "15m",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Interval",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid Interval",
			fields: TargetServer{
				Name:      "test",
				Mac:       "12:34:56:78:90:ab",
				Broadcast: "127.0.0.255",
				Port:      9,
				Config: TargetServerConfig{
					Interval: "invalid!interval",
					Rules:    []string{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fields.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: Fix these tests to fit new structure
//func TestConfig_IsValid(t *testing.T) {
//	type fields struct {
//		WoLTargets []TargetServer
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		{
//			name: "Valid",
//			fields: fields{
//				WoLTargets: []TargetServer{
//					validWoLTarget,
//				},
//			},
//			wantErr: false,
//		},
//		{
//			name: "Empty NutServerMappings",
//			fields: fields{
//				WoLTargets: []TargetServer{},
//			},
//			wantErr: false,
//		},
//		{
//			name: "Invalid TargetServer",
//			fields: fields{
//				WoLTargets: []TargetServer{
//					{
//						Name:      "test",
//						Mac:       "12:34:56:78:90:ab",
//						Broadcast: "127.0.0.255",
//						Port:      9,
//					},
//				},
//			},
//			wantErr: true,
//		},
//		{
//			name: "Invalid NutServer",
//			fields: fields{
//				WoLTargets: []TargetServer{
//					{
//						Name:      "test",
//						Mac:       "12:34:56:78:90:ab",
//						Broadcast: "127.0.0.255",
//						Port:      9,
//					},
//				},
//			},
//			wantErr: true,
//		},
//		{
//			name: "Invalid Credentials",
//			fields: fields{
//				WoLTargets: []TargetServer{
//					{
//						Name:      "test",
//						Mac:       "12:34:56:78:90:ab",
//						Broadcast: "127.0.0.255",
//						Port:      9,
//					},
//				},
//			},
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			cfg := &Config{
//				NutServerMappings: tt.fields.WoLTargets,
//			}
//			if err := cfg.IsValid(); (err != nil) != tt.wantErr {
//				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

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

func TestCreateDefaultConfig(t *testing.T) {
	t.Run("Validate Default Config", func(t *testing.T) {
		got := CreateDefaultConfig()
		if got.IsValid() != nil {
			t.Errorf("CreateDefaultConfig() = %v, want valid config", got)
		}
	})
}
