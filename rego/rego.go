package rego

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/rego"
)

func EvaluateExpression(raw1 string) (bool, error) {
	ctx := context.Background()

	raw := raw1

	d := json.NewDecoder(bytes.NewBufferString(raw))

	// Numeric values must be represented using json.Number.
	d.UseNumber()

	var input interface{}

	if err := d.Decode(&input); err != nil {
		panic(err)
	}
	// Create query that returns a single boolean value.
	rego := rego.New(
		rego.Query("data.authz.allow"),
		rego.Module("example.rego",
			`package authz

default allow = false
allow = true {
	input[_].Name == "cyberpower900"
}`,
		),
		rego.Input(input),
	)

	// Run evaluation.
	rs, err := rego.Eval(ctx)
	if err != nil {
		panic(err)
	}

	// Inspect result.
	fmt.Println("allowed:", rs.Allowed())

	return rs.Allowed(), nil
}
