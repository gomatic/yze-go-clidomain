// Package exttest is a fully conformant verb whose EXTERNAL test package
// declares a Run-shaped driver; the scaffolding passes are never judged, so
// nothing is reported.
package exttest

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run orchestrates the verb.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
