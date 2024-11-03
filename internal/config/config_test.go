package config

import (
	"embed"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/go-playground/validator/v10"
	"reflect"

	//"github.com/go-playground/validator/v10"
	"io/fs"
	"testing"
)

const (
	validMAC  = "01:02:03:04:05:06"
	localhost = "127.0.0.1"
)

var (
	validCredentials = NutCredentials{
		Username: "test",
		Password: "test",
	}
	validNutServer    = ValidNutServerChooseTargets([]TargetServer{validTargetServer})
	validTargetServer = TargetServer{
		Name:      "test",
		Mac:       validMAC,
		Broadcast: "127.0.0.255",
		Port:      DefaultWoLPort,
		Interval:  "15m",
		Rules:     []string{},
	}
	validNutServers = []NutServer{
		validNutServer,
	}
	//go:embed "testing/*"
	fakedRegoFiles embed.FS
)

func ValidNutServerChooseTargets(targets []TargetServer) NutServer {
	return NutServer{
		Name:        "test",
		Host:        localhost,
		Port:        DefaultNUTPort,
		Credentials: validCredentials,
		Targets:     targets,
	}
}

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
			cred := &NutCredentials{
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
		Credentials NutCredentials
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
				Host:        localhost,
				Port:        entity.DefaultNUTServerPort,
				Credentials: validCredentials,
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			fields: fields{
				Name:        "",
				Host:        localhost,
				Port:        entity.DefaultNUTServerPort,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Missing Host",
			fields: fields{
				Name:        "test",
				Host:        "",
				Port:        entity.DefaultNUTServerPort,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Invalid Hostname",
			fields: fields{
				Name:        "test",
				Host:        "invalid!host",
				Port:        entity.DefaultNUTServerPort,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Invalid IP",
			fields: fields{
				Name:        "test",
				Host:        "127.0.0.256",
				Port:        entity.DefaultNUTServerPort,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Negative Port Number",
			fields: fields{
				Name:        "test",
				Host:        localhost,
				Port:        -1,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Port Number Too Large",
			fields: fields{
				Name:        "test",
				Host:        localhost,
				Port:        65536,
				Credentials: validCredentials,
			},
			wantErr: true,
		},
		{
			name: "Missing NutCredentials",
			fields: fields{
				Name:        "test",
				Host:        localhost,
				Port:        entity.DefaultNUTServerPort,
				Credentials: NutCredentials{},
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
				Mac:       validMAC,
				Broadcast: "127.0.0.255",
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: false,
			// TODO: Add tests for rules
		},
		{
			name: "Invalid Rule Location",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: localhost,
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{"fileDoesNotExist.rego"},
			},
			wantErr: true,
		},
		{
			name: "Valid With Rule",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: localhost,
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{"80percentOn.rego"},
			},
			wantErr: false,
		},
		{
			name: "Invalid Rego File",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: localhost,
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{"regoWithSyntaxError.rego"},
			},
			wantErr: true,
		},
		{
			name: "Multiple Valid Rules",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: localhost,
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules: []string{
					"80percentOn.rego",
					"alwaysPasses.rego",
				},
			},
			wantErr: false,
		},
		{
			name: "One Valid One Invalid Rules",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: localhost,
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules: []string{
					"alwaysPasses.rego",
					"regoWithSyntaxError.rego",
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Name",
			fields: TargetServer{
				Name:      "",
				Mac:       validMAC,
				Broadcast: "127.0.0.255",
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Missing MAC",
			fields: TargetServer{
				Name:      "test",
				Mac:       "",
				Broadcast: "127.0.0.255",
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid MAC",
			fields: TargetServer{
				Name:      "test",
				Mac:       "invalid!mac",
				Broadcast: "127.0.0.255",
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Missing Broadcast",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: "",
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid Broadcast",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: "invalid!broadcast",
				Port:      DefaultWoLPort,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Negative Port Number",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: "127.0.0.255",
				Port:      -1,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Port Number Too Large",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: "127.0.0.255",
				Port:      65536,
				Interval:  "15m",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Missing Interval",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: "127.0.0.255",
				Port:      DefaultWoLPort,
				Interval:  "",
				Rules:     []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid Interval",
			fields: TargetServer{
				Name:      "test",
				Mac:       validMAC,
				Broadcast: "127.0.0.255",
				Port:      DefaultWoLPort,
				Interval:  "invalid!interval",
				Rules:     []string{},
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
func TestConfig_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid",
			config: Config{
				NutServers: validNutServers,
			},
			wantErr: false,
		},
		{
			name: "Multiple Valid",
			config: Config{
				NutServers: []NutServer{
					validNutServer,
					validNutServer,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid NutServer",
			config: Config{
				NutServers: []NutServer{
					validNutServer,
					{
						Name:        "test",
						Host:        "invalid9!host",
						Port:        DefaultNUTPort,
						Credentials: validCredentials,
						Targets: []TargetServer{
							validTargetServer,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "One Valid one Invalid NutServer",
			config: Config{
				NutServers: []NutServer{
					{
						Name:        "test",
						Host:        "invalid9!host",
						Port:        DefaultNUTPort,
						Credentials: validCredentials,
						Targets: []TargetServer{
							validTargetServer,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid TargetServer",
			config: Config{
				NutServers: []NutServer{
					{
						Name:        "test",
						Host:        localhost,
						Port:        DefaultNUTPort,
						Credentials: validCredentials,
						Targets: []TargetServer{
							{
								Name:      "test",
								Mac:       "invalid!mac",
								Broadcast: "127.0.0.255",
								Port:      DefaultWoLPort,
								Interval:  "15m",
								Rules:     []string{},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "One Valid one Invalid TargetServer",
			config: Config{
				NutServers: []NutServer{
					ValidNutServerChooseTargets([]TargetServer{validTargetServer}),
					ValidNutServerChooseTargets([]TargetServer{
						{
							Name:      "test",
							Mac:       "invalid!mac",
							Broadcast: "127.0.0.255",
							Port:      DefaultWoLPort,
							Interval:  "15m",
							Rules:     []string{},
						},
					}),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
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

func TestCreateDefaultConfig(t *testing.T) {
	t.Run("Validate Default Config", func(t *testing.T) {
		got := CreateDefaultConfig()
		if got.Validate() != nil {
			t.Errorf("CreateDefaultConfig() = %v, want valid config", got)
		}
	})
}

func TestConfig_FindTarget(t *testing.T) {
	type fields struct {
		NutServers []NutServer
	}
	validSecondTargetServer := TargetServer{
		Name:      "test",
		Mac:       "00:00:00:00:00:00",
		Broadcast: "192.168.1.255",
		Port:      entity.DefaultWoLPort,
		Interval:  "15m",
		Rules:     []string{},
	}
	type args struct {
		mac string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantTS  *TargetServer
		wantNS  *NutServer
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				NutServers: validNutServers,
			},
			args: args{
				mac: validTargetServer.Mac,
			},
			wantTS:  &validTargetServer,
			wantNS:  &validNutServer,
			wantErr: false,
		},
		{
			name: "Valid target invalid mac",
			fields: fields{
				NutServers: validNutServers,
			},
			args: args{
				mac: "invalidmac",
			},
			wantTS:  nil,
			wantNS:  nil,
			wantErr: true,
		},
		{
			name: "multiple targets valid mac",
			fields: fields{
				NutServers: []NutServer{
					{
						Name:        "test1",
						Host:        localhost,
						Port:        DefaultNUTPort,
						Credentials: validCredentials,
						Targets:     []TargetServer{validTargetServer},
					},
					{
						Name:        "test2",
						Host:        localhost,
						Port:        DefaultNUTPort,
						Credentials: validCredentials,
						Targets:     []TargetServer{validSecondTargetServer},
					},
				},
			},
			args: args{
				mac: validSecondTargetServer.Mac,
			},
			wantTS: &validSecondTargetServer,
			wantNS: &NutServer{
				Name:        "test2",
				Host:        localhost,
				Port:        DefaultNUTPort,
				Credentials: validCredentials,
				Targets:     []TargetServer{validSecondTargetServer},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				NutServers: tt.fields.NutServers,
			}
			gotTargetServer, gotNutServer, err := c.FindTarget(tt.args.mac)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTargetServer, tt.wantTS) {
				t.Errorf("FindTarget() gotTargetServer = %v, want %v", gotTargetServer, tt.wantTS)
			}
			if !reflect.DeepEqual(gotNutServer, tt.wantNS) {
				t.Errorf("FindTarget() gotNutServer = %v, want %v", gotNutServer, tt.wantNS)
			}
		})
	}
}
