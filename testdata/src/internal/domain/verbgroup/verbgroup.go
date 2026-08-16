// Package verbgroup is itself a full VERB that also exports an Argument alias.
// A verb's types are its own; being an ancestor of another verb does not make
// them that verb's vocabulary.
package verbgroup

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Argument is this verb's own redeclaration of the shared alias.
type Argument = string

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run orchestrates.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
