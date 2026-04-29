package evaluator

import (
	"errors"
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
)

var (
	ErrFailedReadRegoFile       = errors.New("failed to read rego file")
	ErrFailedEvaluateExpression = errors.New("could not evaluate expression")
)

type RegoEvaluator struct {
	config   *entity.Config
	ruleRepo repository.RuleRepository
	upsRepo  repository.UPSRepository
	mac      string
}

type EvaluationResult struct {
	Target  *entity.TargetServer
	Allowed bool
	Found   bool
}

func NewRegoEvaluator(config *entity.Config, mac string, upsRepo repository.UPSRepository, ruleRepo repository.RuleRepository) *RegoEvaluator {
	return &RegoEvaluator{
		config:   config,
		mac:      mac,
		upsRepo:  upsRepo,
		ruleRepo: ruleRepo,
	}
}

func (r *RegoEvaluator) EvaluateExpressions() (EvaluationResult, error) {
	// For each NUT server
	evaluationResult := EvaluationResult{
		Allowed: false,
		Found:   false,
		Target:  nil,
	}

	for _, nutServer := range r.config.NutServers {
		inputJSON, err := r.upsRepo.GetJSON(nutServer)
		if err != nil {
			return EvaluationResult{
				Allowed: false,
				Found:   false,
				Target:  nil,
			}, err
		}

		// For each target
		for _, target := range nutServer.Targets {
			if target.MAC != r.mac {
				continue
			}
			allowed, err := r.evaluateExpression(target, inputJSON)
			if err != nil {
				return EvaluationResult{}, err
			}

			evaluationResult.Found = true
			evaluationResult.Allowed = evaluationResult.Allowed || allowed
			evaluationResult.Target = target
		}
	}

	return evaluationResult, nil
}

func (r *RegoEvaluator) evaluateExpression(target *entity.TargetServer, inputJSON string) (bool, error) {
	if target == nil {
		return false, nil
	}

	for _, ruleName := range target.Rules {
		allowed, err := r.ruleRepo.Evaluate(ruleName, inputJSON)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrFailedEvaluateExpression, err)
		}

		if allowed {
			return true, nil
		}
	}
	return false, nil
}
