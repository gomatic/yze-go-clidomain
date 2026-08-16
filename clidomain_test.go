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
		"internal/domain/vocabalias", "internal/domain/grouped/setting",
		"internal/domain/other/borrowed")
}

// TestConfigBehaviourCountsEveryRoute pins that behaviour reaching Config is
// reported however it arrives — embedded, promised by an interface, or
// declared on a pointer receiver — and that a method declared only in an
// in-package test file is the tests' own vocabulary rather than the Config's
// behaviour. Each fixture takes one route and nothing else, so a rule reading
// only the declared one leaves three of the four wrong. Beside them sits the
// conforming near-miss: a Config HOLDING a behaviour-carrying type in a named
// field promotes nothing, which is the one-word remedy every reported route
// has available.
func TestConfigBehaviourCountsEveryRoute(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer,
		"internal/domain/embedded", "internal/domain/iface",
		"internal/domain/ptrmethod", "internal/domain/testmethods", "internal/domain/heldfield")
}

// TestRunSignatureConjunctsAreIndividuallyPinned pins every element of the Run
// contract separately: each fixture breaks exactly ONE and is exact in all the
// others, so deleting any single check changes what one of them reports. The
// two wrong-signature fixtures the suite began with each broke several at
// once, which is why none of them was pinned.
func TestRunSignatureConjunctsAreIndividuallyPinned(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer,
		"internal/domain/signonvariadic", "internal/domain/sigextraparam",
		"internal/domain/sigextraresult", "internal/domain/signoresult",
		"internal/domain/sigctx", "internal/domain/siglogger",
		"internal/domain/sigconfig", "internal/domain/sigresult", "internal/domain/sigerror",
		"internal/domain/generic")
}

// TestAContractLookupThatFindsNothingIsNotADereference pins the shape that
// declares Run and Result but no Config: the identity check looks a
// declaration up that is not there, and reporting the absence must not depend
// on the declaration existing.
func TestAContractLookupThatFindsNothingIsNotADereference(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer, "internal/domain/noconfig")
}

// TestTheReportAnchorsInAProductionFile pins where a missing-contract report
// lands for a verb spread over more than one file. The runner drops every
// finding a _test.go file holds, so an anchor that wanders into the test file
// does not move the finding — it deletes it, and every single-file fixture is
// blind to that.
func TestTheReportAnchorsInAProductionFile(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer, "internal/domain/multi")
}

// TestTheVerbSpeaksOnlyItsOwnAncestryOfGroups pins two of the three ways the
// vocabulary exemption is wider or narrower than its stated reason: a sibling
// group whose path is a mere string prefix is not an ancestor, and an ancestor
// that declares the contract is a verb rather than a group. The third — an
// ancestor ABOVE the tier — takes a unit test rather than a fixture, for the
// reason recorded at TestSpeaksAdmitsOnlyGroupsBeneathTheRoot.
func TestTheVerbSpeaksOnlyItsOwnAncestryOfGroups(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer,
		"internal/domain/tenantx/create", "internal/domain/verbgroup/sub", "internal/domain/verbgroup")
}

// TestAPackageIsJudgedByWhatItDeclaresNotByItsName pins that a package clause
// ending in _test does not take a package out of the tier. It is a thing an
// author writes, so a rule keyed on it is a rule with a five-character off
// switch and no record anywhere; what keeps the driver's synthesized passes
// unjudged is that their declarations all sit in _test.go files.
//
// Its sibling — an import path ending in .test, from a directory named that
// way — is reproduced the same way and CANNOT be a fixture here: the shared
// .gitignore ignores *.test, which matches the directory, so the case would be
// written, pass locally, and never reach the repository. The reproduction is
// recorded on clidomain.package-identity-is-not-a-name instead.
func TestAPackageIsJudgedByWhatItDeclaresNotByItsName(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), clidomain.Analyzer,
		"internal/domain/clausetest")
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
