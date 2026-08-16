package clidomain

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

// TestIsTestFilePosReadsTheCompiledName pins the one thing that decides
// whether a declaration counts: the name the COMPILER read, taken from
// token.File, rather than the name Fset.Position reports. A //line directive
// rewrites the second and cannot touch the first, so a production file that
// claims to be a test does not take its own declarations out of the contract.
// A source fixture cannot pin this — analysistest matches its want comments by
// the rewritten position, so the directive moves the expectation with it — so
// the divergence is built here, with AddLineColumnInfo, which is exactly what
// the directive does to a real file.
func TestIsTestFilePosReadsTheCompiledName(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	fset := token.NewFileSet()
	production := fset.AddFile("greet.go", -1, 100)
	production.AddLineColumnInfo(10, "greet_test.go", 1, 1)
	forged := production.Pos(20)

	pass := &analysis.Pass{Fset: fset}
	want.Equal("greet_test.go", fset.Position(forged).Filename, "the directive rewrites the reported name")
	want.False(isTestFilePos(pass, forged), "the compiled name is what decides, and it is greet.go")

	test := fset.AddFile("greet_test.go", -1, 100)
	want.True(isTestFilePos(pass, test.Pos(1)), "a file the compiler read as a test is one")
	want.False(isTestFilePos(pass, token.NoPos), "a position in no file of this pass is no test declaration")
}

// TestProductionAnchorIsTotal pins the answer for a package with no production
// file at all. A verb always has one — declaresContract answers to a
// production declaration, so run never reaches this — and the anchor is still
// written as a total function rather than as an index that would be out of
// range if it ever were reached.
func TestProductionAnchorIsTotal(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	onlyTest := fset.AddFile("greet_test.go", -1, 100)
	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{{Package: onlyTest.Pos(1)}},
	}

	assert.Equal(t, token.NoPos, productionAnchor(pass), "no production file anchors nowhere")
	assert.Equal(t, token.NoPos, productionAnchor(&analysis.Pass{Fset: fset}), "no file at all anchors nowhere")
}
