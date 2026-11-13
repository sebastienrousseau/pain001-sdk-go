// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// The size ceilings from the workspace standard; gocyclo and gocognit
// in `make check` hold the complexity ones.
const (
	maxFuncLines = 60
	maxFileLines = 500
)

// TestFunctionAndFileSizes keeps every Go function, tests included,
// within 60 lines and every Go file within 500.
func TestFunctionAndFileSizes(t *testing.T) {
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		if n := fset.File(file.Pos()).LineCount(); n > maxFileLines {
			t.Errorf("%s: %d lines, ceiling %d", path, n, maxFileLines)
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				checkFunc(t, fset, path, fn)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func checkFunc(t *testing.T, fset *token.FileSet, path string, fn *ast.FuncDecl) {
	t.Helper()
	lines := fset.Position(fn.End()).Line - fset.Position(fn.Pos()).Line + 1
	if lines > maxFuncLines {
		t.Errorf("%s: %s is %d lines, ceiling %d", path, fn.Name.Name, lines, maxFuncLines)
	}
}
