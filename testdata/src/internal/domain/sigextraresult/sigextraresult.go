// Package sigextraresult breaks exactly ONE element of the Run contract: it
// returns a THIRD result. Every other element is exact, including both results
// the contract names, so only the arity is wrong.
package sigextraresult

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run returns a second error beside the contract's two results.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument,
) (Result, error, error) {
	_ = args
	return Result{Out: cfg.Name}, nil, nil
}
