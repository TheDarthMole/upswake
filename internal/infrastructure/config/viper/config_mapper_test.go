package viper

import (
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestFromFileConfig(t *testing.T) {
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
				NutServers: []*entity.NutServer{
					{
						Name:     "TestServer",
						Host:     "localhost",
						Port:     1234,
						Username: "user",
						Password: "pass",
						Targets: []*entity.TargetServer{
							{
								Name: "TestTarget",
								MAC:  "00:11:22:33:44:55",
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
				NutServers: []*entity.NutServer{},
			},
		},
		{
			name: "valid nut server no target servers",
			args: args{
				config: &Config{
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
					NutServers: []*entity.NutServer{
						{
							Name:     "TestServer",
							Host:     "localhost",
							Port:     1234,
							Username: "user",
							Password: "pass",
							Targets: []*entity.TargetServer{
								{
									Name:      "TestTarget",
									MAC:       "00:11:22:33:44:55",
									Rules:     []string{"rule1", "rule2"},
									Interval:  15 * time.Minute,
									Port:      9,
									Broadcast: "127.0.0.255",
								},
							},
						},
					},
				},
			},
			want: &Config{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToFileConfig(tt.args.entityConfig)
			assert.Equal(t, tt.want, got)
		})
	}
}
