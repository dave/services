package astcutil

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

func TestPrinter(t *testing.T) {
	in := `package a

type T struct{
	a int
}`
	out, err := format.Source([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(out))

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "a.go", in, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	//ast.Print(fset, f)

	buf := &bytes.Buffer{}
	if err := format.Node(buf, token.NewFileSet(), f.Decls[0]); err != nil {
		t.Fatal(err)
	}

	fmt.Println(buf.String())

}
