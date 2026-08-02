package grouphelp

import "testing"

// Config is an ordinary test-table type, not the domain contract.
type Config struct{ In, Want string }

func TestNormalize(t *testing.T) {
	for _, tc := range []Config{{In: "a", Want: "a"}} {
		if got := Normalize(tc.In); got != tc.Want {
			t.Fatalf("Normalize(%q) = %q, want %q", tc.In, got, tc.Want)
		}
	}
}
