package ups

import (
	"net"
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/google/uuid"
	levenshtein "github.com/ka-weihe/fast-levenshtein"
	nut "github.com/robbiet480/go.nut"
	"github.com/stretchr/testify/assert"
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
		name string
		args args
		want UPS
		err  error
	}{
		{
			name: "Invalid Server",
			args: args{
				host:     "127.0.0.1",
				port:     12345, // Invalid port
				username: randomUsername,
				password: randomPassword,
			},
			want: UPS{},
			err:  ErrConnectionFailed,
		},
		{
			name: "Invalid IP",
			args: args{
				host:     "755.755.755.755",
				port:     entity.DefaultNUTServerPort,
				username: randomUsername,
				password: randomPassword,
			},
			want: UPS{},
			err:  ErrConnectionFailed,
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
				Version:         "Network UPS Tools upsd 2.8.3 - https://www.networkupstools.org/historic/v2.8.3/index.html",
				ProtocolVersion: "1.3",
				Hostname:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1).To4(), Port: entity.DefaultNUTServerPort},
			}},
			err: nil,
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
			want: UPS{},
			err:  ErrFailureAuthenticating,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := connect(tt.args.host, tt.args.port, tt.args.username, tt.args.password)
			assert.ErrorIs(t, err, tt.err)

			assert.Equal(t, tt.want.Version, got.Version)
			assert.Equal(t, tt.want.ProtocolVersion, got.ProtocolVersion)
			assert.Equal(t, tt.want.Hostname, got.Hostname)
		})
	}
}

func TestGetJSON(t *testing.T) {
	type args struct {
		ns *entity.NutServer
	}
	tests := []struct {
		name  string
		args  args
		want  string
		error error
	}{
		{
			name: "Invalid Server",
			args: args{
				ns: &entity.NutServer{
					Host:     "127.0.0.1",
					Port:     12345, // Invalid port
					Username: randomUsername,
					Password: randomPassword,
				},
			},
			want:  "",
			error: ErrConnectionFailed,
		},
		{
			name: "Valid Server",
			args: args{
				ns: &entity.NutServer{
					Host:     "127.0.0.1",
					Port:     entity.DefaultNUTServerPort, // Invalid port
					Username: "upsmon",
					Password: "upsmon",
				},
			},
			want:  `[{"Name":"cyberpower900","Description":"Simulated UPS for testing","Master":false,"NumberOfLogins":1,"Clients":["127.0.0.1"],"Variables":[{"Name":"[default]","Value":3600,"Type":"INTEGER","Description":"Description unavailable","Writeable":true,"MaximumLength":32,"OriginalType":"STRING"},{"Name":"battery.charge","Value":85,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":true,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"battery.runtime","Value":3600,"Type":"INTEGER","Description":"Battery runtime (seconds)","Writeable":true,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"device.mfr","Value":"Dummy Manufacturer","Type":"STRING","Description":"Device manufacturer","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"device.model","Value":"Dummy UPS","Type":"STRING","Description":"Device model","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"device.type","Value":"ups","Type":"STRING","Description":"Device type","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.debug","Value":0,"Type":"INTEGER","Description":"Current debug verbosity level of the driver program","Writeable":true,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.flag.allow_killpower","Value":0,"Type":"INTEGER","Description":"Safety flip-switch to allow the driver daemon to send UPS shutdown command (accessible via driver.killpower)","Writeable":true,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.name","Value":"dummy-ups","Type":"STRING","Description":"Driver name","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.parameter.mode","Value":"dummy","Type":"STRING","Description":"Description unavailable","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.parameter.pollinterval","Value":2,"Type":"INTEGER","Description":"Description unavailable","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.parameter.port","Value":"dummy","Type":"STRING","Description":"Description unavailable","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.parameter.synchronous","Value":"auto","Type":"STRING","Description":"Description unavailable","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.state","Value":"updateinfo","Type":"STRING","Description":"Current state in driver's lifecycle","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"driver.version","Value":"2.8.3","Type":"NUMBER","Description":"Driver version - NUT release","Writeable":false,"MaximumLength":0,"OriginalType":""},{"Name":"driver.version.internal","Value":0.2,"Type":"FLOAT_64","Description":"Internal driver version","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.mfr","Value":"Dummy Manufacturer","Type":"STRING","Description":"UPS manufacturer","Writeable":true,"MaximumLength":32,"OriginalType":"STRING"},{"Name":"ups.model","Value":"Dummy UPS","Type":"STRING","Description":"UPS model","Writeable":true,"MaximumLength":32,"OriginalType":"STRING"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":true,"MaximumLength":32,"OriginalType":"STRING"}],"Commands":[{"Name":"driver.killpower","Description":"Tell the driver daemon to initiate UPS shutdown; should be unlocked with driver.flag.allow_killpower option or variable setting"},{"Name":"driver.reload","Description":"Reload running driver configuration from the file system (only works for changes in some options)"},{"Name":"driver.reload-or-error","Description":"Reload running driver configuration from the file system (only works for changes in some options); return an error if something changed and could not be applied live (so the caller can restart it with new options)"},{"Name":"driver.reload-or-exit","Description":"Reload running driver configuration from the file system (only works for changes in some options); exit the running driver if something changed and could not be applied live (so service management framework can restart it with new options)"},{"Name":"load.off","Description":"Turn off the load immediately"},{"Name":"shutdown.default","Description":"Run the driver-defined UPS shutdown sequence (as opposed to user-configured 'sdcommands')"}]}]`,
			error: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJSON(tt.args.ns)
			assert.ErrorIs(t, err, tt.error)

			if err == nil {
				// usage of Levenshtein distance as NUT server may return slightly different JSON, depending on state of the UPS
				levenshteinDistPercent := (float64(levenshtein.Distance(tt.want, got)) / float64(len(tt.want))) * 100
				assert.LessOrEqualf(t, levenshteinDistPercent, float64(4), "Levenshtein distance between expected and got JSON is too high, indicating a significant difference.\nexpected	(%s), \ngot			(%s).", tt.want, got)
			}
		})
	}
}
