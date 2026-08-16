// Package create borrows the SIBLING group internal/domain/tenant, whose path
// is a string prefix of this verb's without being a path prefix of it —
// tenantx is not tenant, and a prefix test that forgets the separator cannot
// tell them apart.
package create

import (
	"context"
	"log/slog"

	domain "internal/domain/tenant"
)

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run borrows a sibling group's vocabulary.
func Run(
	_ context.Context, _ *slog.Logger, cfg Config,
	args ...domain.Argument, // want `import alias for "internal/domain/tenant", not the shared "internal/domain"`
) (Result, error) {
	_ = args
	return Result{Out: cfg.Name}, nil
}
