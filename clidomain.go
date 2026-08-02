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
// a type named Run is the same violation. The variadic must use the SHARED
// domain.Argument alias rather than a locally-redeclared one: both spellings
// compile, so nothing but a rule keeps them from diverging.
//
// Scope is every package beneath internal/domain/, at any depth — a nested
// verb like internal/domain/tenant/create is as much a verb as a top-level
// one. The shared internal/domain vocabulary package itself declares no
// command, so it is skipped, and so is any grouping or helper package that
// declares no element of the contract. Only PRODUCTION declarations count:
// test files may freely declare Config-shaped test tables or Run-shaped
// drivers without making their package a verb, and the driver-synthesized
// external-test and test-main passes are never judged.
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
	messageBehaviour  = "Config carries no behaviour: it holds the bound flags and is read by Run, so it declares no methods"
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
	if isScaffoldingPackage(pass) || !isDomainCommand(packagePath(pass.Pkg.Path())) ||
		len(pass.Files) == 0 || !declaresContract(pass) {
		return nil, nil
	}
	at := pass.Files[0].Package
	reportMissing(pass, at)
	reportRun(pass)
	reportConfigMethods(pass)
	return nil, nil
}

// isScaffoldingPackage reports whether pass is a driver-synthesized test
// package rather than a real package: an external test package (clause
// "<pkg>_test") or the test-main package (import path "<pkg>.test"). Neither
// carries domain source, and judging one imposes the contract on a package
// that exists only in the driver.
func isScaffoldingPackage(pass *analysis.Pass) bool {
	return strings.HasSuffix(pass.Pkg.Name(), "_test") || strings.HasSuffix(pass.Pkg.Path(), ".test")
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

// lookup resolves a package-scope name to its object, but only when it is
// declared in a PRODUCTION file. The in-package test-variant pass folds
// _test.go declarations into the same scope, where a Config-shaped test table
// or a Run-shaped test driver would otherwise turn a grouping package into a
// verb — imposing the contract on vocabulary the tests own.
func lookup(pass *analysis.Pass, name declName) types.Object {
	obj := pass.Pkg.Scope().Lookup(string(name))
	if obj == nil || isTestFileDecl(pass, obj) {
		return nil
	}
	return obj
}

// isTestFileDecl reports whether obj is declared in a _test.go file.
func isTestFileDecl(pass *analysis.Pass, obj types.Object) bool {
	file := pass.Fset.File(obj.Pos())
	return file == nil || strings.HasSuffix(file.Name(), "_test.go")
}

// declaresContract reports whether the package declares ANY element of the
// domain contract in its production files, which is what marks it a verb
// rather than a grouping package.
//
// A verb's siblings may be grouped under an intermediate package holding logic
// they share — internal/domain/config exists so get, list and set can reuse its
// argument parsing — and such a package declares no command of its own.
// Requiring the contract of it would demand a Run with nothing to run. Keying
// on "declares part of the contract" makes the rule what it should be: a
// package that sets out to be a verb must be one completely.
func declaresContract(pass *analysis.Pass) bool {
	for _, want := range contract {
		if lookup(pass, want.name) != nil {
			return true
		}
	}
	return false
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
// contract in arity, variadicity, and types. Config and Result are matched by
// IDENTITY against this package's declarations — a Run built on another
// package's types leaves the declared ones dead — and context.Context and
// *slog.Logger by their declaring package's import path, so a look-alike
// package named "context" does not satisfy the contract.
func hasContractShape(pass *analysis.Pass, sig *types.Signature) bool {
	params, results := sig.Params(), sig.Results()
	if !sig.Variadic() || params.Len() != 4 || results.Len() != 2 {
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

// reportConfigMethods reports a Config that declares behaviour. Config is the
// bound-flag record the CLI tier writes and Run reads; a method on it moves
// logic into the tier that exists to hold none. The report anchors at THIS
// package's Config declaration — a Config aliased to another package's type
// still resolves to that type's methods, but the defect (and the fix) is the
// alias declared here.
func reportConfigMethods(pass *analysis.Pass) {
	obj := lookup(pass, "Config")
	if obj == nil {
		return
	}
	if hasMethods(obj.Type()) {
		pass.Reportf(obj.Pos(), "%s", messageBehaviour)
	}
}

// hasMethods reports whether t resolves to a named type declaring methods.
func hasMethods(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	return ok && named.NumMethods() > 0
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
