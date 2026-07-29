// Package cliopinion provides a go/analysis analyzer enforcing the domain tier
// of the opinionated three-tier CLI layout.
//
// The layout splits a CLI into three tiers. The APP tier
// (internal/app/commands/<verb>) owns flags and help text; the DOMAIN tier
// (internal/domain/<verb>) owns orchestration; the implementation tier
// (internal/<capability>) owns reusable logic. Two analyzers already police the
// other parts — yze/layout checks that a command package and its domain package
// correspond, and yze/pkgstd checks the command package's own shape — so this
// one covers the tier neither reaches: the domain package's contract.
//
// That contract is exact, which is what makes it checkable. Every domain
// package declares a Config carrying the flags the CLI binds and NO behaviour, a
// Result, and a single entry point:
//
//	func Run(context.Context, *slog.Logger, Config, ...domain.Argument) (Result, error)
//
// The variadic must use the SHARED domain.Argument alias rather than a
// locally-redeclared one. Both spellings compile and both satisfy the runner
// contract, so nothing but a rule keeps them from diverging — and they already
// have: the reference layout uses domain.Argument while other repositories
// redeclare `type argument = string` per package, which is the same concept
// named differently in every package that uses it.
//
// Scope is packages under internal/domain/<verb> only. The shared
// internal/domain vocabulary package itself declares no command, so it is
// skipped.
package cliopinion

import (
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
)

// Diagnostic messages, one per element of the domain contract.
const (
	messageConfig     = "domain package must declare a Config struct holding the flags the CLI tier binds"
	messageResult     = "domain package must declare a Result type: the outcome its Run returns"
	messageRun        = "domain package must declare the entry point Run(context.Context, *slog.Logger, Config, ...domain.Argument) (Result, error)"
	messageSignature  = "Run must take (context.Context, *slog.Logger, Config, ...domain.Argument) and return (Result, error)"
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
	Name: "cliopinion",
	Doc:  "reports domain packages that break the opinionated three-tier CLI contract (Config, Result, Run)",
	Run:  run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "cliopinion",
	Categories: []goyze.Category{"cli", "structure"},
	URL:        "https://docs.gomatic.dev/yze/cliopinion",
	Analyzer:   Analyzer,
}

// packagePath is the import path of the package under analysis.
type packagePath string

// declName is a declared identifier being looked for.
type declName string

// typeName is a type's identifier, and pkgName the package qualifying it; an
// empty pkgName means the package under analysis.
type (
	typeName string
	pkgName  string
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
	if !isDomainCommand(packagePath(pass.Pkg.Path())) || len(pass.Files) == 0 || !declaresContract(pass) {
		return nil, nil
	}
	at := pass.Files[0].Package
	reportMissing(pass, at)
	reportRun(pass)
	reportConfigMethods(pass)
	return nil, nil
}

// isDomainCommand reports whether the path names a per-verb domain package —
// internal/domain/<verb> — and not the shared internal/domain vocabulary
// package, which declares no command of its own. The synthesized `.test` main
// package is excluded: it is generated, not written.
func isDomainCommand(pkg packagePath) bool {
	trimmed := strings.TrimSuffix(strings.TrimSuffix(string(pkg), "_test"), ".test")
	dir, verb := path.Split(trimmed)
	return strings.HasSuffix(strings.TrimSuffix(dir, "/"), domainDir) && verb != ""
}

// declaresContract reports whether the package declares ANY element of the
// domain contract, which is what marks it a verb rather than a grouping package.
//
// A verb's siblings may be grouped under an intermediate package holding logic
// they share — internal/domain/config exists so get, list and set can reuse its
// argument parsing — and such a package declares no command of its own.
// Requiring the contract of it would demand a Run with nothing to run. Keying
// on "declares part of the contract" makes the rule what it should be: a
// package that sets out to be a verb must be one completely.
func declaresContract(pass *analysis.Pass) bool {
	for _, want := range contract {
		if pass.Pkg.Scope().Lookup(string(want.name)) != nil {
			return true
		}
	}
	return false
}

// reportMissing reports each contract type the package fails to declare.
func reportMissing(pass *analysis.Pass, at token.Pos) {
	for _, want := range contract {
		if pass.Pkg.Scope().Lookup(string(want.name)) == nil {
			pass.Reportf(at, "%s", want.message)
		}
	}
}

