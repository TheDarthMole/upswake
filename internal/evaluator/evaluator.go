package evaluator

import (
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/rego"
	"github.com/TheDarthMole/UPSWake/internal/ups"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/hack-pad/hackpadfs"
)

type RegoEvaluator struct {
	config  *entity.Config
	rulesFS hackpadfs.FS
	mac     string
}

type EvaluationResult struct {
	Allowed bool
	Found   bool
	Target  *entity.TargetServer
}

func NewRegoEvaluator(config *entity.Config, mac string, rulesFS hackpadfs.FS) *RegoEvaluator {
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
		inputJson, err := getUPSJSON(&nutServer)
		if err != nil {
			return EvaluationResult{
				Allowed: false,
				Found:   false,
				Target:  nil,
			}, err
		}

		// For each target
		for _, target := range nutServer.Targets {
			if target.MAC == r.mac {
				allowed, err := r.evaluateExpression(&target, inputJson)
				if err != nil {
					return EvaluationResult{
						Allowed: false,
						Found:   true,
						Target:  &target,
					}, err
				}

				evaluationResult.Found = true
				evaluationResult.Allowed = evaluationResult.Allowed || allowed
				evaluationResult.Target = &target
			}
		}
	}

	return evaluationResult, nil
}

func (r *RegoEvaluator) evaluateExpression(target *entity.TargetServer, inputJson string) (bool, error) {
	if target == nil {
		return false, nil
	}

	for _, ruleName := range target.Rules {
		regoRule, err := util.GetFile(r.rulesFS, ruleName)
		if err != nil {
			return false, fmt.Errorf("could not get file: %s", err)
		}
		allowed, err := rego.EvaluateExpression(inputJson, string(regoRule))
		if err != nil {
			return false, fmt.Errorf("could not evaluate expression: %s", err)
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}
