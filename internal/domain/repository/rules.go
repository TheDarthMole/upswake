package repository

//go:generate mockgen -package mocks -source rules.go -destination mocks/rules_mock.go RuleRepository

// RuleRepository provides access to evaluate Rego rules.
// Implementations should load and compile rules once at startup to avoid
// repeated filesystem reads and OPA compilation on every evaluation cycle.
type RuleRepository interface {
	// Evaluate evaluates a named rule against the provided JSON input.
	// Should return nil if no errors occurred and at least one rule evaluated to true
	// Should return an entity.ErrEvaluationFalse if all rules evaluate false
	Evaluate(ruleName, inputJSON string) error

	// RuleNames returns all available rule names.
	RuleNames() []string
}
