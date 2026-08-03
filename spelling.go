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
// is NOT the shared vocabulary package — empty when it is the genuine
// article. A qualifier that does not resolve to a package at all also yields
// empty: the type checker guarantees a qualified type's qualifier is a
// package in compiled code, so an unresolved one fails open rather than
// inventing a finding.
func impostorPath(info *types.Info, sel *ast.SelectorExpr, pkg packagePath) pkgPath {
	imported := importedBy(info, sel.X.(*ast.Ident))
	if imported == "" || imported == sharedVocabulary(pkg) {
		return ""
	}
	return imported
}

// importedBy is the import path ident resolves to, or empty when ident does
// not name an imported package.
func importedBy(info *types.Info, ident *ast.Ident) pkgPath {
	pkg, ok := info.ObjectOf(ident).(*types.PkgName)
	if !ok {
		return ""
	}
	return pkgPath(pkg.Imported().Path())
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
