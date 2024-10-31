package ups

import (
	"github.com/google/uuid"
	"reflect"
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
				port:     3493,
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
				port:     3493,
				username: randomUsername,
				password: randomPassword,
			},
			want:    UPS{},
			wantErr: true,
		},
		// TODO: Add tests that connect to a real NUT server and test authentication
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := connect(tt.args.host, tt.args.port, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connect() got = %v, want %v", got, tt.want)
			}
		})
	}
}
