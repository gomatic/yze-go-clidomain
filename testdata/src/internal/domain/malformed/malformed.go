package malformed

import "context"

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run is declared but breaks the contract: no logger, not variadic, and it
// returns a bare error rather than (Result, error).
func Run(_ context.Context, cfg Config) error { // want "Run must be a function"
	_ = cfg
	return nil
}
