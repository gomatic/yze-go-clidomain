// Package sigconfig breaks exactly ONE element of the Run contract: Run's
// third parameter is a FOREIGN flag record, leaving the Config declared here
// dead while the CLI tier binds into it. Its Result is this package's.
package sigconfig

import (
	"context"
	"log/slog"

	"internal/domain"
	"internal/domain/shape"
)

// Config holds the bound flags — and Run below never reads it.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run takes a foreign Config.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg shape.Flags, args ...domain.Argument,
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
