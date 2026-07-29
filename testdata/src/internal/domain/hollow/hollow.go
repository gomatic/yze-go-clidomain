package hollow // want "must declare a Config struct" "must declare the entry point Run"

// Result is declared, so this package sets out to be a verb — and must then be
// one completely. Config and Run are missing.
type Result struct{ Out string }

// Helper is all else this package offers.
func Helper() string { return "" }
