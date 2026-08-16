// Package ptrmethod declares its Config's behaviour on a POINTER receiver,
// which the value's own method set omits — behaviour all the same.
package ptrmethod

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the bound flags and a pointer-receiver method.
type Config struct { // want "Config must carry no behaviour"
	Name string
}

// Normalize is behaviour, on a pointer receiver.
func (c *Config) Normalize() { c.Name = c.Name + "!" }

// Result is the outcome.
type Result struct{ Out string }

// Run orchestrates.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	cfg.Normalize()
	return Result{Out: cfg.Name}, nil
}
