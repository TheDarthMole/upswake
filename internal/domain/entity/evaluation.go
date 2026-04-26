package entity

import "context"

type Evaluator interface {
	Evaluate(ctx context.Context, nutJSON string) (bool, error)
}

type EvaluationResult struct {
	Target  *TargetServer
	Allowed bool
	Found   bool
}
