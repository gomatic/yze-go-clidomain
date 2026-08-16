// Package clidomain provides a go/analysis analyzer enforcing the domain tier
// of the opinionated three-tier CLI layout.
//
// The layout splits a CLI into three tiers. The APP tier
// (internal/app/commands/<verb>) owns flags and help text; the DOMAIN tier
// (internal/domain/<verb>) owns orchestration; the implementation tier
// (internal/<capability>) owns reusable logic. The other parts are policed
// elsewhere — stickler/clilayout checks that command and domain packages
// correspond across the whole module, and yze/cliapp checks the command
// package's own shape — so this one covers the tier neither reaches: the
// domain package's contract.
//
// That contract is exact, which is what makes it checkable. Every domain
// package declares a Config carrying the flags the CLI binds and NO behaviour, a
// Result, and a single entry point:
//
//	func Run(context.Context, *slog.Logger, Config, ...domain.Argument) (Result, error)
//
// Config and Result mean THIS package's declarations: Run's third parameter and
// first result are matched by type identity against them, so a Run wired to
// another package's types — leaving the declared ones dead — is a signature
// violation, not a pass. Run itself must be a function; a package-level var or
// a type named Run is the same violation. Nor may it be GENERIC: a type
// parameter cannot be inferred from the call the CLI tier writes, so an entry
// point carrying one is an entry point nothing can call — the same violation
// again, wearing the right signature. The variadic must use the SHARED
// domain.Argument alias rather than a locally-redeclared one: both spellings
// compile, so nothing but a rule keeps them from diverging. Spelling alone is
// not enough — the domain qualifier must also RESOLVE to the repo's shared
// internal/domain vocabulary package, since importing any other package as
// domain reproduces the spelling while defeating the rule.
//
// Config carries no behaviour BY ANY ROUTE. A method declared on it, one
// promoted from an embedded field, and one promised by an interface are the
// same defect, because logic lands in the tier that exists to hold none
// however it arrives — and a rule that reads only the declared route is turned
// off by embedding the type instead of aliasing it, which is one character and
// no record anywhere.
//
// Scope is every package beneath internal/domain/, at any depth — a nested
// verb like internal/domain/tenant/create is as much a verb as a top-level
// one. The shared internal/domain vocabulary package itself declares no
// command, so it is skipped, and so is any grouping or helper package that
// declares no element of the contract. Only PRODUCTION declarations count —
// for the contract's types AND for Config's methods: test files may freely
// declare Config-shaped test tables, Run-shaped drivers and builder methods on
// Config without imposing the contract or making Config behave. A package
// whose files are all _test.go therefore declares nothing this rule reads,
// which is how the driver-synthesized external-test and test-main passes go
// unjudged — no package NAME and no import path is consulted, since a name
// ending in _test or a path ending in .test is a thing an author writes.
//
// A missing-contract report anchors at the package clause of the verb's FIRST
// PRODUCTION file, because a finding a _test.go file holds is dropped rather
// than relocated: this analyzer's Registration below declares no TestScope,
// and the scope is instead applied by the runner's catalog, which lists
// clidomain in its sourceOnly set (gomatic/yze registrations.go:76, verified
// 2026-08-15). That is a claim about ANOTHER repository, so it is cited rather
// than asserted — nothing here fails if it changes, and a reader can check it.
// Anchoring in a test file would therefore not move the finding, it would
// delete it.
package clidomain

import (
	"go/token"
	"go/types"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
)

// Diagnostic messages, one per element of the domain contract.
const (
	messageConfig     = "domain package must declare a Config struct holding the flags the CLI tier binds"
	messageResult     = "domain package must declare a Result type: the outcome its Run returns"
	messageRun        = "domain package must declare the entry point Run(context.Context, *slog.Logger, Config, ...domain.Argument) (Result, error)"
	messageSignature  = "Run must be a function taking (context.Context, *slog.Logger, Config, ...domain.Argument) and returning (Result, error), using THIS package's Config and Result"
	messageLocalAlias = "Run's variadic must use the shared domain.Argument alias, not a package-local redeclaration; one concept spelled per-package is how the tiers drift"
	messageImpostor   = "Run's variadic is spelled domain.Argument, but domain here is an import alias for %q, not the shared %q vocabulary package the contract names"
	messageBehaviour  = "Config must carry no behaviour: it holds the bound flags and is read by Run, so no method may reach it — declared on it, promoted from an embedded field, or promised by an interface"
)

// domainDir is the path segment marking the domain tier.
const domainDir = "internal/domain"

// qualifiedName is a type written as pkg.Name in source.
type qualifiedName string

// sharedArgument is the qualified name every domain Run's variadic must use.
const sharedArgument qualifiedName = "domain.Argument"

// Analyzer reports domain packages that break the three-tier CLI contract.
var Analyzer = &analysis.Analyzer{
	Name: "clidomain",
	Doc:  "reports domain packages that break the opinionated three-tier CLI contract (Config, Result, Run)",
	Run:  run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "clidomain",
	Categories: []goyze.Category{"cli", "structure"},
	URL:        "https://docs.gomatic.dev/yze/clidomain",
	Analyzer:   Analyzer,
}

// packagePath is the import path of the package under analysis.
type packagePath string

// declName is a declared identifier being looked for.
type declName string

// typeName is a type's identifier, and pkgPath the import path of the package
// declaring it.
type (
	typeName string
	pkgPath  string
)

