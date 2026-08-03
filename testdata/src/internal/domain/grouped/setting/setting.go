// Package setting is a verb of the grouped command group. It takes its
// vocabulary from the GROUP package rather than the internal/domain root —
// the sanctioned shape, so nothing is reported.
package setting

import (
	"context"
	"log/slog"

	domain "internal/domain/grouped"
)

// Config holds the bound flags.
type Config struct{ Key string }

// Result is the outcome.
type Result struct{ Out string }

// Run speaks the group's vocabulary.
func Run(_ context.Context, _ *slog.Logger, _ Config, _ ...domain.Argument) (Result, error) {
	return Result{}, nil
}
