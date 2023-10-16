package rego

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/rego"
)

func EvaluateExpression(rawJson, regoRule string) (bool, error) {
	var input interface{}
	ctx := context.Background()
	d := json.NewDecoder(bytes.NewBufferString(rawJson))

	// Numeric values must be represented using json.Number.
	d.UseNumber()
	//d.DisallowUnknownFields() // TODO: does this work?

	if err := d.Decode(&input); err != nil {
		return false, fmt.Errorf("failed to decode input: %w", err)
	}
	// Create query that returns a single boolean value.
	rego := rego.New(
		rego.Query("data.authz.allow"),
		rego.Module(
			"example.rego",
			regoRule,
		),
		rego.Input(input),
	)

	// Run evaluation.
	rs, err := rego.Eval(ctx)
	if err != nil {
		panic(err)
	}

	return rs.Allowed(), nil
}
