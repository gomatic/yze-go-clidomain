// Package embedded EMBEDS a behaviour-carrying type in its Config. The methods
// are promoted, so Run can call them: the same logic in the same tier as an
// alias, reached by the commonest spelling in Go.
package embedded

import (
	"context"
	"log/slog"

	"internal/domain"
	"internal/domain/shape"
)

// Config embeds the shared flag record — promoting its behaviour.
type Config struct { // want "Config must carry no behaviour"
	shape.Flags
}

// Result is the outcome.
type Result struct{ Out string }

// Run calls the promoted behaviour.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	_ = args
	if err := cfg.Validate(); err != nil {
		return Result{}, err
	}
	return Result{Out: cfg.Name}, nil
}
