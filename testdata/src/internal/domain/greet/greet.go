// Package greet is a CONFORMANT domain package: Config holds the bound flags
// and declares no behaviour, Result is the outcome, and Run has the exact
// contract signature using the shared domain.Argument. Nothing is reported.
package greet

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the flags the CLI tier binds; it carries no behavior.
type Config struct {
	Greeting string
}

// Result is the outcome of the greet command.
type Result struct {
	Message string
}

// Run orchestrates the greeting.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	if len(args) == 0 {
		return Result{}, nil
	}
	return Result{Message: cfg.Greeting + " " + args[0]}, nil
}
