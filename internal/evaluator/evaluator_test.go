package evaluator

import (
	"errors"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository/mocks"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/rules"
	"github.com/stretchr/testify/assert"
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
	err   error
	times int
}

func TestNewRegoEvaluator(t *testing.T) {
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
				mac:    &entity.MacAddress{MAC: "00:00:00:00:00:00"},
			},
			want: &RegoEvaluator{
				config: entity.CreateDefaultConfig(),
				mac:    &entity.MacAddress{MAC: "00:00:00:00:00:00"},
			},
		},
		{
			name: "valid config 2",
			args: args{
				config: entity.CreateDefaultConfig(),
				mac:    &entity.MacAddress{MAC: "22:00:00:00:00:00"},
			},
			want: &RegoEvaluator{
				config: entity.CreateDefaultConfig(),
				mac:    &entity.MacAddress{MAC: "22:00:00:00:00:00"},
			},
		},
		{
			name: "valid config 3",
			args: args{
				config:   entity.CreateDefaultConfig(),
				mac:      &entity.MacAddress{MAC: "00:00:00:00:00:00"},
				upsRepo:  &mocks.MockUPSRepository{},
				ruleRepo: &mocks.MockRuleRepository{},
			},
			want: &RegoEvaluator{
				config:   entity.CreateDefaultConfig(),
				mac:      &entity.MacAddress{MAC: "00:00:00:00:00:00"},
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
	}{
		{
			name: "nil target server",
			args: args{
				target:    nil,
				inputJSON: "",
			},
			fields: fields{
				ruleRepo: ruleRepository{times: 0},
			},
			wantErr: ErrFailedEvaluateExpression, // because the target is nil
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
					times: 1,
					err:   nil,
				},
			},
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
					times: 1,
					err:   entity.ErrEvaluationFalse,
				},
			},
			wantErr: entity.ErrEvaluationFalse,
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
					times: 1,
					err:   rules.ErrRuleNotFound,
				},
			},
			wantErr: rules.ErrRuleNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := gomock.NewController(t)
			ruleRepo := mocks.NewMockRuleRepository(mock)
			ruleRepo.EXPECT().Evaluate(gomock.Any(), gomock.Any()).Return(tt.fields.ruleRepo.err).Times(tt.fields.ruleRepo.times)

			r := &RegoEvaluator{
				ruleRepo: ruleRepo,
			}
			err := r.evaluateExpression(tt.args.target, tt.args.inputJSON)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRegoEvaluator_evaluateExpressions(t *testing.T) {
	failingNUTOutputError := errors.New("failed to get NUT output")

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
									Name:       "test server",
									MacAddress: &entity.MacAddress{MAC: "00:11:22:33:44:55"},
									Broadcast:  "192.168.1.255",
									Port:       entity.DefaultWoLPort,
									Interval:   15 * time.Minute,
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
					times: 1,
					err:   nil,
				},
				mac: &entity.MacAddress{MAC: "00:11:22:33:44:55"},
			},
			want: &EvaluationResult{
				Allowed: true,
				Found:   true,
				Target: &entity.TargetServer{
					Name:       "test server",
					MacAddress: &entity.MacAddress{MAC: "00:11:22:33:44:55"},
					Broadcast:  "192.168.1.255",
					Port:       entity.DefaultWoLPort,
					Interval:   15 * time.Minute,
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
									Name:       "test server",
									MacAddress: &entity.MacAddress{MAC: "00:11:22:33:44:55"},
									Broadcast:  "192.168.1.255",
									Port:       entity.DefaultWoLPort,
									Interval:   15 * time.Minute,
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
				mac:       &entity.MacAddress{MAC: "00:00:00:00:00:00"},
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
									Name:       "test server",
									MacAddress: nil,
									Broadcast:  "192.168.1.255",
									Port:       entity.DefaultWoLPort,
									Interval:   15 * time.Minute,
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
				mac:       &entity.MacAddress{MAC: "00:11:22:33:44:55"},
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
									Name:       "test server",
									MacAddress: &entity.MacAddress{MAC: "00:11:22:33:44:55"},
									Broadcast:  "192.168.1.255",
									Port:       entity.DefaultWoLPort,
									Interval:   15 * time.Minute,
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
				mac:       &entity.MacAddress{MAC: "00:11:22:33:44:55"},
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
			ruleRepo.EXPECT().Evaluate(gomock.Any(), gomock.Any()).Return(tt.fields.rulesRepo.err).Times(tt.fields.rulesRepo.times)

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
