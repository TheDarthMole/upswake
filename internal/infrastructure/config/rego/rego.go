// SPDX-License-Identifier: MPL-2.0
package rego

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

var (
	ErrDecodeFailed       = errors.New("failed to decode input")
	ErrFailedCreatingRule = errors.New("failed to create rule")
	ErrFailedEvaluateRule = errors.New("failed to evaluate rego rule")
	ErrInvalidRegoRule    = errors.New("invalid rego rule")
	ErrPackageName        = errors.New("rego rule must be in package 'upswake'")
)

func IsValidRego(rego string) error {
	mod, err := ast.ParseModule("test", rego)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRegoRule, err)
	}
	if mod.Package.String() != "package upswake" {
		return ErrPackageName
	}
	return nil
}

func EvaluateExpression(rawJSON, regoRule string) (bool, error) {
	var input any
	ctx := context.Background()

	if err := IsValidRego(regoRule); err != nil {
		return false, err
	}

	d := json.NewDecoder(bytes.NewBufferString(rawJSON))

	// Numeric values must be represented using json.Number.
	d.UseNumber()
	d.DisallowUnknownFields()

	if err := d.Decode(&input); err != nil {
		return false, fmt.Errorf("%w: %w", ErrDecodeFailed, err)
	}
	// Create query that returns a single boolean value.
	regoEngine := rego.New(
		rego.Query("data.upswake.wake"),
		rego.Module(
			"main.rego",
			regoRule,
		),
		rego.Input(input),
	)

	// Run evaluation.
	rs, err := regoEngine.Eval(ctx)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrFailedEvaluateRule, err)
	}

	return rs.Allowed(), nil
}

type Evaluator struct {
	queries []rego.PreparedEvalQuery
}

func (e *Evaluator) Evaluate(ctx context.Context, rawJSON string) (bool, error) {
	for _, query := range e.queries {
		rs, err := query.Eval(ctx, rego.EvalInput(rawJSON))
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrFailedEvaluateRule, err)
		}
		if rs.Allowed() {
			return true, nil
		}
	}
	return false, nil
}

func NewRegoEvaluator() *Evaluator {
	return &Evaluator{
		queries: nil,
	}
}

func (e *Evaluator) Load(ctx context.Context, cfg *entity.Config) error {
	for _, nutServer := range cfg.NutServers {
		for _, target := range nutServer.Targets {
			queries, err := newQueries(ctx, target.RulesContent)
			if err != nil {
				return err
			}
			target.Evaluator = &Evaluator{
				queries: queries,
			}
		}
	}
	return nil
}

func newQuery(ctx context.Context, ruleContents string) (rego.PreparedEvalQuery, error) {
	a := rego.New(
		rego.Query("data.upswake.wake"),
		rego.Module(
			"main.rego",
			ruleContents,
		),
	)
	return a.PrepareForEval(ctx)
}

func newQueries(ctx context.Context, ruleContents []string) ([]rego.PreparedEvalQuery, error) {
	queries := make([]rego.PreparedEvalQuery, len(ruleContents))

	for index, rule := range ruleContents {
		query, err := newQuery(ctx, rule)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedCreatingRule, err)
		}
		queries[index] = query
	}

	return queries, nil
}
