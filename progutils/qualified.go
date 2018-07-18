package progutils

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/loader"
)

// QualifiedIdentifierInfo gets info about a package from the SelectorExpr
// packagePath: the package path of the imported package
// packageName: the actual name of the imported package
// importAlias: the alias in the import statement (can be "")
// codeAlias: the alias used in the code (if importAlias == "", codeAlias == packageName)
func QualifiedIdentifierInfo(se *ast.SelectorExpr, path string, prog *loader.Program) (packagePath, packageName, importAlias, codeAlias string) {
	id, ok := se.X.(*ast.Ident)
	if !ok {
		return
	}
	use, ok := prog.Package(path).Uses[id]
	if !ok {
		return
	}
	pn, ok := use.(*types.PkgName)
	if !ok {
		return
	}
	packagePath = pn.Imported().Path()
	packageName = prog.Package(packagePath).Pkg.Name()
	codeAlias = pn.Name()
	if packageName != codeAlias {
		importAlias = codeAlias
	}
	return
}
