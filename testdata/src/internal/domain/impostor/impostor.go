// Package impostor declares its own Config and Result but wires Run to a
// FOREIGN package's type — the declared contract types are dead, and the CLI
// tier would bind into a Config that Run never reads.
package impostor

import (
	"context"
	"log/slog"

	"internal/domain"
	"internal/domain/shape"
)

// Config holds the bound flags — and Run below ignores it.
type Config struct{ Name string }

// Result is the outcome — equally dead.
type Result struct{ Out string }

// Run uses shape.Flags where the contract demands THIS package's
// Config/Result.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg shape.Flags, args ...domain.Argument,
) (shape.Flags, error) {
	_ = args
	return cfg, nil
}
