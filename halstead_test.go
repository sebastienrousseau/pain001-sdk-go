// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// maxHalsteadDifficulty is the workspace standard's ceiling per
// function. gocyclo and gocognit hold the other two complexity ones.
const maxHalsteadDifficulty = 30

// halstead counts one function's operators and operands from its
// tokens: keywords, operators and delimiters are operators; identifiers
// and literals are operands.
type halstead struct {
	operators, operands map[string]int
}

// difficulty is (distinct operators / 2) * (operands / distinct operands).
func (h halstead) difficulty() float64 {
	if len(h.operands) == 0 {
		return 0
	}
	total := 0
	for _, n := range h.operands {
		total += n
	}
	return float64(len(h.operators)) / 2 * float64(total) / float64(len(h.operands))
}

// measure tokenizes src, the text of one function.
func measure(src []byte) halstead {
	h := halstead{operators: map[string]int{}, operands: map[string]int{}}
	var s scanner.Scanner
	fset := token.NewFileSet()
	s.Init(fset.AddFile("", -1, len(src)), src, nil, 0)
	for {
		_, tok, lit := s.Scan()
		switch {
		case tok == token.EOF:
			return h
		case tok == token.SEMICOLON && lit == "\n":
			// An inserted semicolon is a line end, not an operator.
		case tok == token.IDENT || tok.IsLiteral():
			h.operands[lit]++
		default:
			h.operators[tok.String()]++
		}
	}
}

// TestHalsteadDifficulty keeps every Go function, tests included, at a
// Halstead difficulty of 30 or less.
func TestHalsteadDifficulty(t *testing.T) {
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return err
		}
		base := fset.File(file.Pos()).Base()
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				body := src[int(fn.Pos())-base : int(fn.End())-base]
				if d := measure(body).difficulty(); d > maxHalsteadDifficulty {
					t.Errorf("%s: %s has Halstead difficulty %.1f, ceiling %d", path, fn.Name.Name, d, maxHalsteadDifficulty)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHalsteadCountsOperatorsAndOperands(t *testing.T) {
	// Operators: func ( ) { return + * }, 8 distinct. Operands: f, a
	// three times, int twice and 2: 7 in all, 4 distinct. 8/2 * 7/4 = 7.
	h := measure([]byte("func f(a int) int {\n\treturn a + a*2\n}"))
	if got := h.difficulty(); got != 7 {
		t.Errorf("difficulty %v, want 7", got)
	}
	if (halstead{}).difficulty() != 0 {
		t.Error("no operands must give 0")
	}
}