// reportRun checks the declared Run's signature against the contract.
func reportRun(pass *analysis.Pass) {
	fn, ok := pass.Pkg.Scope().Lookup("Run").(*types.Func)
	if !ok {
		return
	}
	sig, _ := fn.Type().(*types.Signature)
	if sig == nil || !hasContractShape(sig) {
		pass.Reportf(fn.Pos(), "%s", messageSignature)
		return
	}
	reportVariadic(pass, fn, sig)
}

// hasContractShape reports whether the signature matches the domain Run
// contract in arity, variadicity, and result types.
func hasContractShape(sig *types.Signature) bool {
	params, results := sig.Params(), sig.Results()
	if !sig.Variadic() || params.Len() != 4 || results.Len() != 2 {
		return false
	}
	return isNamed(params.At(0).Type(), "context", "Context") &&
		isPointerTo(params.At(1).Type(), "slog", "Logger") &&
		isNamed(params.At(2).Type(), "", "Config") &&
		isNamed(results.At(0).Type(), "", "Result") &&
		types.Identical(results.At(1).Type(), types.Universe.Lookup("error").Type())
}

// reportVariadic reports a Run whose variadic is a package-local alias rather
// than the shared domain.Argument.
func reportVariadic(pass *analysis.Pass, fn *types.Func, sig *types.Signature) {
	if !usesSharedArgument(pass, fn) {
		pass.Reportf(sig.Params().At(3).Pos(), "%s", messageLocalAlias)
	}
}

// usesSharedArgument reports whether Run's variadic is written as the shared
// domain.Argument. Both it and a local alias resolve to string, so the check
// reads the SOURCE spelling rather than the resolved type — the distinction is
// the name, which is the whole point of standardising it.
func usesSharedArgument(pass *analysis.Pass, fn *types.Func) bool {
	return variadicSpelling(declAt(pass.Files, fn.Pos())) == sharedArgument
}

// declAt finds the function declared at pos, or nil when no file holds one.
func declAt(files []*ast.File, pos token.Pos) *ast.FuncDecl {
	for _, file := range files {
		for _, decl := range file.Decls {
			found, ok := decl.(*ast.FuncDecl)
			if ok && found.Name.Pos() == pos && found.Type.Params != nil {
				return found
			}
		}
	}
	return nil
}

// variadicSpelling is the SOURCE spelling of decl's variadic element type —
// "domain.Argument" for the shared alias — or empty when there is no
// declaration, no variadic, or the element is not a qualified name.
func variadicSpelling(decl *ast.FuncDecl) qualifiedName {
	if decl == nil || len(decl.Type.Params.List) == 0 {
		return ""
	}
	params := decl.Type.Params.List
	last, ok := params[len(params)-1].Type.(*ast.Ellipsis)
	if !ok {
		return ""
	}
	sel, ok := last.Elt.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return qualifiedName(ident.Name + "." + sel.Sel.Name)
}

// reportConfigMethods reports a Config that declares behaviour. Config is the
// bound-flag record the CLI tier writes and Run reads; a method on it moves
// logic into the tier that exists to hold none.
func reportConfigMethods(pass *analysis.Pass) {
	obj := pass.Pkg.Scope().Lookup("Config")
	if obj == nil {
		return
	}
	if at := firstMethod(obj.Type()); at.IsValid() {
		pass.Reportf(at, "%s", messageBehaviour)
	}
}

// firstMethod is the position of a type's first declared method, or NoPos when
// it declares none. A Config written as an alias to another package's type
// resolves to something unnamed here, which likewise declares no method OF
// Config.
func firstMethod(t types.Type) token.Pos {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok || named.NumMethods() == 0 {
		return token.NoPos
	}
	return named.Method(0).Pos()
}

// isNamed reports whether t is the named type pkg.name, with an empty pkg
// meaning the package under analysis.
func isNamed(t types.Type, pkg pkgName, name typeName) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok || named.Obj().Name() != string(name) {
		return false
	}
	return pkg == "" || (named.Obj().Pkg() != nil && named.Obj().Pkg().Name() == string(pkg))
}

// isPointerTo reports whether t is a pointer to the named type pkg.name.
func isPointerTo(t types.Type, pkg pkgName, name typeName) bool {
	ptr, ok := types.Unalias(t).(*types.Pointer)
	return ok && isNamed(ptr.Elem(), pkg, name)
}
