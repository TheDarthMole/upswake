package viper

import (
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_Load(t *testing.T) {
	type args struct {
		fs       afero.Fs
		filePath string
	}
	testFS := afero.NewBasePathFs(afero.NewOsFs(), "./testing/")

	tests := []struct {
		wantErr error
		want    *entity.Config
		args    args
		name    string
	}{
		{
			name: "valid config",
			args: args{
				fs:       testFS,
				filePath: "valid_config.yaml",
			},
			wantErr: nil,
			want: &entity.Config{
				NutServers: []*entity.NutServer{
					{
						Name:     "nut_server_1",
						Host:     "192.168.1.133",
						Port:     3493,
						Username: "upsmon",
						Password: "password",
						Targets: []*entity.TargetServer{
							{
								Name:      "nas_1",
								MAC:       "00:11:22:33:44:55",
								Broadcast: "192.168.1.255",
								Port:      9,
								Interval:  5 * time.Minute,
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
			wantErr: nil,
			want: &entity.Config{
				NutServers: []*entity.NutServer{
					{
						Name:     "nut_server_1",
						Host:     "192.168.1.133",
						Port:     3493,
						Username: "upsmon",
						Password: "password",
						Targets:  []*entity.TargetServer{},
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
			wantErr: entity.ErrInvalidHost,
			want:    nil,
		},
		{
			name: "port number greater than 65535",
			args: args{
				fs:       testFS,
				filePath: "invalid_port_too_large.yaml",
			},
			wantErr: entity.ErrInvalidPort,
			want:    nil,
		},
		{
			name: "port number less than 1",
			args: args{
				fs:       testFS,
				filePath: "invalid_port_too_small.yaml",
			},
			wantErr: entity.ErrInvalidPort,
			want:    nil,
		},
		{
			name: "invalid target mac",
			args: args{
				fs:       testFS,
				filePath: "invalid_target_mac.yaml",
			},
			wantErr: entity.ErrInvalidMac,
			want:    nil,
		},
		{
			name: "config file does not exist",
			args: args{
				fs:       testFS,
				filePath: "does_not_exist.yaml",
			},
			wantErr: ErrReadingConfigFile,
			want:    nil,
		},
		{
			name: "config file username is array",
			args: args{
				fs:       testFS,
				filePath: "invalid_type.yaml",
			},
			wantErr: ErrUnmarshallingConfig,
			want:    nil,
		},
		{
			name: "invalid interval",
			args: args{
				fs:       testFS,
				filePath: "invalid_interval.yaml",
			},
			wantErr: ErrFailedParsingInterval,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewConfigLoader(tt.args.fs, tt.args.filePath)
			got, err := cl.Load()

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("config doesn't exist", func(t *testing.T) {
		fileSystem := afero.NewMemMapFs()
		cl := NewConfigLoader(fileSystem, "non_existent_config.yaml")
		_, err := cl.Load()
		assert.Error(t, err, "Expected error when config file does not exist")
		assert.ErrorContains(t, err, "file does not exist")
	})

	t.Run("malformed yaml", func(t *testing.T) {
		fileSystem := testFS
		cl := NewConfigLoader(fileSystem, "malformed_file.yaml")
		_, err := cl.Load()
		assert.Error(t, err, "Expected error when config file is malformed")
		assert.ErrorContains(t, err, "expected type 'string'")
	})
}
