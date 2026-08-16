// Package testmethods is a CONFORMANT verb whose Config declares no method in
// any production file. Its in-package test declares an ordinary table-test
// builder on Config, which is the tests' own vocabulary — nothing is reported.
package testmethods

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the flags the CLI tier binds; it carries no behaviour.
type Config struct{ Greeting string }

// Result is the outcome.
type Result struct{ Message string }

// Run orchestrates.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	return Result{Message: cfg.Greeting}, nil
}
