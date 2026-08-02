package clidomain

// The variadic-spelling rule: Run's variadic must be WRITTEN as the shared
// domain.Argument. Both it and a package-local alias resolve to string, so
// these checks read the source spelling, not the resolved type — the
// distinction is the name, which is the whole point of standardising it.

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

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
