// Package sigerror breaks exactly ONE element of the Run contract: its second
// result is a message rather than an error, so a composed caller cannot tell
// one failure from another with errors.Is.
package sigerror

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Reason is the message Run returns instead of an error.
type Reason string

// Run returns a Reason where the contract demands an error.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument,
) (Result, Reason) {
	_ = args
	return Result{Out: cfg.Name}, ""
}
