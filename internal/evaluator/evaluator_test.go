package evaluator

import (
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
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
	tempFS         = afero.NewMemMapFs()
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

func TestNewRegoEvaluator(t *testing.T) {
	alwaysTrueRegoFS := afero.NewMemMapFs()
	writeMemFile(t, alwaysTrueRegoFS, "test.rego", regoAlwaysTrue)

	alwaysFalseRegoFS := afero.NewMemMapFs()
	writeMemFile(t, alwaysFalseRegoFS, "test.rego", regoAlwaysFalse)

	type args struct {
		config  *entity.Config
		mac     string
		rulesFS afero.Fs
	}

	tests := []struct {
		name string
		args args
		want *RegoEvaluator
	}{
		{
			name: "valid config 1",
			args: args{
				config:  defaultConfig,
				mac:     "00:00:00:00:00:00",
				rulesFS: alwaysTrueRegoFS,
			},
			want: &RegoEvaluator{
				config:  defaultConfig,
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:00:00:00:00:00",
			},
		},
		{
			name: "valid config 2",
			args: args{
				config:  defaultConfig,
				mac:     "00:00:00:00:00:55",
				rulesFS: alwaysFalseRegoFS,
			},
			want: &RegoEvaluator{
				config:  defaultConfig,
				rulesFS: alwaysFalseRegoFS,
				mac:     "00:00:00:00:00:55",
			},
		},
		//	TODO: Add more test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRegoEvaluator(tt.args.config, tt.args.mac, tt.args.rulesFS)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegoEvaluator_evaluateExpression(t *testing.T) {
	testFS := afero.NewMemMapFs()

	writeMemFile(t, testFS, "alwaysTrue.rego", regoAlwaysTrue)
	writeMemFile(t, testFS, "alwaysFalse.rego", regoAlwaysFalse)
	writeMemFile(t, testFS, "check100Percent.rego", regoCheck100Percent)

	type fields struct {
		rulesFS afero.Fs
	}
	type args struct {
		target    *entity.TargetServer
		inputJSON string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "nothing to evaluate",
			fields: fields{
				rulesFS: tempFS,
			},
			args: args{
				target:    nil,
				inputJSON: "",
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "evaluate with always true rule",
			fields: fields{
				rulesFS: testFS,
			},
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
			fields: fields{
				rulesFS: testFS,
			},
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
			fields: fields{
				rulesFS: testFS,
			},
			args: args{
				target: &entity.TargetServer{
					Rules: []string{
						"doesnotexist.rego",
					},
				},
				inputJSON: validNUTOutput,
			},
			want:    false,
			wantErr: ErrFailedReadRegoFile,
		},
		{
			name: "ups 100% check positive",
			fields: fields{
				rulesFS: testFS,
			},
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
			fields: fields{
				rulesFS: testFS,
			},
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
				rulesFS: tt.fields.rulesFS,
			}
			got, err := r.evaluateExpression(tt.args.target, tt.args.inputJSON)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func writeMemFile(t *testing.T, fs afero.Fs, fileName string, contents []byte) {
	require.NoError(t, afero.WriteFile(fs, fileName, contents, 0o644))
}

func TestRegoEvaluator_evaluateExpressions(t *testing.T) {
	alwaysTrueRegoFS := afero.NewMemMapFs()
	writeMemFile(t, alwaysTrueRegoFS, "test.rego", regoAlwaysTrue)

	type fields struct {
		config  *entity.Config
		rulesFS afero.Fs
		mac     string
	}

	type args struct {
		getUPSJSON func(server *entity.NutServer) (string, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    EvaluationResult
		wantErr error
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
									Interval:  "15m",
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					},
				},
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:11:22:33:44:55",
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
					Interval:  "15m",
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
									Interval:  "15m",
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					},
				},
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:11:22:33:44:55",
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
									Interval:  "15m",
									Rules: []string{
										"test.rego",
									},
								},
							},
						},
					},
				},
				rulesFS: alwaysTrueRegoFS,
				mac:     "00:11:22:33:44:55",
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
				config:  tt.fields.config,
				rulesFS: tt.fields.rulesFS,
				mac:     tt.fields.mac,
			}
			got, err := r.evaluateExpressions(tt.args.getUPSJSON)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