// requirement is one element of the domain contract: the identifier a domain
// package must declare, and what to say when it does not.
type requirement struct {
	name    declName
	message string
}

// contract is every declaration a domain package must provide.
var contract = []requirement{
	{name: "Config", message: messageConfig},
	{name: "Result", message: messageResult},
	{name: "Run", message: messageRun},
}

// run reports each element of the domain contract the package fails to meet.
func run(pass *analysis.Pass) (any, error) {
	if !isDomainCommand(packagePath(pass.Pkg.Path())) || !declaresContract(pass) {
		return nil, nil
	}
	reportMissing(pass, productionAnchor(pass))
	reportRun(pass)
	reportConfigMethods(pass)
	return nil, nil
}

// isDomainCommand reports whether the path names a package beneath the domain
// tier — internal/domain/<path>, at ANY depth, so a nested verb like
// internal/domain/tenant/create is in scope — and not the shared
// internal/domain vocabulary package, which declares no command of its own.
// Whether a package beneath the tier is a verb (checked) or a grouping/helper
// package (skipped) is decided by declaresContract, not by depth.
func isDomainCommand(pkg packagePath) bool {
	before, rest, found := strings.Cut(string(pkg), domainDir+"/")
	return found && rest != "" && (before == "" || strings.HasSuffix(before, "/"))
}

// reportMissing reports each contract element the package fails to declare in
// its production files.
func reportMissing(pass *analysis.Pass, at token.Pos) {
	for _, want := range contract {
		if lookup(pass, want.name) == nil {
			pass.Reportf(at, "%s", want.message)
		}
	}
}

// reportRun checks the declared Run against the contract. A Run that is not a
// function at all — a package-level var, or a type named Run — is the same
// violation as a wrong signature: the entry point the contract names does not
// exist.
func reportRun(pass *analysis.Pass) {
	obj := lookup(pass, "Run")
	if obj == nil {
		return
	}
	fn, ok := obj.(*types.Func)
	if !ok {
		pass.Reportf(obj.Pos(), "%s", messageSignature)
		return
	}
	sig, _ := fn.Type().(*types.Signature)
	if sig == nil || !hasContractShape(pass, sig) {
		pass.Reportf(fn.Pos(), "%s", messageSignature)
		return
	}
	reportVariadic(pass, fn, sig)
}

// hasContractShape reports whether the signature matches the domain Run
// contract in arity, variadicity, type parameters and types. A GENERIC Run is
// rejected here: the CLI tier's call cannot infer the type argument, so the
// entry point is uncallable however well its parameters read. Config and
// Result are matched by
// IDENTITY against this package's declarations — a Run built on another
// package's types leaves the declared ones dead — and context.Context and
// *slog.Logger by their declaring package's import path, so a look-alike
// package named "context" does not satisfy the contract.
func hasContractShape(pass *analysis.Pass, sig *types.Signature) bool {
	params, results := sig.Params(), sig.Results()
	if !sig.Variadic() || sig.TypeParams().Len() != 0 || params.Len() != 4 || results.Len() != 2 {
		return false
	}
	return isPackageType(params.At(0).Type(), "context", "Context") &&
		isPointerTo(params.At(1).Type(), "log/slog", "Logger") &&
		isDeclaredHere(pass, params.At(2).Type(), "Config") &&
		isDeclaredHere(pass, results.At(0).Type(), "Result") &&
		types.Identical(results.At(1).Type(), types.Universe.Lookup("error").Type())
}

// isDeclaredHere reports whether t is identical to the type this package
// declares under name.
func isDeclaredHere(pass *analysis.Pass, t types.Type, name declName) bool {
	obj := lookup(pass, name)
	return obj != nil && types.Identical(t, obj.Type())
}

// reportConfigMethods reports a Config that carries behaviour. Config is the
// bound-flag record the CLI tier writes and Run reads; a method on it moves
// logic into the tier that exists to hold none. The report anchors at THIS
// package's Config declaration — a Config aliased to, embedding, or promising
// another package's methods still resolves to them, but the defect (and the
// fix) is the declaration made here.
func reportConfigMethods(pass *analysis.Pass) {
	obj := lookup(pass, "Config")
	if obj != nil && hasBehaviour(pass, obj.Type()) {
		pass.Reportf(obj.Pos(), "%s", messageBehaviour)
	}
}

// hasBehaviour reports whether any method reaches t by any route — declared on
// it, promoted from an embedded field, or promised by an interface — since a
// rule reading only the declared route is turned off by embedding the type
// instead of aliasing it. The method SET is what answers that, and it is taken
// through a pointer as well, because a pointer-receiver method is behaviour
// the value's own set omits.
func hasBehaviour(pass *analysis.Pass, t types.Type) bool {
	resolved := types.Unalias(t)
	return declaredInProduction(pass, types.NewMethodSet(resolved)) ||
		declaredInProduction(pass, types.NewMethodSet(types.NewPointer(resolved)))
}

// isPackageType reports whether t is the named type declared as path.name —
// matched by the declaring package's import PATH, never its name, so a
// look-alike local package cannot spoof a standard-library type.
func isPackageType(t types.Type, path pkgPath, name typeName) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok || named.Obj().Name() != string(name) {
		return false
	}
	return named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == string(path)
}

// isPointerTo reports whether t is a pointer to the named type path.name.
func isPointerTo(t types.Type, path pkgPath, name typeName) bool {
	ptr, ok := types.Unalias(t).(*types.Pointer)
	return ok && isPackageType(ptr.Elem(), path, name)
}
