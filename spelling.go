package clidomain

// The variadic-spelling rule: Run's variadic must be WRITTEN as the shared
// domain.Argument, and the qualifier must actually BE the shared vocabulary
// package. Both the shared alias and a package-local one resolve to string,
// so the first check reads the source spelling — the name is the whole point
// of standardising it. The spelling alone is spoofable, though: importing any
// package AS domain reproduces the spelling exactly, so the qualifier is then
// resolved and required to be the repo's internal/domain root.

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// reportVariadic reports a Run whose variadic breaks the shared-vocabulary
// rule: one spelled as a package-local alias, or one spelled domain.Argument
// where "domain" is an import alias for some other package — the same drift
// wearing the standard spelling.
func reportVariadic(pass *analysis.Pass, fn *types.Func, sig *types.Signature) {
	at := sig.Params().At(3).Pos()
	sel := variadicSelector(declAt(pass.Files, fn.Pos()))
	if spellingOf(sel) != sharedArgument {
		pass.Reportf(at, "%s", messageLocalAlias)
		return
	}
	if impostor := impostorPath(pass.TypesInfo, sel, packagePath(pass.Pkg.Path())); impostor != "" {
		pass.Reportf(at, messageImpostor, impostor, sharedVocabulary(packagePath(pass.Pkg.Path())))
	}
}

// impostorPath is the import path the variadic qualifier resolves to when it
// is NOT a vocabulary package the verb may speak — empty when it is genuine. A
// qualifier that does not resolve to a package at all also yields empty: the
// type checker guarantees a qualified type's qualifier is a package in
// compiled code, so an unresolved one fails open rather than inventing a
// finding.
func impostorPath(info *types.Info, sel *ast.SelectorExpr, pkg packagePath) pkgPath {
	imported := importedBy(info, sel.X.(*ast.Ident))
	if imported == nil || speaks(pkg, imported) {
		return ""
	}
	return pkgPath(imported.Path())
}

// speaks reports whether a verb package may take its vocabulary from imported:
// the internal/domain root it sits beneath, or a GROUP package between that
// root and the verb. A command group extends the root vocabulary for its own
// subcommands — internal/domain/client for the client verbs — so a group
// ancestor is the shared vocabulary at its own level. An ancestor that
// declares the contract is a VERB rather than a group, and a verb's types are
// its own: borrowing them makes one command's private spelling the next
// command's vocabulary, which is the drift the shared alias exists to stop. A
// package that is not an ancestor at all is not this verb's vocabulary however
// it is aliased — another group's package is borrowed, and anything outside
// internal/domain is an impostor, including an ancestor above the tier.
func speaks(verb packagePath, imported *types.Package) bool {
	path, root := pkgPath(imported.Path()), sharedVocabulary(verb)
	if path == root {
		return true
	}
	return strings.HasPrefix(string(path), string(root)+"/") &&
		strings.HasPrefix(string(verb), string(path)+"/") &&
		!declaresCommand(imported)
}

// declaresCommand reports whether pkg declares any element of the domain
// contract, which is what makes a package beneath the tier a verb rather than
// a group package. It is declaresContract's question asked of an IMPORTED
// package, whose scope carries production declarations only.
func declaresCommand(pkg *types.Package) bool {
	for _, want := range contract {
		if pkg.Scope().Lookup(string(want.name)) != nil {
			return true
		}
	}
	return false
}

// importedBy is the package ident resolves to, or nil when ident does not name
// an imported package.
func importedBy(info *types.Info, ident *ast.Ident) *types.Package {
	pkg, ok := info.ObjectOf(ident).(*types.PkgName)
	if !ok {
		return nil
	}
	return pkg.Imported()
}

// sharedVocabulary is the import path of the shared vocabulary package for a
// domain command's package path: the internal/domain root the command sits
// beneath, module prefix and all.
func sharedVocabulary(pkg packagePath) pkgPath {
	before, _, _ := strings.Cut(string(pkg), domainDir+"/")
	return pkgPath(before + domainDir)
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

// variadicSelector is decl's variadic element type when it is a qualified
// name — pkg.Name with a plain identifier qualifier — or nil when there is no
// declaration, no variadic, or the element is spelled some other way.
func variadicSelector(decl *ast.FuncDecl) *ast.SelectorExpr {
	if decl == nil || len(decl.Type.Params.List) == 0 {
		return nil
	}
	params := decl.Type.Params.List
	last, ok := params[len(params)-1].Type.(*ast.Ellipsis)
	if !ok {
		return nil
	}
	sel, ok := last.Elt.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	if _, ok := sel.X.(*ast.Ident); !ok {
		return nil
	}
	return sel
}

// spellingOf is the SOURCE spelling of the qualified name sel writes —
// "domain.Argument" for the shared alias — or empty when there is no
// selector.
func spellingOf(sel *ast.SelectorExpr) qualifiedName {
	if sel == nil {
		return ""
	}
	return qualifiedName(sel.X.(*ast.Ident).Name + "." + sel.Sel.Name)
}
