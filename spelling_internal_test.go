package clidomain

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVariadicSpellingReadsTheSourceName pins that the check reads how the
// variadic was WRITTEN, not what it resolves to: a shared qualified alias, a
// package-local one, a bare type, a non-variadic signature, and a missing
// declaration are each distinguished.
func TestVariadicSpellingReadsTheSourceName(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	qualified := &ast.SelectorExpr{X: &ast.Ident{Name: "domain"}, Sel: &ast.Ident{Name: "Argument"}}
	want.Equal(sharedArgument, spellingOf(variadicSelector(declWith(&ast.Ellipsis{Elt: qualified}))))

	other := &ast.SelectorExpr{X: &ast.Ident{Name: "other"}, Sel: &ast.Ident{Name: "Argument"}}
	want.Equal(qualifiedName("other.Argument"), spellingOf(variadicSelector(declWith(&ast.Ellipsis{Elt: other}))))

	want.Empty(spellingOf(variadicSelector(nil)), "no declaration spells nothing")
	want.Empty(spellingOf(variadicSelector(declWith(&ast.Ident{Name: "argument"}))), "a non-variadic parameter")
	want.Empty(
		spellingOf(variadicSelector(&ast.FuncDecl{Type: &ast.FuncType{Params: &ast.FieldList{}}})),
		"no parameters",
	)

	local := &ast.Ellipsis{Elt: &ast.Ident{Name: "argument"}}
	want.Empty(spellingOf(variadicSelector(declWith(local))), "a package-local alias is not a qualified name")

	notAnIdent := &ast.Ellipsis{Elt: &ast.SelectorExpr{X: &ast.CallExpr{}, Sel: &ast.Ident{Name: "Argument"}}}
	want.Nil(variadicSelector(declWith(notAnIdent)), "a qualifier that is not a plain identifier")
}

// TestSharedVocabularyIsTheDomainRoot pins the derivation of the shared
// vocabulary package's path from a domain command's: the internal/domain root
// the command sits beneath, with or without a module prefix.
func TestSharedVocabularyIsTheDomainRoot(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	want.Equal(pkgPath("internal/domain"), sharedVocabulary("internal/domain/greet"))
	want.Equal(pkgPath("github.com/o/r/internal/domain"),
		sharedVocabulary("github.com/o/r/internal/domain/tenant/create"))
}

// TestImportedByResolvesOnlyPackageNames pins the qualifier resolution: an
// identifier bound to an imported package yields that package's path, and one
// bound to anything else — or to nothing — yields empty, so the impostor
// check fails open rather than inventing a finding.
func TestImportedByResolvesOnlyPackageNames(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	ident := &ast.Ident{Name: "domain"}
	vocab := types.NewPackage("example.com/mod/internal/vocab", "vocab")
	info := &types.Info{Uses: map[*ast.Ident]types.Object{
		ident: types.NewPkgName(0, nil, "domain", vocab),
	}}

	want.Equal("example.com/mod/internal/vocab", importedBy(info, ident).Path())
	want.Nil(importedBy(&types.Info{}, ident), "an unresolved qualifier is not a package")
}

// TestImpostorPathNamesTheAliasedPackage pins the impostor check: a qualifier
// aliasing a package other than the shared vocabulary is named, the genuine
// vocabulary package is not, and an unresolved qualifier fails open.
func TestImpostorPathNamesTheAliasedPackage(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	ident := &ast.Ident{Name: "domain"}
	sel := &ast.SelectorExpr{X: ident, Sel: &ast.Ident{Name: "Argument"}}
	bound := func(path string) *types.Info {
		pkg := types.NewPackage(path, "vocab")
		return &types.Info{Uses: map[*ast.Ident]types.Object{ident: types.NewPkgName(0, nil, "domain", pkg)}}
	}

	want.Equal(pkgPath("example.com/mod/internal/vocab"),
		impostorPath(bound("example.com/mod/internal/vocab"), sel, "example.com/mod/internal/domain/greet"))
	want.Empty(impostorPath(bound("example.com/mod/internal/domain"), sel, "example.com/mod/internal/domain/greet"),
		"the genuine vocabulary package is no impostor")
	want.Empty(impostorPath(&types.Info{}, sel, "example.com/mod/internal/domain/greet"),
		"an unresolved qualifier fails open")
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

// TestSpeaksAdmitsOnlyGroupsBeneathTheRoot pins each conjunct of the ancestry
// rule separately: the root itself speaks, a group between the root and the
// verb speaks, a sibling group whose path is a string prefix without being a
// path prefix does not, an ancestor that declares the contract is a verb and
// does not, and an ancestor ABOVE the tier does not.
//
// That last one is here rather than in a fixture because it cannot be one: the
// only ancestor of internal/domain/<verb> that is not the root is the import
// path "internal", and in the GOPATH tree analysistest builds, GOROOT's own
// src/internal shadows it — the loader reports "no Go files in
// .../libexec/src/internal" and the case never reaches the analyzer. A
// constructed package has no loader to disagree with.
func TestSpeaksAdmitsOnlyGroupsBeneathTheRoot(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	const verb packagePath = "m/internal/domain/tenant/create"
	vocabulary := func(path string, decls ...string) *types.Package {
		pkg := types.NewPackage(path, "vocab")
		pkg.Scope().Insert(types.NewTypeName(0, pkg, "Argument", types.Typ[types.String]))
		for _, decl := range decls {
			pkg.Scope().Insert(types.NewTypeName(0, pkg, decl, types.NewStruct(nil, nil)))
		}
		return pkg
	}

	want.True(speaks(verb, vocabulary("m/internal/domain")), "the root is the shared vocabulary")
	want.True(speaks(verb, vocabulary("m/internal/domain/tenant")), "a group between root and verb extends it")
	want.False(speaks(verb, vocabulary("m/internal")), "an ancestor above the tier is not vocabulary")
	want.False(speaks(verb, vocabulary("m/internal/domain/ten")),
		"a string prefix of the verb's path is not an ancestor of it")
	want.False(speaks(verb, vocabulary("m/internal/domain/tenant", "Result")),
		"an ancestor declaring the contract is a verb, and a verb's types are its own")
	want.False(speaks(verb, vocabulary("m/internal/other/tenant")), "another tree's package is an impostor")
}

// TestDeclaresCommandAsksTheImportedPackage pins what separates a group
// package from a verb when the package is only reachable through its exported
// scope: declaring any element of the contract makes it a command, and a
// package of helpers is not one.
func TestDeclaresCommandAsksTheImportedPackage(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	group := types.NewPackage("m/internal/domain/tenant", "tenant")
	group.Scope().Insert(types.NewTypeName(0, group, "Argument", types.Typ[types.String]))
	want.False(declaresCommand(group), "a vocabulary alias is not a command")

	verb := types.NewPackage("m/internal/domain/verbgroup", "verbgroup")
	verb.Scope().Insert(types.NewTypeName(0, verb, "Result", types.NewStruct(nil, nil)))
	want.True(declaresCommand(verb), "declaring Result is setting out to be a verb")
}
