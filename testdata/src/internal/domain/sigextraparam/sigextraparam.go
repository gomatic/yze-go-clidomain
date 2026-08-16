// Package sigextraparam breaks exactly ONE element of the Run contract: it
// takes a FIFTH parameter, so the CLI tier's call does not compile.
package sigextraparam

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Extra is the parameter that does not belong in the signature.
type Extra string

// Run takes an extra parameter between Config and the variadic.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg Config, _ Extra, args ...domain.Argument,
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
