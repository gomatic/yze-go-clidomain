// Package list is a CONFORMANT nested domain verb: the full contract at depth
// two, using the shared domain.Argument. Nothing is reported.
package list

import (
	"context"
	"log/slog"

	"internal/domain"
)

// Config holds the flags the CLI tier binds; it carries no behavior.
type Config struct {
	Limit int
}

// Result is the outcome of the tenant list command.
type Result struct {
	Names []string
}

// Run orchestrates the listing.
func Run(_ context.Context, _ *slog.Logger, cfg Config, args ...domain.Argument) (Result, error) {
	names := make([]string, 0, cfg.Limit)
	names = append(names, args...)
	return Result{Names: names}, nil
}
