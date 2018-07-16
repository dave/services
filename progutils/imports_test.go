package progutils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"golang.org/x/tools/go/loader"
)

func TestImports(t *testing.T) {
	tests := map[string]testspec{
		"simple": {
			files: map[string]map[string]string{
				"a": {"a.go": `func main(){}`},
				"b": {"b.go": `package b`},
			},
			action: func(ih *ImportsHelper) error {
				if _, err := ih.RegisterImport("b"); err != nil {
					return err
				}
				return nil
			},
			expected: `import "b"; func main(){}`,
		},
		"multi": {
			files: map[string]map[string]string{
				"a": {"a.go": `func main(){}`},
				"b": {"b.go": `package b`},
				"c": {"c.go": `package c`},
			},
			action: func(ih *ImportsHelper) error {
				if _, err := ih.RegisterImport("b"); err != nil {
					return err
				}
				if _, err := ih.RegisterImport("c"); err != nil {
					return err
				}
				return nil
			},
			expected: `import ("b"; "c"); func main(){}`,
		},
	}

	single := "" // during dev, set this to the name of a test case to just run that single case

	if single != "" {
		tests = map[string]testspec{single: tests[single]}
	}

	var skipped bool
	for _, spec := range tests {
		if spec.skip {
			skipped = true
			continue
		}
		if err := runTest(spec); err != nil {
			t.Fatal(err)
			return
		}
	}

	if single != "" {
		t.Fatal("test passed, but failed because single mode is set")
	}
	if skipped {
		t.Fatal("test passed, but skipped some")
	}
}

type testspec struct {
	skip     bool
	name     string
	files    interface{}                // either map[string]map[string]string, map[string]string or string
	action   func(*ImportsHelper) error // either Mutator or []Mutator
	expected interface{}                // either map[string]map[string]string, map[string]string or string
}

func runTest(spec testspec) error {

	fset := token.NewFileSet()
	packages := normalize(spec.files)
	expected := normalize(spec.expected)

	astPackages := map[string][]*ast.File{}
	var mainFile *ast.File
	for path, files := range packages {
		for fname, contents := range files {
			f, err := parser.ParseFile(fset, fname, contents, parser.ParseComments)
			if err != nil {
				return err
			}
			astPackages[path] = append(astPackages[path], f)
			if path == "a" && fname == "a.go" {
				mainFile = f
			}
		}
	}

	c := loader.Config{
		ParserMode: parser.ParseComments,
		Fset:       fset,
		Cwd:        "/",
	}
	for path, files := range astPackages {
		c.CreateFromFiles(path, files...)
	}
	prog, err := c.Load()
	if err != nil {
		return err
	}

	ih := NewImportsHelper(mainFile, prog)
	if err := spec.action(ih); err != nil {
		return err
	}

	// first count the files in the expected
	var count int
	for _, files := range expected {
		count += len(files)
	}

	// only interested in package a, file a.go
	buf := &bytes.Buffer{}
	if err := format.Node(buf, fset, mainFile); err != nil {
		return err
	}

	expectedBytes, err := format.Source([]byte(expected["a"]["a.go"]))
	if err != nil {
		return err
	}

	if buf.String() != string(expectedBytes) {
		return fmt.Errorf("unexpected contents - expected:\n------------------------------------\n%s\n------------------------------------\nactual:\n------------------------------------\n%s\n------------------------------------\n", string(expectedBytes), buf.String())
	}

	return nil
}

func normalize(i interface{}) map[string]map[string]string {
	var m map[string]map[string]string
	switch v := i.(type) {
	case map[string]map[string]string:
		m = v
	case map[string]string:
		m = map[string]map[string]string{"a": v}
	case string:
		m = map[string]map[string]string{"a": {"a.go": v}}
	}
	for path, files := range m {
		for name, contents := range files {
			if !strings.HasPrefix(strings.TrimSpace(contents), "package ") {
				m[path][name] = "package a\n" + contents
			}
		}
	}
	return m
}
