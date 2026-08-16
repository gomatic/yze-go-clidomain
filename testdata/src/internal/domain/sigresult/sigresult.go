// Package sigresult breaks exactly ONE element of the Run contract: Run's
// first result is a FOREIGN type, leaving the Result declared here dead. Its
// Config is this package's.
package sigresult

import (
	"context"
	"log/slog"

	"internal/domain"
	"internal/domain/shape"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome — and Run below never returns it.
type Result struct{ Out string }

// Run returns a foreign Result.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument,
) (shape.Flags, error) {
	_ = args
	return shape.Flags{Name: cfg.Name}, nil
}
