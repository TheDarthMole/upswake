package config

import (
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
