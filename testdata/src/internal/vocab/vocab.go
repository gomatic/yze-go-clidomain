// Package vocab redeclares the shared vocabulary OUTSIDE the domain tier, so
// aliasing it as domain reproduces the shared spelling without the shared
// package.
package vocab

// Argument shadows the shared domain.Argument — the drift.
type Argument = string
