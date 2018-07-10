package astcutil

import (
	"fmt"
	"go/ast"
)

func Width(n ast.Node) int {
	switch n := n.(type) {
	case nil:
		return 0

	// Comments and fields
	case *ast.Comment:
		if n == nil {
			return 0
		}
		return len(n.Text)

	case *ast.CommentGroup:
		if n == nil {
			return 0
		}
		var count int
		for i, c := range n.List {
			if i > 0 {
				// between comments, there can be multiple whitespace. We calculate this by comparing
				// the start and end positions of the comments.
				// count += int(c.Pos() - n.List[i-1].End())

				// actually, once formatted correctly, there is always a single whitespace between comments
				// within a comment group
				// TODO: confirm this.
				count++
			}
			count += Width(c)
		}
		return count

	case *ast.Field:
		if n == nil {
			return 0
		}
		var count int
		count += Width(n.Doc)
		for i, name := range n.Names {
			if i > 0 {
				count++ // comma
			}
			count += Width(name)
			count++ // separator / whitespace after name
		}
		count += Width(n.Type)
		if n.Tag != nil {
			count++ // whitespace before tag
		}
		count += Width(n.Tag)
		if n.Comment != nil {
			// TODO: needs test
			count++ // whitespace before comment
		}
		count += Width(n.Comment)
		return count

	case *ast.FieldList:
		var count int
		for i, f := range n.List {
			if i > 0 {
				fmt.Println(n.List[i-1].End(), f.Pos())
			}
			count += Width(f)
		}
		return count

		// Expressions
	case *ast.BadExpr:
		if n == nil {
			return 0
		}
		return int(n.To - n.From)

	case *ast.Ident:
		if n == nil {
			return 0
		}
		return len(n.Name)

	case *ast.BasicLit:
		if n == nil {
			return 0
		}
		return len(n.Value)

	case *ast.Ellipsis:
		if n == nil {
			return 0
		}
		return Width(n.Elt) + 3

	case *ast.FuncLit:
		//a.apply(n, "Type", nil, n.Type)
		//a.apply(n, "Body", nil, n.Body)

	case *ast.CompositeLit:
		//a.apply(n, "Type", nil, n.Type)
		//a.applyList(n, "Elts")

	case *ast.ParenExpr:
		//a.apply(n, "X", nil, n.X)

	case *ast.SelectorExpr:
		//a.apply(n, "X", nil, n.X)
		//a.apply(n, "Sel", nil, n.Sel)

	case *ast.IndexExpr:
		//a.apply(n, "X", nil, n.X)
		//a.apply(n, "Index", nil, n.Index)

	case *ast.SliceExpr:
		//a.apply(n, "X", nil, n.X)
		//a.apply(n, "Low", nil, n.Low)
		//a.apply(n, "High", nil, n.High)
		//a.apply(n, "Max", nil, n.Max)

	case *ast.TypeAssertExpr:
		//a.apply(n, "X", nil, n.X)
		//a.apply(n, "Type", nil, n.Type)

	case *ast.CallExpr:
		//a.apply(n, "Fun", nil, n.Fun)
		//a.applyList(n, "Args")

	case *ast.StarExpr:
		//a.apply(n, "X", nil, n.X)

	case *ast.UnaryExpr:
		//a.apply(n, "X", nil, n.X)

	case *ast.BinaryExpr:
		//a.apply(n, "X", nil, n.X)
		//a.apply(n, "Y", nil, n.Y)

	case *ast.KeyValueExpr:
		//a.apply(n, "Key", nil, n.Key)
		//a.apply(n, "Value", nil, n.Value)

		// Types
	case *ast.ArrayType:
		//a.apply(n, "Len", nil, n.Len)
		//a.apply(n, "Elt", nil, n.Elt)

	case *ast.StructType:
		//a.apply(n, "Fields", nil, n.Fields)

	case *ast.FuncType:
		//a.apply(n, "Params", nil, n.Params)
		//a.apply(n, "Results", nil, n.Results)

	case *ast.InterfaceType:
		//a.apply(n, "Methods", nil, n.Methods)

	case *ast.MapType:
		//a.apply(n, "Key", nil, n.Key)
		//a.apply(n, "Value", nil, n.Value)

	case *ast.ChanType:
		//a.apply(n, "Value", nil, n.Value)

		// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		//a.apply(n, "Decl", nil, n.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		//a.apply(n, "Label", nil, n.Label)
		//a.apply(n, "Stmt", nil, n.Stmt)

	case *ast.ExprStmt:
		//a.apply(n, "X", nil, n.X)

	case *ast.SendStmt:
		//a.apply(n, "Chan", nil, n.Chan)
		//a.apply(n, "Value", nil, n.Value)

	case *ast.IncDecStmt:
		//a.apply(n, "X", nil, n.X)

	case *ast.AssignStmt:
		//a.applyList(n, "Lhs")
		//a.applyList(n, "Rhs")

	case *ast.GoStmt:
		//a.apply(n, "Call", nil, n.Call)

	case *ast.DeferStmt:
		//a.apply(n, "Call", nil, n.Call)

	case *ast.ReturnStmt:
		//a.applyList(n, "Results")

	case *ast.BranchStmt:
		//a.apply(n, "Label", nil, n.Label)

	case *ast.BlockStmt:
		//a.applyList(n, "List")

	case *ast.IfStmt:
		//a.apply(n, "Init", nil, n.Init)
		//a.apply(n, "Cond", nil, n.Cond)
		//a.apply(n, "Body", nil, n.Body)
		//a.apply(n, "Else", nil, n.Else)

	case *ast.CaseClause:
		//a.applyList(n, "List")
		//a.applyList(n, "Body")

	case *ast.SwitchStmt:
		//a.apply(n, "Init", nil, n.Init)
		//a.apply(n, "Tag", nil, n.Tag)
		//a.apply(n, "Body", nil, n.Body)

	case *ast.TypeSwitchStmt:
		//a.apply(n, "Init", nil, n.Init)
		//a.apply(n, "Assign", nil, n.Assign)
		//a.apply(n, "Body", nil, n.Body)

	case *ast.CommClause:
		//a.apply(n, "Comm", nil, n.Comm)
		//a.applyList(n, "Body")

	case *ast.SelectStmt:
		//a.apply(n, "Body", nil, n.Body)

	case *ast.ForStmt:
		//a.apply(n, "Init", nil, n.Init)
		//a.apply(n, "Cond", nil, n.Cond)
		//a.apply(n, "Post", nil, n.Post)
		//a.apply(n, "Body", nil, n.Body)

	case *ast.RangeStmt:
		//a.apply(n, "Key", nil, n.Key)
		//a.apply(n, "Value", nil, n.Value)
		//a.apply(n, "X", nil, n.X)
		//a.apply(n, "Body", nil, n.Body)

		// Declarations
	case *ast.ImportSpec:
		//a.apply(n, "Doc", nil, n.Doc)
		//a.apply(n, "Name", nil, n.Name)
		//a.apply(n, "Path", nil, n.Path)
		//a.apply(n, "Comment", nil, n.Comment)

	case *ast.ValueSpec:
		//a.apply(n, "Doc", nil, n.Doc)
		//a.applyList(n, "Names")
		//a.apply(n, "Type", nil, n.Type)
		//a.applyList(n, "Values")
		//a.apply(n, "Comment", nil, n.Comment)

	case *ast.TypeSpec:
		//a.apply(n, "Doc", nil, n.Doc)
		//a.apply(n, "Name", nil, n.Name)
		//a.apply(n, "Type", nil, n.Type)
		//a.apply(n, "Comment", nil, n.Comment)

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		//a.apply(n, "Doc", nil, n.Doc)
		//a.applyList(n, "Specs")

	case *ast.FuncDecl:
		//a.apply(n, "Doc", nil, n.Doc)
		//a.apply(n, "Recv", nil, n.Recv)
		//a.apply(n, "Name", nil, n.Name)
		//a.apply(n, "Type", nil, n.Type)
		//a.apply(n, "Body", nil, n.Body)

		// Files and packages
	case *ast.File:
		//a.apply(n, "Doc", nil, n.Doc)
		//a.apply(n, "Name", nil, n.Name)
		//a.applyList(n, "Decls")

		// Don't walk n.Comments; they have either been walked already if
		// they are Doc comments, or they can be easily walked explicitly.

	case *ast.Package:
		// collect and sort names for reproducible behavior
		//var names []string
		//for name := range n.Files {
		//	names = append(names, name)
		//}
		//sort.Strings(names)
		//for _, name := range names {
		//	a.apply(n, name, nil, n.Files[name])
		//}

	default:

	}

	panic(fmt.Errorf("node type %T not implemented", n))
}
