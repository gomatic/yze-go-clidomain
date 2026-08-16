// Package multi is a verb spread over two PRODUCTION files, missing Config and
// Run. The report belongs at this file's package clause: the runner drops a
// finding a _test.go file holds, so anchoring one there deletes it.
package multi // want "must declare a Config struct" "must declare the entry point Run"

// Result is the outcome, so this package sets out to be a verb.
type Result struct{ Out string }
