// SPDX-License-Identifier: MPL-2.0
package rego

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

func IsValidRego(rego string) error {
	mod, err := ast.ParseModule("test", rego)
	if err != nil {
		return err
	}
	if mod.Package.String() != "package upswake" {
		return fmt.Errorf("rego rule must be in package 'upswake'")
	}
	return nil
}

func EvaluateExpression(rawJson, regoRule string) (bool, error) {
	var input interface{}
	ctx := context.Background()

	if err := IsValidRego(regoRule); err != nil {
		return false, err
	}

	d := json.NewDecoder(bytes.NewBufferString(rawJson))

	// Numeric values must be represented using json.Number.
	d.UseNumber()
	d.DisallowUnknownFields()

	if err := d.Decode(&input); err != nil {
		return false, fmt.Errorf("failed to decode input: %w", err)
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
		return false, fmt.Errorf("failed to evaluate rego rule: %w", err)
	}

	return rs.Allowed(), nil
}
