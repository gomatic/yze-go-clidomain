// Package noconfig declares Run and Result but NO Config, so Run is wired to a
// foreign flag record. Config's absence must be reported, and the identity
// check must survive looking a declaration up that is not there.
package noconfig // want "must declare a Config struct"

import (
	"context"
	"log/slog"

	"internal/domain"
	"internal/domain/shape"
)

// Result is the outcome, so this package sets out to be a verb.
type Result struct{ Out string }

// Run takes a foreign flag record because this package declares no Config.
func Run( // want "Run must be a function"
	_ context.Context, _ *slog.Logger, cfg shape.Flags, args ...domain.Argument,
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
