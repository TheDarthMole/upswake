package viper

import (
	"testing"

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
		name    string
		args    args
		want    *entity.Config
		wantErr error
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
			want:    &entity.Config{},
		},
		{
			name: "port number greater than 65535",
			args: args{
				fs:       testFS,
				filePath: "invalid_port_too_large.yaml",
			},
			wantErr: entity.ErrInvalidPort,
			want:    &entity.Config{},
		},
		{
			name: "port number less than 1",
			args: args{
				fs:       testFS,
				filePath: "invalid_port_too_small.yaml",
			},
			wantErr: entity.ErrInvalidPort,
			want:    &entity.Config{},
		},
		{
			name: "invalid target mac",
			args: args{
				fs:       testFS,
				filePath: "invalid_target_mac.yaml",
			},
			wantErr: entity.ErrInvalidMac,
			want:    &entity.Config{},
		},
		{
			name: "config file does not exist",
			args: args{
				fs:       testFS,
				filePath: "does_not_exist.yaml",
			},
			wantErr: ErrReadingConfigFile,
			want:    &entity.Config{},
		},
		{
			name: "config file username is array",
			args: args{
				fs:       testFS,
				filePath: "invalid_type.yaml",
			},
			wantErr: ErrUnmarshallingConfig,
			want:    &entity.Config{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitConfig(tt.args.fs, tt.args.filePath)
			got, err := Load()

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("config doesn't exist", func(t *testing.T) {
		fileSystem := afero.NewMemMapFs()
		InitConfig(fileSystem, "non_existent_config.yaml")
		_, err := Load()
		assert.Error(t, err, "Expected error when config file does not exist")
		assert.ErrorContains(t, err, "file does not exist")
	})

	t.Run("malformed yaml", func(t *testing.T) {
		fileSystem := testFS
		InitConfig(fileSystem, "malformed_file.yaml")
		_, err := Load()
		assert.Error(t, err, "Expected error when config file is malformed")
		assert.ErrorContains(t, err, "expected type 'string'")
	})
}

func TestCreateDefaultConfig(t *testing.T) {
	got := CreateDefaultConfig()

	want := &entity.Config{
		NutServers: []*entity.NutServer{
			{
				Name:     "NUT Server 1",
				Host:     "192.168.1.13",
				Port:     entity.DefaultNUTServerPort,
				Username: "",
				Password: "",
				Targets: []*entity.TargetServer{
					{
						Name:      "NAS 1",
						MAC:       "00:00:00:00:00:00",
						Broadcast: "192.168.1.255",
						Port:      entity.DefaultWoLPort,
						Interval:  "15m",
						Rules: []string{
							"80percentOn.rego",
						},
					},
				},
			},
		},
	}
	assert.Equal(t, want, got)
}
