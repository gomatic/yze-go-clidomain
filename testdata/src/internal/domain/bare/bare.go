package bare

import (
	"context"
	"log/slog"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run matches the contract's shape but spells its variadic as a bare string
// rather than the shared domain.Argument, so the concept goes unnamed.
func Run(
	_ context.Context, _ *slog.Logger, cfg Config,
	args ...string, // want "shared domain.Argument alias"
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
