package clidomain

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
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
}

// TestVariadicSpellingReadsTheSourceName pins that the check reads how the
// variadic was WRITTEN, not what it resolves to: a shared qualified alias, a
// package-local one, a bare type, a non-variadic signature, and a missing
// declaration are each distinguished.
func TestVariadicSpellingReadsTheSourceName(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	qualified := &ast.SelectorExpr{X: &ast.Ident{Name: "domain"}, Sel: &ast.Ident{Name: "Argument"}}
	want.Equal(sharedArgument, variadicSpelling(declWith(&ast.Ellipsis{Elt: qualified})))

	other := &ast.SelectorExpr{X: &ast.Ident{Name: "other"}, Sel: &ast.Ident{Name: "Argument"}}
	want.Equal(qualifiedName("other.Argument"), variadicSpelling(declWith(&ast.Ellipsis{Elt: other})))

	want.Empty(variadicSpelling(nil), "no declaration spells nothing")
	want.Empty(variadicSpelling(declWith(&ast.Ident{Name: "argument"})), "a non-variadic parameter")
	want.Empty(variadicSpelling(&ast.FuncDecl{Type: &ast.FuncType{Params: &ast.FieldList{}}}), "no parameters")

	local := &ast.Ellipsis{Elt: &ast.Ident{Name: "argument"}}
	want.Empty(variadicSpelling(declWith(local)), "a package-local alias is not a qualified name")

	notAnIdent := &ast.Ellipsis{Elt: &ast.SelectorExpr{X: &ast.CallExpr{}, Sel: &ast.Ident{Name: "Argument"}}}
	want.Empty(variadicSpelling(declWith(notAnIdent)), "a qualifier that is not a plain identifier")
}

// declWith builds a one-parameter function declaration carrying the given type.
func declWith(paramType ast.Expr) *ast.FuncDecl {
	return &ast.FuncDecl{Type: &ast.FuncType{
		Params: &ast.FieldList{List: []*ast.Field{{Type: paramType}}},
	}}
}

// TestDeclAtFindsTheDeclarationOrNothing pins the lookup: the declaration at a
// position is returned, and an unknown position yields nothing rather than a
// nil dereference downstream.
func TestDeclAtFindsTheDeclarationOrNothing(t *testing.T) {
	t.Parallel()

	name := &ast.Ident{Name: "Run", NamePos: token.Pos(42)}
	decl := &ast.FuncDecl{Name: name, Type: &ast.FuncType{Params: &ast.FieldList{}}}
	files := []*ast.File{{Decls: []ast.Decl{&ast.GenDecl{}, decl}}}

	assert.Same(t, decl, declAt(files, token.Pos(42)))
	assert.Nil(t, declAt(files, token.Pos(99)), "an unknown position finds nothing")
	assert.Nil(t, declAt(nil, token.Pos(42)), "no files find nothing")
}

// TestHasMethodsFindsBehaviourOnConfig pins the behaviour check: a type with a
// method has behaviour, and a type with none — or one that is not a named type
// at all — has none.
func TestHasMethodsFindsBehaviourOnConfig(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	want.False(hasMethods(types.Typ[types.String]), "a basic type declares no method")

	bare := types.NewNamed(types.NewTypeName(0, nil, "Config", nil), types.NewStruct(nil, nil), nil)
	want.False(hasMethods(bare), "a Config with no methods carries no behaviour")

	sig := types.NewSignatureType(types.NewVar(0, nil, "c", bare), nil, nil, nil, nil, false)
	bare.AddMethod(types.NewFunc(token.Pos(7), nil, "Validate", sig))
	want.True(hasMethods(bare), "a declared method is behaviour")
}
