// Package borrowed is a verb of the `other` group that takes its vocabulary
// from a DIFFERENT group's package — not an ancestor of this verb, so it is
// the same drift an outside impostor is, however it is aliased.
package borrowed

import (
	"context"
	"log/slog"

	domain "internal/domain/grouped"
)

// Config holds the bound flags.
type Config struct{ Key string }

// Result is the outcome.
type Result struct{ Out string }

// Run borrows another group's vocabulary.
func Run(
	_ context.Context,
	_ *slog.Logger,
	_ Config,
	_ ...domain.Argument, // want `not the shared .* vocabulary package`
) (Result, error) {
	return Result{}, nil
}
