package progutils

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

func RefreshImports(f *ast.File) {

	// Delete all imports from the ast tree
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.GenDecl:
			if n.Tok == token.IMPORT {
				c.Delete()
			}
		case *ast.ImportSpec:
			c.Delete()
		}
		return true
	}, nil)

	// Add them again from f.Imports
	if len(f.Imports) > 0 {
		var importSpecs []ast.Spec
		for _, is := range f.Imports {
			importSpecs = append(importSpecs, is)
		}
		importGenDecl := &ast.GenDecl{
			Tok:    token.IMPORT,
			Lparen: token.Pos(1), // must be non-zero to render as a list
			Specs:  importSpecs,
			Rparen: token.Pos(1), // must be non-zero to render as a list
		}
		f.Decls = append([]ast.Decl{importGenDecl}, f.Decls...)
	}

}
