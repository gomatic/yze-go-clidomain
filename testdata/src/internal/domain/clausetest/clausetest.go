// Package clausetest_test is ORDINARY production source whose package clause
// merely ends in _test. The go tool compiles and links it like any other
// package, so a rule keyed on the clause is a rule an author turns off with
// five characters. Result is declared and the rest of the contract is not.
package clausetest_test // want "must declare a Config struct" "must declare the entry point Run"

// Result is the outcome, so this package sets out to be a verb.
type Result struct{ Out string }
