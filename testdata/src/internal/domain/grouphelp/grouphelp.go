// Package grouphelp is a grouping package: helpers only, no contract. Its
// in-package test declares a Config-shaped table type, which must not turn
// this package into a verb — test files own their own vocabulary.
package grouphelp

// Normalize is the shared helper the sibling verbs reuse.
func Normalize(raw string) string { return raw }
