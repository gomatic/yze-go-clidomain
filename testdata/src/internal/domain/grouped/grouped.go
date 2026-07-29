// Package grouped is an intermediate domain package holding logic its sibling
// verb subpackages share. It declares no element of the command contract, so it
// is not a verb and nothing is reported — the shape that made the reference
// layout's own internal/domain/config report three false findings.
package grouped

// Key validates and normalises a configuration key for the verbs that use it.
func Key(raw string) string { return raw }
