// Package iface declares its Config as an INTERFACE promising behaviour. It
// declares no method itself, and every method it promises is callable inside
// Run — the same logic in the same tier.
package iface

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config promises behaviour rather than holding flags.
type Config interface { // want "Config must carry no behaviour"
	Validate() error
	Name() string
}

// Result is the outcome.
type Result struct{ Out string }

// Run calls the promised behaviour.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	if err := cfg.Validate(); err != nil {
		return Result{}, err
	}
	return Result{Out: cfg.Name()}, nil
}
