package viper

import (
	"errors"
	"reflect"
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
		{
			name: "config file does not exist",
			args: args{
				fs:       testFS,
				filePath: "does_not_exist.yaml",
			},
			wantErr:    true,
			wantErrMsg: nil,
			want:       &entity.Config{},
		},
		{
			name: "config file username is array",
			args: args{
				fs:       testFS,
				filePath: "invalid_type.yaml",
			},
			wantErr:    true,
			wantErrMsg: nil, // TODO: tricky error to replicate, just checking that an error occurred
			want:       &entity.Config{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileSystem := tt.args.fs
			InitConfig(fileSystem, tt.args.filePath)
			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantErrMsg == nil {
				// If the error is tool specific, we might not have a predefined error to compare to
				// So we just check that an error occurred and skip the message comparison
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
	got, err := CreateDefaultConfig()
	assert.NoError(t, err)

	want := &entity.Config{
		NutServers: []entity.NutServer{
			{
				Name:     "NUT Server 1",
				Host:     "192.168.1.13",
				Port:     entity.DefaultNUTServerPort,
				Username: "",
				Password: "",
				Targets: []entity.TargetServer{
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

	if !reflect.DeepEqual(got, want) {
		t.Errorf("CreateDefaultConfig() got = %v, want %v", got, want)
	}
}
