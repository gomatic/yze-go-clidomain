// Package vocabalias spells its variadic domain.Argument, but domain is an
// import alias for internal/vocab — the impostor the shared-vocabulary rule
// exists to catch.
package vocabalias

import (
	"context"
	"log/slog"

	domain "internal/vocab"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run spells the shared alias while importing an impostor package as domain.
func Run(
	_ context.Context, _ *slog.Logger, cfg Config,
	args ...domain.Argument, // want `import alias for "internal/vocab", not the shared "internal/domain"`
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
