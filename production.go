package clidomain

// A declaration counts when the COMPILER read it from a production file, and
// that sentence carries the whole of this file. The in-package test-variant
// pass folds a package's _test.go declarations into the same scope as its
// source, so a Config-shaped table or a Run-shaped driver would otherwise turn
// a grouping package into a verb, and a builder method on Config would turn a
// flag record into behaviour — each prescribing that the author delete a
// legitimate test helper. Which file a declaration came from is read from
// token.File rather than from a Position, because a Position is a thing the
// judged file rewrites.

import (
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// productionAnchor is where a missing-contract report belongs: the package
// clause of the verb's FIRST PRODUCTION file. A finding a _test.go file holds
// is dropped by the source-only runner, so anchoring one there deletes it
// rather than moving it. A verb always has a production file, because
// declaresContract answers to a production declaration; token.NoPos is the
// total answer for a package that has none.
func productionAnchor(pass *analysis.Pass) token.Pos {
	for _, file := range pass.Files {
		if !isTestFilePos(pass, file.Package) {
			return file.Package
		}
	}
	return token.NoPos
}

// lookup resolves a package-scope name to its object, but only when it is
// declared in a PRODUCTION file. The in-package test-variant pass folds
// _test.go declarations into the same scope, where a Config-shaped test table
// or a Run-shaped test driver would otherwise turn a grouping package into a
// verb — imposing the contract on vocabulary the tests own.
func lookup(pass *analysis.Pass, name declName) types.Object {
	obj := pass.Pkg.Scope().Lookup(string(name))
	if obj == nil || isTestFilePos(pass, obj.Pos()) {
		return nil
	}
	return obj
}

// isTestFilePos reports whether pos sits in a _test.go file. The name comes
// from token.File, which is the name the COMPILER read — never
// Fset.Position, whose Filename a //line directive in the judged file
// rewrites, letting a production file declare itself a test and take every
// declaration it holds out of the contract. A position from outside this
// pass's file set — a method reached through an imported package's export
// data — belongs to no file here and is not a test declaration.
func isTestFilePos(pass *analysis.Pass, pos token.Pos) bool {
	file := pass.Fset.File(pos)
	return file != nil && strings.HasSuffix(file.Name(), "_test.go")
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

// declaredInProduction reports whether set holds a method declared outside a
// _test.go file. A method set does not inherit lookup's production-only fold,
// so it is applied here: an in-package test's builder method on Config is the
// tests' own vocabulary, and calling it behaviour would prescribe deleting a
// legitimate test helper.
func declaredInProduction(pass *analysis.Pass, set *types.MethodSet) bool {
	for method := range set.Methods() {
		if !isTestFilePos(pass, method.Obj().Pos()) {
			return true
		}
	}
	return false
}
