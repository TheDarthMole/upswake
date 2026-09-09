package viper

import (
	"log/slog"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromFileConfig(t *testing.T) {
	validMac, err := entity.NewMacAddress("00:11:22:33:44:55")
	require.NoError(t, err)

	type args struct {
		config *Config
	}
	tests := []struct {
		err  error
		args args
		want *entity.Config
		name string
	}{
		{
			name: "full file config with one nut server and one target server",
			args: args{
				config: &Config{
					Profiler: &Profiler{
						Enabled: true,
					},
					NutServers: []*NutServer{
						{
							Name:     "TestServer",
							Host:     "localhost",
							Port:     1234,
							Username: "user",
							Password: "pass",
							Targets: []*TargetServer{
								{
									Name: "TestTarget",
									MAC:  "00:11:22:33:44:55",
									Rules: []string{
										"rule1",
										"rule2",
									},
									Interval:  "15m",
									Port:      9,
									Broadcast: "127.0.0.255",
								},
							},
						},
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{
					Enabled: true,
				},
				Logging: &entity.Logging{},
				NutServers: []*entity.NutServer{
					{
						Name:     "TestServer",
						Host:     "localhost",
						Port:     1234,
						Username: "user",
						Password: "pass",
						Targets: []*entity.TargetServer{
							{
								Name:       "TestTarget",
								MacAddress: validMac,
								Rules: []string{
									"rule1",
									"rule2",
								},
								Interval:  15 * time.Minute,
								Port:      9,
								Broadcast: "127.0.0.255",
							},
						},
					},
				},
			},
		},
		{
			name: "empty config",
			args: args{
				config: &Config{},
			},
			want: &entity.Config{
				Profiler:   &entity.Profiler{},
				Logging:    &entity.Logging{},
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "profiler enabled",
			args: args{
				config: &Config{
					Profiler: &Profiler{
						Enabled: true,
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{
					Enabled: true,
				},
				Logging:    &entity.Logging{},
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "profiler disabled",
			args: args{
				config: &Config{
					Profiler: &Profiler{
						Enabled: false,
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{
					Enabled: false,
				},
				Logging:    &entity.Logging{},
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "logging debug",
			args: args{
				config: &Config{
					Logging: &Logging{
						Level: "DEBUG",
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{},
				Logging: &entity.Logging{
					Level: slog.LevelDebug,
				},
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "logging info",
			args: args{
				config: &Config{
					Logging: &Logging{
						Level: "INFO",
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{},
				Logging: &entity.Logging{
					Level: slog.LevelInfo,
				},
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "logging warn",
			args: args{
				config: &Config{
					Logging: &Logging{
						Level: "WARN",
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{},
				Logging: &entity.Logging{
					Level: slog.LevelWarn,
				},
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "logging error",
			args: args{
				config: &Config{
					Logging: &Logging{
						Level: "ERROR",
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{},
				Logging: &entity.Logging{
					Level: slog.LevelError,
				},
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "valid nut server no target servers",
			args: args{
				config: &Config{
					Profiler: &Profiler{},
					Logging: &Logging{
						Level: "INFO",
					},
					NutServers: []*NutServer{
						{
							Name:     "TestServer",
							Host:     "localhost",
							Port:     1234,
							Username: "user",
							Password: "pass",
							Targets:  []*TargetServer{},
						},
					},
				},
			},
			want: &entity.Config{
				Profiler: &entity.Profiler{},
				Logging: &entity.Logging{
					Level: slog.LevelInfo,
				},
				NutServers: []*entity.NutServer{
					{
						Name:     "TestServer",
						Host:     "localhost",
						Port:     1234,
						Username: "user",
						Password: "pass",
						Targets:  []*entity.TargetServer{},
					},
				},
			},
		},
		{
			name: "invalid target server interval",
			args: args{
				config: &Config{
					NutServers: []*NutServer{
						{
							Name:     "TestServer",
							Host:     "localhost",
							Port:     1234,
							Username: "user",
							Password: "pass",
							Targets: []*TargetServer{
								{
									Name:     "TestTarget",
									MAC:      "00:11:22:33:44:55",
									Rules:    []string{"rule1"},
									Interval: "invalid",
									Port:     9,
								},
							},
						},
					},
				},
			},
			want: nil,
			err:  ErrFailedParsingInterval,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromFileConfig(tt.args.config)
			assert.ErrorIs(t, err, tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToFileConfig(t *testing.T) {
	type args struct {
		entityConfig *entity.Config
	}
	tests := []struct {
		args args
		want *Config
		name string
	}{
		{
			name: "full entity config with one nut server and one target server",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{},
					Logging: &entity.Logging{
						Level: slog.LevelInfo,
					},
					NutServers: []*entity.NutServer{
						{
							Name:     "TestServer",
							Host:     "localhost",
							Port:     1234,
							Username: "user",
							Password: "pass",
							Targets: []*entity.TargetServer{
								{
									Name:       "TestTarget",
									MacAddress: &entity.MacAddress{MAC: "00:11:22:33:44:55"},
									Rules:      []string{"rule1", "rule2"},
									Interval:   15 * time.Minute,
									Port:       9,
									Broadcast:  "127.0.0.255",
								},
							},
						},
					},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "INFO",
				},
				NutServers: []*NutServer{
					{
						Name:     "TestServer",
						Host:     "localhost",
						Port:     1234,
						Username: "user",
						Password: "pass",
						Targets: []*TargetServer{
							{
								Name:      "TestTarget",
								MAC:       "00:11:22:33:44:55",
								Rules:     []string{"rule1", "rule2"},
								Interval:  "15m0s", // Trailing zero values are included in the string representation of durations. Annoying I know, but this is how time.Duration.String() works in Go.
								Port:      9,
								Broadcast: "127.0.0.255",
							},
						},
					},
				},
			},
		},
		{
			name: "nil profiler",
			args: args{
				entityConfig: &entity.Config{
					Profiler: nil,
					Logging: &entity.Logging{
						Level: slog.LevelInfo,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "INFO",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "empty profiler",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{},
					Logging: &entity.Logging{
						Level: slog.LevelInfo,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "INFO",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "profiler enabled",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{
						Enabled: true,
					},
					Logging: &entity.Logging{
						Level: slog.LevelInfo,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{
					Enabled: true,
				},
				Logging: &Logging{
					Level: "INFO",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "nil logging",
			args: args{
				entityConfig: &entity.Config{
					Profiler:   &entity.Profiler{},
					Logging:    nil,
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "INFO",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "empty logging",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{},
					Logging: &entity.Logging{
						Level: slog.LevelInfo,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "INFO",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "logging level info",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{},
					Logging: &entity.Logging{
						Level: slog.LevelInfo,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "INFO",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "logging level debug",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{},
					Logging: &entity.Logging{
						Level: slog.LevelDebug,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "DEBUG",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "logging level warning",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{},
					Logging: &entity.Logging{
						Level: slog.LevelWarn,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "WARN",
				},
				NutServers: []*NutServer{},
			},
		},
		{
			name: "logging level error",
			args: args{
				entityConfig: &entity.Config{
					Profiler: &entity.Profiler{},
					Logging: &entity.Logging{
						Level: slog.LevelError,
					},
					NutServers: []*entity.NutServer{},
				},
			},
			want: &Config{
				Profiler: &Profiler{},
				Logging: &Logging{
					Level: "ERROR",
				},
				NutServers: []*NutServer{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToFileConfig(tt.args.entityConfig)
			assert.Equal(t, tt.want, got)
		})
	}
}
