// Package sub takes its vocabulary from its PARENT VERB rather than from a
// group package. The parent declares the contract, so it is a command and not
// a vocabulary its subcommands inherit.
package sub

import (
	"context"
	"log/slog"

	domain "internal/domain/verbgroup"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run speaks the parent verb's private alias.
func Run(
	_ context.Context, _ *slog.Logger, cfg Config,
	args ...domain.Argument, // want `import alias for "internal/domain/verbgroup", not the shared "internal/domain"`
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
