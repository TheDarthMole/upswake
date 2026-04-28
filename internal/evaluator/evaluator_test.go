package evaluator

import (
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/rules"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validNUTOutput   = `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`
	invalidNUTOutput = "invalid!"
)

var (
	defaultConfig  = viper.CreateDefaultConfig()
	regoAlwaysTrue = []byte(`package upswake
default wake := true`)
	regoAlwaysFalse = []byte(`package upswake
default wake := false`)
	regoCheck100Percent = []byte(`package upswake
default wake := false
wake if {
	input[i].Name == "cyberpower900"
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value == 100
}`)
)

func writeMemFile(t *testing.T, fs afero.Fs, fileName string, contents []byte) {
	require.NoError(t, afero.WriteFile(fs, fileName, contents, 0o644))
}

func newRuleRepo(t *testing.T, files map[string][]byte) *rules.PreparedRepository {
	t.Helper()
	fs := afero.NewMemMapFs()
	for name, content := range files {
		writeMemFile(t, fs, name, content)
	}
	repo, err := rules.NewPreparedRepository(fs)
	require.NoError(t, err)
	// Type assert to get the concrete type for tests
	return repo.(*rules.PreparedRepository)
}

func TestNewRegoEvaluator(t *testing.T) {
	type args struct {
		config   *entity.Config
		ruleRepo repository.RuleRepository
		mac      string
	}

	tests := []struct {
		args args
		want *RegoEvaluator
		name string
	}{
		{
			name: "valid config 1",
			args: args{
				config: defaultConfig,
				mac:    "00:00:00:00:00:00",
			},
			want: &RegoEvaluator{
				config: defaultConfig,
				mac:    "00:00:00:00:00:00",
			},
		},
		{
			name: "valid config 2",
			args: args{
				config: defaultConfig,
				mac:    "00:00:00:00:00:55",
			},
			want: &RegoEvaluator{
				config: defaultConfig,
				mac:    "00:00:00:00:00:55",
			},
		},
		//	TODO: Add more test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRegoEvaluator(tt.args.config, tt.args.mac, tt.args.ruleRepo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegoEvaluator_evaluateExpression(t *testing.T) {
	ruleRepo := newRuleRepo(t, map[string][]byte{
		"alwaysTrue.rego":      regoAlwaysTrue,
		"alwaysFalse.rego":     regoAlwaysFalse,
		"check100Percent.rego": regoCheck100Percent,
	})

	type args struct {
		target    *entity.TargetServer
		inputJSON string
	}
	tests := []struct {
		wantErr error
		args    args
		name    string
		want    bool
	}{
		{
			name: "nothing to evaluate",
			args: args{
				target:    nil,
				inputJSON: "",
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "evaluate with always true rule",
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"alwaysTrue.rego",
					},
				},
				inputJSON: validNUTOutput,
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "evaluate with always false rule",
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"alwaysFalse.rego",
					},
				},
				inputJSON: validNUTOutput,
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "file not found",
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"doesnotexist.rego",
					},
				},
				inputJSON: validNUTOutput,
			},
			want:    false,
			wantErr: rules.ErrRuleNotFound,
		},
		{
			name: "ups 100% check positive",
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"check100Percent.rego",
					},
				},
				inputJSON: `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`,
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "ups 100% check negative",
			args: args{
				target: &entity.TargetServer{
					Port: entity.DefaultWoLPort,
					Rules: []string{
						"check100Percent.rego",
					},
				},
				inputJSON: `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":10,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`,
			},
			want:    false,
			wantErr: nil,
		},
		// TODO: Add more rules that tests inputJSON, e.g. faulty fs
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegoEvaluator{
				ruleRepo: ruleRepo,
			}
			got, err := r.evaluateExpression(tt.args.target, tt.args.inputJSON)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegoEvaluator_evaluateExpressions(t *testing.T) {
	ruleRepo := newRuleRepo(t, map[string][]byte{
		"test.rego": regoAlwaysTrue,
	})

	type fields struct {
		config *entity.Config
		mac    string
	}

	type args struct {
		getUPSJSON func(server *entity.NutServer) (string, error)
	}
	tests := []struct {
		want    EvaluationResult
		wantErr error
		args    args
		fields  fields
		name    string
	}{
		{
			name: "valid eval",
			fields: fields{
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "test",
							Host:     "",
							Port:     entity.DefaultNUTServerPort,
							Username: "",
							Password: "",
							Targets: []*entity.TargetServer{
								{
									Name:      "test server",
									MAC:       "00:11:22:33:44:55",
									Broadcast: "192.168.1.255",
									Port:      entity.DefaultWoLPort,
									Interval:  15 * time.Minute,
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					},
				},
				mac: "00:11:22:33:44:55",
			},
			args: args{
				getUPSJSON: func(_ *entity.NutServer) (string, error) { return validNUTOutput, nil },
			},
			want: EvaluationResult{
				Allowed: true,
				Found:   true,
				Target: &entity.TargetServer{
					Name:      "test server",
					MAC:       "00:11:22:33:44:55",
					Broadcast: "192.168.1.255",
					Port:      entity.DefaultWoLPort,
					Interval:  15 * time.Minute,
					Rules: []string{
						"test.rego",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "missing mac in config",
			fields: fields{
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "test",
							Host:     "",
							Port:     entity.DefaultNUTServerPort,
							Username: "",
							Password: "",
							Targets: []*entity.TargetServer{
								{
									Name:      "test server",
									MAC:       "00:00:00:00:00:00",
									Broadcast: "192.168.1.255",
									Port:      entity.DefaultWoLPort,
									Interval:  15 * time.Minute,
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					},
				},
				mac: "00:11:22:33:44:55",
			},
			args: args{
				getUPSJSON: func(_ *entity.NutServer) (string, error) { return validNUTOutput, nil },
			},
			want: EvaluationResult{
				Allowed: false,
				Found:   false,
				Target:  nil,
			},
			wantErr: nil,
		},
		{
			name: "invalid nutserver output",
			fields: fields{
				config: &entity.Config{
					NutServers: []*entity.NutServer{
						{
							Name:     "test",
							Host:     "",
							Port:     entity.DefaultNUTServerPort,
							Username: "",
							Password: "",
							Targets: []*entity.TargetServer{
								{
									Name:      "test server",
									MAC:       "00:11:22:33:44:55",
									Broadcast: "192.168.1.255",
									Port:      entity.DefaultWoLPort,
									Interval:  15 * time.Minute,
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					},
				},
				mac: "00:11:22:33:44:55",
			},
			args: args{
				getUPSJSON: func(_ *entity.NutServer) (string, error) { return invalidNUTOutput, nil },
			},
			want:    EvaluationResult{},
			wantErr: ErrFailedEvaluateExpression,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegoEvaluator{
				config:   tt.fields.config,
				mac:      tt.fields.mac,
				ruleRepo: ruleRepo,
			}
			got, err := r.evaluateExpressions(tt.args.getUPSJSON)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
