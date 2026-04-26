package evaluator

import (
	"context"
	"errors"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
)

var (
	ErrFailedReadFile           = errors.New("failed to read file")
	ErrFailedEvaluateExpression = errors.New("could not evaluate expression")
)

type Evaluator struct {
	config *entity.Config
	entity.Evaluator
}

func NewRulesEvaluator(evaluator entity.Evaluator, config *entity.Config) *Evaluator {
	return &Evaluator{
		Evaluator: evaluator,
		config:    config,
	}
}

func (r *Evaluator) EvaluateExpressions(ctx context.Context, upsJSON, mac string) (entity.EvaluationResult, error) {
	evaluationResult := entity.EvaluationResult{
		Allowed: false,
		Found:   false,
		Target:  nil,
	}

	for _, nutServer := range r.config.NutServers {
		// For each target
		for _, target := range nutServer.Targets {
			if target.MAC != mac {
				continue
			}
			allowed, err := r.Evaluate(ctx, upsJSON)
			if err != nil {
				return entity.EvaluationResult{}, err
			}

			evaluationResult.Found = true
			evaluationResult.Allowed = evaluationResult.Allowed || allowed
			evaluationResult.Target = target
		}
	}

	return evaluationResult, nil
}
