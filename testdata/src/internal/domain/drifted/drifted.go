package drifted

import (
	"context"
	"log/slog"
)

// argument is a package-local redeclaration of the shared domain.Argument —
// the drift this analyzer exists to catch.
type argument = string

// Config here wrongly carries behaviour.
type Config struct { // want "Config carries no behaviour"
	Name string
}

// Validate is behaviour on Config, which belongs in Run.
func (c Config) Validate() error { return nil }

// Result is the outcome.
type Result struct{ Out string }

// Run uses the local alias instead of domain.Argument.
func Run(
	_ context.Context, _ *slog.Logger, cfg Config,
	args ...argument, // want "shared domain.Argument alias"
) (Result, error) {
	return Result{Out: cfg.Name}, nil
}
