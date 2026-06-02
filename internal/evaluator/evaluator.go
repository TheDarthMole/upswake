package evaluator

import (
	"errors"
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
)

var ErrFailedEvaluateExpression = errors.New("could not evaluate expression")

type RegoEvaluator struct {
	config   *entity.Config
	ruleRepo repository.RuleRepository
	upsRepo  repository.UPSRepository
	mac      *entity.MacAddress
}

type EvaluationResult struct {
	Target  *entity.TargetServer
	Allowed bool
	Found   bool
}

// NewRegoEvaluator creates a RegoEvaluator configured with the provided configuration, MAC address,
// UPS repository and rule repository.
// The returned evaluator uses the MAC to select matching targets, upsRepo to fetch per-server JSON
// input and ruleRepo to evaluate rules against that input.
func NewRegoEvaluator(config *entity.Config, mac *entity.MacAddress, upsRepo repository.UPSRepository, ruleRepo repository.RuleRepository) *RegoEvaluator {
	return &RegoEvaluator{
		config:   config,
		mac:      mac,
		upsRepo:  upsRepo,
		ruleRepo: ruleRepo,
	}
}

func (r *RegoEvaluator) EvaluateExpressions() (*EvaluationResult, error) {
	// For each NUT server
	evaluationResult := &EvaluationResult{
		Allowed: false,
		Found:   false,
		Target:  nil,
	}

	for _, nutServer := range r.config.NutServers {
		inputJSON, err := r.upsRepo.GetJSON(nutServer)
		if err != nil {
			return nil, err
		}

		// For each target
		for _, target := range nutServer.Targets {
			if target.MacAddress == nil || r.mac == nil {
				return nil, fmt.Errorf("error comparing mac addresses: %w", entity.ErrMACRequired)
			}
			if target.MAC != r.mac.MAC {
				continue
			}
			err := r.evaluateExpression(target, inputJSON)
			if errors.Is(err, entity.ErrEvaluationFalse) {
				evaluationResult.Target = target
				evaluationResult.Found = true
				continue
			}
			if err != nil {
				return nil, err
			}

			return &EvaluationResult{
				Target:  target,
				Allowed: true,
				Found:   true,
			}, nil
		}
	}

	return evaluationResult, nil
}

func (r *RegoEvaluator) evaluateExpression(target *entity.TargetServer, inputJSON string) error {
	if target == nil {
		return fmt.Errorf("%w: target is nil", ErrFailedEvaluateExpression)
	}

	for _, ruleName := range target.Rules {
		err := r.ruleRepo.Evaluate(ruleName, inputJSON)
		if errors.Is(err, entity.ErrEvaluationFalse) {
			continue
		}
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedEvaluateExpression, err)
		}
		return nil
	}
	return entity.ErrEvaluationFalse
}
