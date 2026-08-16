// Package signonvariadic breaks exactly ONE element of the Run contract: its
// last parameter is a single domain.Argument rather than a variadic one, so a
// verb taking two arguments cannot be called at all.
package signonvariadic

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run takes one argument rather than a variadic list.
func Run(_ context.Context, _ *slog.Logger, cfg Config, _ domain.Argument) (Result, error) { // want "Run must be a function"
	return Result{Out: cfg.Name}, nil
}
