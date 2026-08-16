// Package context is a LOOK-ALIKE of the standard library's, declaring a type
// of the same name. Importing it as context reproduces the contract's spelling
// exactly while defeating it.
package context

// Context shadows the standard library's cancellation type.
type Context interface {
	Done() <-chan struct{}
}
