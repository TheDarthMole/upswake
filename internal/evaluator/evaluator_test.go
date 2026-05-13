package evaluator

import (
	"errors"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/rules"
	"github.com/TheDarthMole/UPSWake/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	validNUTOutput = `[{"Name":"cyberpower900","Description":"Unavailable","Master":false,"NumberOfLogins":0,"Clients":[],"Variables":[{"Name":"battery.charge","Value":100,"Type":"INTEGER","Description":"Battery charge (percent of full)","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"},{"Name":"ups.status","Value":"OL","Type":"STRING","Description":"UPS status","Writeable":false,"MaximumLength":0,"OriginalType":"NUMBER"}]}]`
)

type upsRepository struct {
	err   error
	json  string
	times int
}
type ruleRepository struct {
	err     error
	times   int
	allowed bool
}

func TestNewRegoEvaluator(t *testing.T) {
	validMac1, err := entity.NewMacAddress("00:00:00:00:00:00")
	require.NoError(t, err)

	validMac2, err := entity.NewMacAddress("00:00:00:00:00:55")
	require.NoError(t, err)

	type args struct {
		config   *entity.Config
		upsRepo  repository.UPSRepository
		ruleRepo repository.RuleRepository
		mac      *entity.MacAddress
	}

	tests := []struct {
		args args
		want *RegoEvaluator
		name string
	}{
		{
			name: "valid config 1",
			args: args{
				config: entity.CreateDefaultConfig(),
				mac:    validMac1,
			},
			want: &RegoEvaluator{
				config: entity.CreateDefaultConfig(),
				mac:    validMac1,
			},
		},
		{
			name: "valid config 2",
			args: args{
				config: entity.CreateDefaultConfig(),
				mac:    validMac2,
			},
			want: &RegoEvaluator{
				config: entity.CreateDefaultConfig(),
				mac:    validMac2,
			},
		},
		{
			name: "valid config 3",
			args: args{
				config:   entity.CreateDefaultConfig(),
				mac:      validMac1,
				upsRepo:  &mocks.MockUPSRepository{},
				ruleRepo: &mocks.MockRuleRepository{},
			},
			want: &RegoEvaluator{
				config:   entity.CreateDefaultConfig(),
				mac:      validMac1,
				upsRepo:  &mocks.MockUPSRepository{},
				ruleRepo: &mocks.MockRuleRepository{},
			},
		},
		//	TODO: Add more test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRegoEvaluator(tt.args.config, tt.args.mac, tt.args.upsRepo, tt.args.ruleRepo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegoEvaluator_evaluateExpression(t *testing.T) {
	type fields struct {
		ruleRepo ruleRepository
	}
	type args struct {
		target    *entity.TargetServer
		inputJSON string
	}
	tests := []struct {
		wantErr error
		args    args
		name    string
		fields  fields
		want    bool
	}{
		{
			name: "nothing to evaluate",
			args: args{
				target:    nil,
				inputJSON: "",
			},
			fields: fields{
				ruleRepo: ruleRepository{times: 0},
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
			fields: fields{
				ruleRepo: ruleRepository{
					times:   1,
					allowed: true,
				},
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
			fields: fields{
				ruleRepo: ruleRepository{
					times:   1,
					allowed: false,
				},
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
			fields: fields{
				ruleRepo: ruleRepository{
					times:   1,
					allowed: false,
					err:     rules.ErrRuleNotFound,
				},
			},
			want:    false,
			wantErr: rules.ErrRuleNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := gomock.NewController(t)
			ruleRepo := mocks.NewMockRuleRepository(mock)
			ruleRepo.EXPECT().Evaluate(gomock.Any(), gomock.Any()).Return(tt.fields.ruleRepo.allowed, tt.fields.ruleRepo.err).Times(tt.fields.ruleRepo.times)

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
	failingNUTOutputError := errors.New("failed to get NUT output")

	validMac, err := entity.NewMacAddress("00:11:22:33:44:55")
	require.NoError(t, err)

	notFoundMac, err := entity.NewMacAddress("00:00:00:00:00:00")
	require.NoError(t, err)

	type fields struct {
		rulesRepo ruleRepository
		config    *entity.Config
		mac       *entity.MacAddress
		upsRepo   upsRepository
	}

	tests := []struct {
		wantErr error
		want    *EvaluationResult
		name    string
		fields  fields
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
									MAC:       validMac,
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
				upsRepo: upsRepository{
					times: 1,
					json:  validNUTOutput,
				},
				rulesRepo: ruleRepository{
					times:   1,
					allowed: true,
				},
				mac: validMac,
			},
			want: &EvaluationResult{
				Allowed: true,
				Found:   true,
				Target: &entity.TargetServer{
					Name:      "test server",
					MAC:       validMac,
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
			name: "mismatched mac in config",
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
									MAC:       validMac,
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
				upsRepo: upsRepository{
					times: 1,
					json:  validNUTOutput,
				},
				rulesRepo: ruleRepository{times: 0},
				mac:       notFoundMac,
			},
			want: &EvaluationResult{
				Allowed: false,
				Found:   false,
				Target:  nil,
			},
			wantErr: nil,
		},
		{
			name: "no mac in config",
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
									MAC:       nil,
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
				upsRepo: upsRepository{
					times: 1,
					json:  validNUTOutput,
				},
				rulesRepo: ruleRepository{times: 0},
				mac:       validMac,
			},
			want:    nil,
			wantErr: entity.ErrMACRequired,
		},
		{
			name: "failing nutserver output",
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
									MAC:       validMac,
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
				upsRepo: upsRepository{
					times: 1,
					err:   failingNUTOutputError,
					json:  "",
				},
				rulesRepo: ruleRepository{times: 0},
				mac:       validMac,
			},
			want:    nil,
			wantErr: failingNUTOutputError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := gomock.NewController(t)
			upsRepo := mocks.NewMockUPSRepository(mock)
			upsRepo.EXPECT().GetJSON(gomock.Any()).Return(tt.fields.upsRepo.json, tt.fields.upsRepo.err).Times(tt.fields.upsRepo.times)

			ruleRepo := mocks.NewMockRuleRepository(mock)
			ruleRepo.EXPECT().Evaluate(gomock.Any(), gomock.Any()).Return(tt.fields.rulesRepo.allowed, tt.fields.rulesRepo.err).Times(tt.fields.rulesRepo.times)

			r := &RegoEvaluator{
				config:   tt.fields.config,
				mac:      tt.fields.mac,
				upsRepo:  upsRepo,
				ruleRepo: ruleRepo,
			}
			got, err := r.EvaluateExpressions()
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
