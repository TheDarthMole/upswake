package evaluator

import (
	"fmt"
	"github.com/TheDarthMole/UPSWake/config"
	"github.com/TheDarthMole/UPSWake/rego"
	"github.com/TheDarthMole/UPSWake/ups"
	"github.com/TheDarthMole/UPSWake/util"
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
	found := false
	// For each NUT server
	for _, mapping := range r.config.NutServerMappings {

		// For each client defined for the server
		for _, target := range mapping.Targets {
			if target.Mac == r.mac {
				found = true
				allowed, err := r.evaluateExpression(&target, &mapping.NutServer)
				if err != nil {
					return EvaluationResult{
						Allowed: false,
						Found:   true,
						Error:   err,
						Target:  &target,
					}
				}
				if allowed {
					return EvaluationResult{
						Allowed: true,
						Found:   true,
						Error:   nil,
						Target:  &target,
					}
				}
			}
		}
	}

	return EvaluationResult{
		Allowed: false,
		Found:   found,
		Error:   nil,
		Target:  nil,
	}
}

func (r *regoEvaluator) evaluateExpression(target *config.TargetServer, nutServer *config.NutServer) (bool, error) {
	if target == nil || nutServer == nil {
		return false, nil
	}

	inputJson, err := ups.GetJSON(nutServer)
	if err != nil {
		return false, err
	}
	for _, ruleName := range target.Config.Rules {
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
