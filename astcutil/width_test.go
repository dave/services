package astcutil

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestWidth(t *testing.T) {
	parseExpr := func(code string) ast.Expr {

		expr, err := parser.ParseExpr(code)
		if err != nil {
			panic(err)
		}

		buf := &bytes.Buffer{}
		if err := format.Node(buf, token.NewFileSet(), expr); err != nil {
			panic(err)
		}
		if buf.String() != code {
			panic(fmt.Errorf("input code doesn't match formatted AST. input:\n----------------\n%s\n----------------\nAST:\n----------------\n%s\n----------------", code, buf.String()))
		}

		return expr
	}
	parseFile := func(code string) *ast.File {

		if !strings.HasPrefix(code, "package ") {
			// for convenience: handle file with omitted package
			code = "package a\n\n" + code
		}
		if !strings.HasSuffix(code, "\n") {
			code += "\n"
		}

		formatted, err := format.Source([]byte(code))
		if err != nil {
			panic(err)
		}
		if string(formatted) != code {
			panic(fmt.Errorf("input code doesn't match formatted AST. input:\n----------------\n%s\n----------------\nAST:\n----------------\n%s\n----------------", code, string(formatted)))
		}

		fset := token.NewFileSet()

		f, err := parser.ParseFile(fset, "a.go", code, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		// ensure the input string is well formatted

		return f
	}
	tests := map[string]testspec{
		"simple": {
			node:     parseExpr("a").(*ast.Ident),
			expected: 1,
		},
		"string literal": {
			node:     parseExpr("\"a\"").(*ast.BasicLit),
			expected: 3,
		},
		"float literal": {
			node:     parseExpr("1.1").(*ast.BasicLit),
			expected: 3,
		},
		"bad expr": {
			node:     &ast.BadExpr{From: 5, To: 10},
			expected: 5,
		},
		"elipsis param": {
			node:     parseExpr("func(i ...T) {\n}").(*ast.FuncLit).Type.Params.List[0].Type.(*ast.Ellipsis),
			expected: 4,
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
			node:     parseFile("// a\n// b").Comments[0],
			expected: 9, // includes whitespace between comments
		},
		"comment group inline": {
			node:     parseFile("/* a */ /* b */").Comments[0],
			expected: 15, // includes whitespace between comments
		},
		"comment group mixed": {
			node:     parseFile("// a\n/* b */ // c").Comments[0],
			expected: 17,
		},
		"field": {
			node:     parseFile("type T struct{ a int }").Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List[0],
			expected: 5,
		},
		"field multiple names": {
			node:     parseFile("type T struct{ a, b int }").Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List[0],
			expected: 8,
		},
		"field multiple names new lines": {
			node:     parseFile("type T struct {\n\ta,\n\tb int\n}").Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List[0],
			expected: 999,
		},
		"field with tag": {
			node:     parseFile("type T struct {\n\ta int `a`\n}").Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List[0],
			expected: 9,
		},
		"fieldlist": {
			node:     parseFile("type T struct {\n\ta int\n\tb string\n}").Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields,
			expected: 999,
		},
		// "field with line comment":
		// TODO: *ast.Field.Comment only for interface / signature?

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
		t.Fatal("single mode so failing")
	}
}

type testspec struct {
	node     ast.Node
	expected int
}
