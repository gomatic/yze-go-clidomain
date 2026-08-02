// Package shape is grouping vocabulary shared by sibling verbs. It declares
// none of the contract, so it is not a verb — and its Flags type carries the
// behaviour that a verb's Config must not.
package shape

// Flags is a shared flag record with behaviour.
type Flags struct{ Name string }

// Validate is ordinary behaviour here, where it is allowed.
func (f Flags) Validate() error { return nil }
