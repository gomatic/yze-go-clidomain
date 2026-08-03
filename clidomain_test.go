package clidomain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	clidomain "github.com/gomatic/yze-go-clidomain"
)

// TestDomainContract pins the whole contract against the fixtures: a conformant
// package is silent, a drifted one is reported for its package-local argument
// alias and for behaviour on Config, and a package missing the contract types
// is reported for each one it lacks.
func TestDomainContract(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer,
		"internal/domain/greet", "internal/domain/drifted",
		"internal/domain/hollow", "internal/domain/malformed", "internal/domain/bare", "internal/domain/grouped",
		"internal/domain/tenant", "internal/domain/tenant/create", "internal/domain/tenant/list",
		"internal/domain/shape", "internal/domain/varrun", "internal/domain/impostor",
		"internal/domain/exttest", "internal/domain/grouphelp", "internal/domain/aliasmethods",
		"internal/domain/vocabalias")
}

// TestTheSharedVocabularyPackageIsNotACommand pins that internal/domain itself
// declares no command and is therefore out of scope — it holds the vocabulary
// the per-verb packages share.
func TestTheSharedVocabularyPackageIsNotACommand(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer, "internal/domain")
}

// TestRegistrationIsWellFormed pins the yze wiring.
func TestRegistrationIsWellFormed(t *testing.T) {
	t.Parallel()

	assert.NoError(t, clidomain.Registration.Validate())
	assert.Equal(t, "yze/clidomain", clidomain.Registration.RuleID())
	assert.Same(t, clidomain.Analyzer, clidomain.Registration.Analyzer)
}
