package evaluator

import (
	"errors"
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/rego"
	"github.com/TheDarthMole/UPSWake/internal/ups"
	"github.com/spf13/afero"
)

var (
	ErrFailedReadRegoFile       = errors.New("failed to read rego file")
	ErrFailedEvaluateExpression = errors.New("could not evaluate expression")
)

type RegoEvaluator struct {
	config  *entity.Config
	rulesFS afero.Fs
	mac     string
}

type EvaluationResult struct {
	Allowed bool
	Found   bool
	Target  *entity.TargetServer
}

func NewRegoEvaluator(config *entity.Config, mac string, rulesFS afero.Fs) *RegoEvaluator {
	return &RegoEvaluator{
		config:  config,
		mac:     mac,
		rulesFS: rulesFS,
	}
}

func (r *RegoEvaluator) EvaluateExpressions() (EvaluationResult, error) {
	return r.evaluateExpressions(ups.GetJSON)
}

// EvaluateExpressions evaluates the expressions in the rules files
func (r *RegoEvaluator) evaluateExpressions(getUPSJSON func(server *entity.NutServer) (string, error)) (EvaluationResult, error) {
	// For each NUT server
	evaluationResult := EvaluationResult{
		Allowed: false,
		Found:   false,
		Target:  nil,
	}

	for _, nutServer := range r.config.NutServers {
		inputJSON, err := getUPSJSON(&nutServer)
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
			allowed, err := r.evaluateExpression(&target, inputJSON)
			if err != nil {
				return EvaluationResult{}, err
			}

			evaluationResult.Found = true
			evaluationResult.Allowed = evaluationResult.Allowed || allowed
			evaluationResult.Target = &target
		}
	}

	return evaluationResult, nil
}

func (r *RegoEvaluator) evaluateExpression(target *entity.TargetServer, inputJSON string) (bool, error) {
	if target == nil {
		return false, nil
	}

	for _, ruleName := range target.Rules {
		regoRule, err := afero.ReadFile(r.rulesFS, ruleName)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrFailedReadRegoFile, err)
		}
		allowed, err := rego.EvaluateExpression(inputJSON, string(regoRule))
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrFailedEvaluateExpression, err)
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}
