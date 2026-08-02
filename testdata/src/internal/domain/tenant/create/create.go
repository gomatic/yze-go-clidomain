package create // want "must declare a Config struct" "must declare the entry point Run"

// Result is declared, so this NESTED package sets out to be a verb — and the
// contract reaches it at any depth. Config and Run are missing. This is the
// shape the depth-1 predicate could never see.
type Result struct{ ID string }
