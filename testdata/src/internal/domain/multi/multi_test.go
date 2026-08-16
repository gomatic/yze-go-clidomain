package multi

import "testing"

func TestHelper(t *testing.T) {
	if Helper() != "" {
		t.Fatal("the point of this file is that it sorts last")
	}
}
