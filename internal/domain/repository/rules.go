package repository

// RuleRepository provides access to pre-loaded and pre-compiled Rego rules.
// Implementations should load and compile rules once at startup to avoid
// repeated filesystem reads and OPA compilation on every evaluation cycle.
type RuleRepository interface {
	// Evaluate evaluates a named rule against the provided JSON input.
	// The rule should already be compiled; this only runs the evaluation.
	Evaluate(ruleName, inputJSON string) (bool, error)

	// RuleNames returns all available rule names.
	RuleNames() []string
}
