package progutils

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/loader"
)

func NewImportsHelper(path string, file *ast.File, prog *loader.Program) *ImportsHelper {
	return &ImportsHelper{path: path, file: file, prog: prog}
}

type ImportsHelper struct {
	path string
	file *ast.File
	prog *loader.Program
}

// RegisterImport
func (ih *ImportsHelper) RegisterImport(path string) (name string, err error) {
	var found bool
	for _, is := range ih.file.Imports {
		importpath, err := strconv.Unquote(is.Path.Value)
		if err != nil {
			return "", err
		}
		if importpath == path {
			if is.Name != nil && is.Name.Name != "" && is.Name.Name != "_" {
				// if current import is aliased, just use that name
				return is.Name.Name, nil
			}
			found = true
			if is.Name != nil && is.Name.Name == "_" {
				is.Name = ast.NewIdent("")
			}
			break
		}
	}
	if !found {
		// if not found, add the import to the ast file imports
		ih.file.Imports = append(ih.file.Imports, &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(path)},
		})
		if err := ih.RefreshFromFile(); err != nil {
			return "", err
		}
	}
	pkg := ih.prog.Package(path)
	if pkg == nil {
		return "", fmt.Errorf("package %s not found in program", path)
	}
	return pkg.Pkg.Name(), nil
}

func (ih *ImportsHelper) ImportsFromTree() (decl *ast.GenDecl, specs map[string]*ast.ImportSpec, imports map[string]string, err error) {
	imports = map[string]string{}
	specs = map[string]*ast.ImportSpec{}
	astutil.Apply(ih.file, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.GenDecl:
			switch n.Tok {
			case token.PACKAGE:
				return true // package -> continue
			case token.IMPORT:
				if decl == nil {
					decl = n
				}
				for _, spec := range n.Specs {
					spec := spec.(*ast.ImportSpec)
					var path, name string
					path, err = strconv.Unquote(spec.Path.Value)
					if err != nil {
						return false
					}
					if spec.Name != nil {
						name = spec.Name.Name
					}
					imports[path] = name
					specs[path] = spec
				}
				return true
			default:
				return false // any other GenDecl -> stop (only package comes before import)
			}
		}
		return true
	}, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	return decl, specs, imports, nil
}

func (ih *ImportsHelper) ImportsFromFile() (map[string]string, error) {
	imports := map[string]string{}
	for _, spec := range ih.file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}
		var name string
		if spec.Name != nil {
			name = spec.Name.Name
		}
		imports[path] = name
	}
	return imports, nil
}

func (ih *ImportsHelper) RefreshFromFile() error {
	imports, err := ih.ImportsFromFile()
	if err != nil {
		return err
	}
	return ih.Refresh(imports)

}

// RefreshFromCode scans all the code for SelectorElements
func (ih *ImportsHelper) RefreshFromCode() error {
	imports := map[string]string{}
	var err error
	astutil.Apply(ih.file, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.SelectorExpr:
			packagePath, _, importAlias, _ := QualifiedIdentifierInfo(n, ih.path, ih.prog)
			if packagePath == "" {
				return true
			}
			currentAlias, ok := imports[packagePath]
			if !ok {
				imports[packagePath] = importAlias
				return true
			}
			if importAlias != currentAlias {
				err = fmt.Errorf("import for %s uses different name in 2 files", packagePath)
				return false
			}
		}
		return true
	}, nil)
	if err != nil {
		return err
	}
	if err := ih.Refresh(imports); err != nil {
		return err
	}
	return nil
}

func (ih *ImportsHelper) Refresh(imports map[string]string) error {

	// first clear any import aliases that are the same as the package name
	for path, name := range imports {
		if ih.prog.Package(path).Pkg.Name() == name {
			imports[path] = ""
		}
	}

	gd, specsFromTree, importsFromTree, err := ih.ImportsFromTree()
	if err != nil {
		return err
	}

	if compareMaps(imports, importsFromTree) {
		// nothing to do here
		return nil
	}

	// Update AST with missing, updated and deleted imports
	missing := map[string]string{}
	deleted := map[string]bool{}
	changed := map[string]string{}

	for path, name := range imports {
		treeName, ok := importsFromTree[path]
		if !ok {
			missing[path] = name
		} else if name != treeName {
			changed[path] = name
		}
	}

	for path := range importsFromTree {
		if _, ok := imports[path]; !ok {
			deleted[path] = true
		}
	}

	/*
		if len(missing) > 0 {
			fmt.Println("adding missing imports:", missing)
		}
		if len(deleted) > 0 {
			fmt.Println("removing deleted imports:", deleted)
		}
		if len(changed) > 0 {
			fmt.Println("updating changed imports:", changed)
		}
	*/

	for path, name := range missing {
		is := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(path),
			},
		}
		if name != "" {
			is.Name = ast.NewIdent(name)
		}
		if gd == nil {
			gd = &ast.GenDecl{
				Tok: token.IMPORT,
			}
			// TODO: do we need to insert after the first (package) Decl?
			ih.file.Decls = append([]ast.Decl{gd}, ih.file.Decls...)
		}
		gd.Specs = append(gd.Specs, is)
		if len(gd.Specs) > 1 {
			// Lparen and Rparen must be non-zero to render as a list
			gd.Lparen = 1
			gd.Rparen = 1
		}
	}

	for path, name := range changed {
		spec := specsFromTree[path]
		if name == "" {
			spec.Name = nil
		} else {
			spec.Name = ast.NewIdent(name)
		}
	}

	if len(deleted) > 0 {
		var err error
		astutil.Apply(ih.file, func(c *astutil.Cursor) bool {
			switch n := c.Node().(type) {
			case *ast.ImportSpec:
				var path string
				path, err = strconv.Unquote(n.Path.Value)
				if err != nil {
					return false
				}
				if deleted[path] {
					c.Delete()
				}
				return true
			case *ast.GenDecl:
				switch n.Tok {
				case token.PACKAGE, token.IMPORT:
					return true
				default:
					return false // stop as soon as we reach a GenDecl that's not a package or import
				}
			}
			return true
		}, nil)
		if err != nil {
			return err
		}
		// Scan again deleting any empty GenDecls
		astutil.Apply(ih.file, func(c *astutil.Cursor) bool {
			switch n := c.Node().(type) {
			case *ast.GenDecl:
				switch n.Tok {
				case token.IMPORT:
					if len(n.Specs) == 0 {
						c.Delete()
					}
					return true
				case token.PACKAGE:
					return true
				default:
					return false // stop as soon as we reach a GenDecl that's not a package or import
				}
			}
			return true
		}, nil)
	}

	// update File.Imports
	var updated []*ast.ImportSpec
	astutil.Apply(ih.file, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.ImportSpec:
			updated = append(updated, n)
		case *ast.GenDecl:
			switch n.Tok {
			case token.PACKAGE, token.IMPORT:
				return true
			default:
				return false // stop as soon as we reach a GenDecl that's not a package or import
			}
		}
		return true
	}, nil)
	ih.file.Imports = updated

	return nil

}

func compareMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if v != vb {
			return false
		}
	}
	return true
}
