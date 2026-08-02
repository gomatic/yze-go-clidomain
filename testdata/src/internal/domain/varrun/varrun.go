// Package varrun declares Run as a package-level var: the entry point the
// contract names does not exist, whatever the var's type.
package varrun

// Config holds the bound flags.
type Config struct{ Name string }

// Result is the outcome.
type Result struct{ Out string }

// Run is not a function declaration.
var Run = func(cfg Config) (Result, error) { return Result{Out: cfg.Name}, nil } // want "Run must be a function"
