package evaluator

import (
	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/infrastructure/config/viper"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"

	"reflect"
	"testing"
)

var (
	defaultConfig, _ = viper.CreateDefaultConfig()
	tempFS, _        = mem.NewFS()
)

func TestNewRegoEvaluator(t *testing.T) {
	type args struct {
		config  *entity.Config
		mac     string
		rulesFS hackpadfs.FS
	}

	tests := []struct {
		name string
		args args
		want *regoEvaluator
	}{
		{
			name: "valid config",
			args: args{
				config:  defaultConfig,
				mac:     "00:00:00:00:00:00",
				rulesFS: tempFS,
			},
			want: &regoEvaluator{
				config:  defaultConfig,
				rulesFS: tempFS,
				mac:     "00:00:00:00:00:00",
			},
		},
		//	TODO: Add more test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRegoEvaluator(tt.args.config, tt.args.mac, tt.args.rulesFS); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRegoEvaluator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegoEvaluator_EvaluateExpressions(t *testing.T) {
	type fields struct {
		config  *entity.Config
		rulesFS hackpadfs.FS
		mac     string
	}
	var tests []struct {
		name   string
		fields fields
		want   EvaluationResult
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &regoEvaluator{
				config:  tt.fields.config,
				rulesFS: tt.fields.rulesFS,
				mac:     tt.fields.mac,
			}
			if got := r.EvaluateExpressions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateExpressions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegoEvaluator_evaluateExpression(t *testing.T) {
	type fields struct {
		config  *entity.Config
		rulesFS hackpadfs.FS
		mac     string
	}
	type args struct {
		target    *entity.TargetServer
		nutServer *entity.NutServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "nothing to evaluate",
			fields: fields{
				config:  defaultConfig,
				rulesFS: tempFS,
				mac:     "00:00:00:00:00:00",
			},
			args: args{
				target:    nil,
				nutServer: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &regoEvaluator{
				config:  tt.fields.config,
				rulesFS: tt.fields.rulesFS,
				mac:     tt.fields.mac,
			}
			got, err := r.evaluateExpression(tt.args.target, tt.args.nutServer)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("evaluateExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}
