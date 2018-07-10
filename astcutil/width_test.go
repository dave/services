package astcutil

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestWidthExpr(t *testing.T) {
	parseExpr := func(code string) ast.Expr {
		expr, err := parser.ParseExpr(code)
		if err != nil {
			panic(err)
		}
		return expr
	}
	parseFile := func(code string) *ast.File {
		if !strings.HasPrefix(code, "package ") {
			code = "package a\n" + code
		}
		f, err := parser.ParseFile(token.NewFileSet(), "a.go", code, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		return f
	}
	tests := map[string]testspec{
		"simple": {
			node:     parseExpr("i").(*ast.Ident),
			expected: 1,
		},
		"string literal": {
			node:     parseExpr("\"foo\"").(*ast.BasicLit),
			expected: 5,
		},
		"float literal": {
			node:     parseExpr("1.45").(*ast.BasicLit),
			expected: 4,
		},
		"bad expr": {
			node:     &ast.BadExpr{From: 5, To: 10},
			expected: 5,
		},
		"elipsis param": {
			node:     parseExpr("func(i ...int){}").(*ast.FuncLit).Type.Params.List[0].Type.(*ast.Ellipsis),
			expected: 6,
		},
		"elipsis type": {
			node:     parseExpr("[...]T{1}").(*ast.CompositeLit).Type.(*ast.ArrayType).Len.(*ast.Ellipsis),
			expected: 3,
		},
		"comment line": {
			node:     parseFile("// a").Comments[0].List[0],
			expected: 4,
		},
		"comment inline": {
			node:     parseFile("/* a */").Comments[0].List[0],
			expected: 7,
		},
		"comment group line": {
			node:     parseFile("// a\n// bb").Comments[0],
			expected: 10, // includes one whitespace between comments
		},
		"comment group inline": {
			node:     parseFile("/* a *//* bb */").Comments[0],
			expected: 15,
		},
		"comment group inline with whitespace": {
			node:     parseFile("/* a */   /* bb */").Comments[0],
			expected: 18, // includes three whitespace between comments
		},
	}
	single := ""
	if single != "" {
		tests = map[string]testspec{single: tests[single]}
	}
	for name, spec := range tests {

		calculated := Width(spec.node)

		// check against spec
		if calculated != spec.expected {
			t.Errorf("%s: calculated width (%d) doesn't match expected (%d)", name, calculated, spec.expected)
		}

		// check against AST
		fromast := int(spec.node.End() - spec.node.Pos())
		if calculated != fromast {
			t.Errorf("%s: calculated width (%d) doesn't match AST (%d)", name, calculated, fromast)
		}
	}
	if single != "" {
		t.Fatal("test passed but in single mode so failing")
	}
}

type testspec struct {
	node     ast.Node
	expected int
}
