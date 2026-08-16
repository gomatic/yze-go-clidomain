package testmethods

import "testing"

// withGreeting is the ordinary table-test builder a Config-shaped fixture
// needs. Deleting it is the only remedy a behaviour finding here could
// prescribe, which is what makes such a finding a manufactured disablement.
func (c Config) withGreeting(g string) Config { c.Greeting = g; return c }

func TestRunGreets(t *testing.T) {
	if (Config{}).withGreeting("hi").Greeting != "hi" {
		t.Fatal("the builder is the point of this file")
	}
}
