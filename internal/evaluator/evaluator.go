package evaluator

import (
	"fmt"
	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/rego"
	"github.com/TheDarthMole/UPSWake/internal/ups"
	"github.com/TheDarthMole/UPSWake/internal/util"
	"github.com/hack-pad/hackpadfs"
)

type regoEvaluator struct {
	config  *config.Config
	rulesFS hackpadfs.FS
	mac     string
}

type EvaluationResult struct {
	Allowed bool
	Found   bool
	Error   error
	Target  *config.TargetServer
}

func NewRegoEvaluator(config *config.Config, mac string, rulesFS hackpadfs.FS) *regoEvaluator {
	return &regoEvaluator{
		config:  config,
		mac:     mac,
		rulesFS: rulesFS,
	}
}

// EvaluateExpressions evaluates the expressions in the rules files
func (r *regoEvaluator) EvaluateExpressions() EvaluationResult {
	// For each NUT server
	var evaluationResults []EvaluationResult

	for _, nutServer := range r.config.NutServers {
		// For each target
		for _, target := range nutServer.Targets {
			if target.MAC == r.mac {
				allowed, err := r.evaluateExpression(&target, &nutServer)
				evaluationResults = append(evaluationResults, EvaluationResult{
					Allowed: allowed,
					Found:   true,
					Error:   err,
					Target:  &target,
				})
			}
		}
	}

	return EvaluationResult{
		Allowed: false,
		Found:   false,
		Error:   nil,
		Target:  nil,
	}

	//target, nutServer, err := r.config.FindTarget(r.mac)
	//
	//if err != nil {
	//	return EvaluationResult{
	//		Allowed: false,
	//		Found:   target != nil, // couldn't find the target in the config
	//		Error:   err,
	//		Target:  nil,
	//	}
	//}
	//
	//allowed, err := r.evaluateExpression(target, nutServer)
	//
	//return EvaluationResult{
	//	Allowed: allowed,
	//	Found:   true,
	//	Error:   err,
	//	Target:  target,
	//}
}

func (r *regoEvaluator) evaluateExpression(target *config.TargetServer, nutServer *config.NutServer) (bool, error) {
	if target == nil || nutServer == nil {
		return false, nil
	}

	inputJson, err := ups.GetJSON(nutServer)
	if err != nil {
		return false, err
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
