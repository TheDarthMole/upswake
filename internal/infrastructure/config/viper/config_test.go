package viper

import (
	"errors"
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/spf13/afero"
	"reflect"
	"testing"
)

func TestCreateDefaultConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *entity.Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateDefaultConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDefaultConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateDefaultConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_load(t *testing.T) {
	type args struct {
		fs       afero.Fs
		filePath string
	}
	testFS := afero.NewBasePathFs(afero.NewOsFs(), "./testing/")

	tests := []struct {
		name       string
		args       args
		want       *entity.Config
		wantErr    bool
		wantErrMsg error
	}{
		{
			name: "valid config",
			args: args{
				fs:       testFS,
				filePath: "valid_config.yaml",
			},
			wantErr:    false,
			wantErrMsg: nil,
			want: &entity.Config{
				NutServers: []entity.NutServer{
					{
						Name:     "nut_server_1",
						Host:     "192.168.1.133",
						Port:     3493,
						Username: "upsmon",
						Password: "password",
						Targets: []entity.TargetServer{
							{
								Name:      "nas_1",
								MAC:       "00:11:22:33:44:55",
								Broadcast: "192.168.1.255",
								Port:      9,
								Interval:  "5m",
								Rules: []string{
									"80percentOn.rego",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "no target servers",
			args: args{
				fs:       testFS,
				filePath: "no_targets_config.yaml",
			},
			wantErr:    false,
			wantErrMsg: nil,
			want: &entity.Config{
				NutServers: []entity.NutServer{
					{
						Name:     "nut_server_1",
						Host:     "192.168.1.133",
						Port:     3493,
						Username: "upsmon",
						Password: "password",
						Targets:  []entity.TargetServer{},
					},
				},
			},
		},
		{
			name: "invalid host",
			args: args{
				fs:       testFS,
				filePath: "invalid_hostname.yaml",
			},
			wantErr:    true,
			wantErrMsg: entity.ErrorInvalidHost,
			want:       &entity.Config{},
		},
		{
			name: "port number greater than 65535",
			args: args{
				fs:       testFS,
				filePath: "invalid_port_too_large.yaml",
			},
			wantErr:    true,
			wantErrMsg: entity.ErrorInvalidPort,
			want:       &entity.Config{},
		},
		{
			name: "port number less than 1",
			args: args{
				fs:       testFS,
				filePath: "invalid_port_too_small.yaml",
			},
			wantErr:    true,
			wantErrMsg: entity.ErrorInvalidPort,
			want:       &entity.Config{},
		},
		{
			name: "invalid target mac",
			args: args{
				fs:       testFS,
				filePath: "invalid_target_mac.yaml",
			},
			wantErr:    true,
			wantErrMsg: entity.ErrorInvalidMac,
			want:       &entity.Config{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := load(tt.args.fs, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !errors.Is(err, tt.wantErrMsg) {
				t.Errorf("load() error = %v, want error message %v", err, tt.wantErrMsg)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("load() got = %v, want %v", got, tt.want)
			}
		})
	}
}
