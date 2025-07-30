package ups

import (
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/google/uuid"
	nut "github.com/robbiet480/go.nut"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

var (
	randomUsername = uuid.New().String()
	randomPassword = uuid.New().String()
)

func TestConnect(t *testing.T) {
	type args struct {
		host     string
		port     int
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    UPS
		wantErr bool
	}{
		{
			name: "Invalid Server",
			args: args{
				host:     "127.0.0.1",
				port:     12345, // Invalid port
				username: randomUsername,
				password: randomPassword,
			},
			want:    UPS{},
			wantErr: true,
		},
		{
			name: "Invalid IP",
			args: args{
				host:     "755.755.755.755",
				port:     entity.DefaultNUTServerPort,
				username: randomUsername,
				password: randomPassword,
			},
			want:    UPS{},
			wantErr: true,
		},
		{
			name: "Valid Server",
			args: args{
				host:     "127.0.0.1",
				port:     entity.DefaultNUTServerPort,
				username: "upsmon",
				password: "upsmon",
			},
			want: UPS{nut.Client{
				Version:         "Network UPS Tools upsd 2.8.2 - https://www.networkupstools.org/",
				ProtocolVersion: "1.3",
				Hostname:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1).To4(), Port: entity.DefaultNUTServerPort},
			}},
			wantErr: false,
		},
		{
			// This test is bad, as any username/password will work with the default NUT server, however
			// empty username/password is not valid for the NUT server.
			name: "Valid Server, empty credentials",
			args: args{
				host:     "127.0.0.1",
				port:     entity.DefaultNUTServerPort,
				username: "",
				password: "",
			},
			want:    UPS{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := connect(tt.args.host, tt.args.port, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.Version, got.Version)
			assert.Equal(t, tt.want.ProtocolVersion, got.ProtocolVersion)
			assert.Equal(t, tt.want.Hostname, got.Hostname)
		})
	}
}
