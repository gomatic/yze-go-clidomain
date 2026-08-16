// Package siglogger breaks exactly ONE element of the Run contract: it takes
// the logger by VALUE rather than by pointer.
package siglogger

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run takes a slog.Logger value.
func Run( // want "Run must be a function"
	_ context.Context, _ slog.Logger, cfg Config, args ...domain.Argument,
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
