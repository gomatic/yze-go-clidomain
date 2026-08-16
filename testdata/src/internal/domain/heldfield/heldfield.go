// Package heldfield is the CONFORMING near-miss of the embedding case: its
// Config HOLDS a behaviour-carrying type in a named field rather than
// embedding it, so no method is promoted onto Config and Run reaches the
// behaviour through the field. That is the whole remedy an embedded Config is
// asked for — one word — which is what keeps complying cheaper than evading.
package heldfield

import (
	"context"
	"log/slog"

	"internal/domain"
	"internal/domain/shape"
)

// Config holds the bound flags, one of which carries its own behaviour.
type Config struct {
	Flags shape.Flags
}

// Result is the outcome.
type Result struct{ Out string }

// Run reaches the behaviour through the field.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	if err := cfg.Flags.Validate(); err != nil {
		return Result{}, err
	}
	return Result{Out: cfg.Flags.Name}, nil
}
