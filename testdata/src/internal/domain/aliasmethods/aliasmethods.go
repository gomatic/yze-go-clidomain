// Package aliasmethods aliases its Config to a foreign type that carries
// behaviour: the alias is legal Go and Run's signature is contract-identical,
// but the resolved Config still declares methods — reported at THIS package's
// alias declaration, where the fix is.
package aliasmethods

import (
	"context"
	"log/slog"

	"internal/domain"
	"internal/domain/shape"
)

// Config aliases the shared flag record — importing its behaviour.
type Config = shape.Flags // want "Config must carry no behaviour"

// Result is the outcome.
type Result struct{ Out string }

// Run is contract-identical: the alias resolves to the same type it names.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
