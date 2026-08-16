// Package generic declares a GENERIC Run. Its parameters read exactly like the
// contract's, and no caller can write the call: the type argument is inferable
// from nothing, so the entry point the contract names does not exist.
package generic

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run carries a type parameter no call site can infer.
func Run[T any]( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument,
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
