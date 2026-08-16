package clidomain

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

// TestIsPackageTypeMatchesByImportPath pins the type classifier's guards: a
// type that is not named at all, one whose name differs, and — the spoofing
// case — a type of the right NAME from a package that merely shares the
// standard library package's short name are none of them the contract's type.
// Only the declaring package's import PATH matches.
func TestIsPackageTypeMatchesByImportPath(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	spoof := types.NewPackage("example.com/context", "context")
	spoofed := types.NewNamed(types.NewTypeName(0, spoof, "Context", nil), types.NewStruct(nil, nil), nil)
	stdlib := types.NewPackage("context", "context")
	genuine := types.NewNamed(types.NewTypeName(0, stdlib, "Context", nil), types.NewStruct(nil, nil), nil)

	want.False(isPackageType(types.Typ[types.String], "context", "Context"), "a basic type is not the contract type")
	want.False(isPackageType(spoofed, "context", "Context"), "a look-alike package name does not match the path")
	want.False(isPackageType(genuine, "context", "Timer"), "the wrong name does not match")
	want.True(isPackageType(genuine, "context", "Context"), "the declaring path and name match")
}

// TestIsPointerToRequiresAPointer pins that a non-pointer is never a pointer to
// the contract's type.
func TestIsPointerToRequiresAPointer(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	pkg := types.NewPackage("log/slog", "slog")
	logger := types.NewNamed(types.NewTypeName(0, pkg, "Logger", nil), types.NewStruct(nil, nil), nil)

	want.False(isPointerTo(logger, "log/slog", "Logger"), "the value type is not the pointer type")
	want.True(isPointerTo(types.NewPointer(logger), "log/slog", "Logger"))
}

// TestIsDomainCommandBoundsTheScope pins which package paths are in scope: any
// package beneath the domain tier is — at any depth — while the shared
// vocabulary package, anything outside the tier, and look-alike segments are
// not.
func TestIsDomainCommandBoundsTheScope(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	want.True(isDomainCommand("github.com/o/r/internal/domain/greet"))
	want.True(isDomainCommand("internal/domain/greet"))
	want.True(isDomainCommand("github.com/o/r/internal/domain/tenant/create"), "a nested verb is in scope")
	want.True(isDomainCommand("internal/domain/tenant/create"), "a nested verb is in scope without a module prefix")
	want.False(isDomainCommand("internal/domain"), "the shared vocabulary package declares no command")
	want.False(isDomainCommand("github.com/o/r/internal/domain"), "the vocabulary package with a module prefix")
	want.False(isDomainCommand("internal/app/commands/greet"))
	want.False(isDomainCommand("internal/greeting"))
	want.False(isDomainCommand("myinternal/domain/greet"), "a segment merely ending in the tier name is not the tier")
	want.False(isDomainCommand("internal/domainx/greet"), "a segment merely STARTING with the tier name is not it")
	want.False(isDomainCommand("internal/domainx"), "nor is the look-alike segment itself")
}

// TestHasBehaviourFindsBehaviourOnConfig pins the behaviour check against a
// type built here rather than parsed: a type with a method has behaviour, and
// a type with none — or one that is not a named type at all — has none. The
// routes behaviour ARRIVES by (embedding, an interface, a pointer receiver)
// are pinned by the fixtures, which is where a promoted method is real.
func TestHasBehaviourFindsBehaviourOnConfig(t *testing.T) {
	t.Parallel()
	want := assert.New(t)
	pass := &analysis.Pass{Fset: token.NewFileSet()}

	want.False(hasBehaviour(pass, types.Typ[types.String]), "a basic type declares no method")

	bare := types.NewNamed(types.NewTypeName(0, nil, "Config", nil), types.NewStruct(nil, nil), nil)
	want.False(hasBehaviour(pass, bare), "a Config with no methods carries no behaviour")

	sig := types.NewSignatureType(types.NewVar(0, nil, "c", bare), nil, nil, nil, nil, false)
	bare.AddMethod(types.NewFunc(token.Pos(7), nil, "Validate", sig))
	want.True(hasBehaviour(pass, bare), "a declared method is behaviour")
}

// TestHasContractShapeRejectsAGenericRun pins the single difference a type
// parameter makes. The signature built here is the contract's, element for
// element; adding one type parameter and changing nothing else must reject it,
// because the type argument is inferable from no part of the call the CLI tier
// writes, so the entry point the contract names is one nothing can call. The
// generic fixture proves the same thing end to end; this proves that the type
// parameter, and not some other element of that fixture, is what does it.
func TestHasContractShapeRejectsAGenericRun(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	verb := types.NewPackage("m/internal/domain/greet", "greet")
	declare := func(name string) *types.Named {
		named := types.NewNamed(types.NewTypeName(0, verb, name, nil), types.NewStruct(nil, nil), nil)
		verb.Scope().Insert(named.Obj())
		return named
	}
	foreign := func(path, pkgName, name string) *types.Named {
		pkg := types.NewPackage(path, pkgName)
		return types.NewNamed(types.NewTypeName(0, pkg, name, nil), types.NewStruct(nil, nil), nil)
	}

	params := types.NewTuple(
		types.NewParam(0, verb, "ctx", foreign("context", "context", "Context")),
		types.NewParam(0, verb, "log", types.NewPointer(foreign("log/slog", "slog", "Logger"))),
		types.NewParam(0, verb, "cfg", declare("Config")),
		types.NewParam(0, verb, "args", types.NewSlice(types.Typ[types.String])),
	)
	results := types.NewTuple(
		types.NewParam(0, verb, "", declare("Result")),
		types.NewParam(0, verb, "", types.Universe.Lookup("error").Type()),
	)
	pass := &analysis.Pass{Fset: token.NewFileSet(), Pkg: verb}

	want.True(hasContractShape(pass, types.NewSignatureType(nil, nil, nil, params, results, true)),
		"the contract's own signature, element for element")

	parameterised := types.NewTypeParam(types.NewTypeName(0, verb, "T", nil), types.NewInterfaceType(nil, nil))
	generic := types.NewSignatureType(nil, nil, []*types.TypeParam{parameterised}, params, results, true)
	want.False(hasContractShape(pass, generic),
		"the same signature carrying one type parameter is an entry point nothing can call")
}
