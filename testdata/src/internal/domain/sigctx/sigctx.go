// Package sigctx breaks exactly ONE element of the Run contract: its first
// parameter is a Context from a LOOK-ALIKE package imported as context, which
// reproduces the contract's spelling while carrying none of its meaning.
package sigctx

import (
	context "internal/context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run takes the look-alike Context.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument,
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
