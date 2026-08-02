// Package tenant is a grouping package holding helpers its nested verb
// packages share. It declares no element of the command contract, so it is not
// a verb and nothing is reported — depth alone never makes a package a verb.
package tenant

// Normalize canonicalises a tenant name for the verbs beneath this package.
func Normalize(raw string) string { return raw }
