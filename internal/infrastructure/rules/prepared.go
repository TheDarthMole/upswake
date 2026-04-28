package rules

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/spf13/afero"
)

var (
	ErrRuleNotFound    = errors.New("rule not found")
	ErrDecodeFailed    = errors.New("failed to decode input JSON")
	ErrEvaluationError = errors.New("failed to evaluate rego rule")
	ErrCompileError    = errors.New("failed to compile rego rule")
	ErrInvalidRegoRule = errors.New("invalid rego rule")
	ErrPackageName     = errors.New("rego rule must be in package 'upswake'")
)

// PreparedRepository loads and pre-compiles all Rego rules from the
// filesystem at construction time. Evaluate() only runs the prepared
// query against new input, skipping parsing and compilation entirely.
type PreparedRepository struct {
	rules map[string]rego.PreparedEvalQuery
}

// NewPreparedRepository reads every .rego file from fs, validates it,
// and compiles it into a PreparedEvalQuery. Returns an error if any
// rule fails validation or compilation.
func NewPreparedRepository(fs afero.Fs) (repository.RuleRepository, error) {
	rules := make(map[string]rego.PreparedEvalQuery)

	entries, err := afero.ReadDir(fs, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read rules directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		raw, readErr := afero.ReadFile(fs, name)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read rule %s: %w", name, readErr)
		}

		if err = IsValidRego(string(raw)); err != nil {
			return nil, fmt.Errorf("invalid rule %s: %w", name, err)
		}

		prepared, prepErr := prepareRule(name, string(raw))
		if prepErr != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrCompileError, name, prepErr)
		}
		rules[name] = prepared
	}

	return &PreparedRepository{rules: rules}, nil
}

func prepareRule(name, raw string) (rego.PreparedEvalQuery, error) {
	r := rego.New(
		rego.Query("data.upswake.wake"),
		rego.Module(name, raw),
	)
	return r.PrepareForEval(context.Background())
}

func (r *PreparedRepository) Evaluate(ruleName, inputJSON string) (bool, error) {
	prepared, ok := r.rules[ruleName]
	if !ok {
		return false, fmt.Errorf("%w: %s", ErrRuleNotFound, ruleName)
	}

	var input any
	d := json.NewDecoder(bytes.NewBufferString(inputJSON))
	d.UseNumber()
	d.DisallowUnknownFields()

	if err := d.Decode(&input); err != nil {
		return false, fmt.Errorf("%w: %w", ErrDecodeFailed, err)
	}

	rs, err := prepared.Eval(context.Background(), rego.EvalInput(input))
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrEvaluationError, err)
	}

	return rs.Allowed(), nil
}

func (r *PreparedRepository) RuleNames() []string {
	names := make([]string, 0, len(r.rules))
	for name := range r.rules {
		names = append(names, name)
	}
	return names
}

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
