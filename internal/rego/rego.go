// SPDX-License-Identifier: MPL-2.0
package rego

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

var (
	ErrDecodeFailed       = errors.New("failed to decode input")
	ErrFailedEvaluateRule = errors.New("failed to evaluate rego rule")
	ErrInvalidRegoRule    = errors.New("invalid rego rule")
	ErrPackageName        = errors.New("rego rule must be in package 'upswake'")
)

func IsValidRego(rego string) error {
	mod, err := ast.ParseModule("test", rego)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidRegoRule, err)
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
