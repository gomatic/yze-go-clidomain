// Package signoresult breaks exactly ONE element of the Run contract: it
// returns NOTHING, so the caller learns neither the outcome nor the failure.
// A result arity the analyzer does not check first is also a result it indexes
// out of range.
package signoresult

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome, declared and never returned.
type Result struct{ Out string }

// Run returns no results at all.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) { // want "Run must be a function"
	_, _ = cfg, args
}
