package includer

import (
	"bytes"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func New(source map[string]string, tags []string) *Includer {
	return &Includer{
		bctx: newBuildContext(source, tags),
	}
}

type Includer struct {
	bctx *build.Context
}

func (i *Includer) Include(name string) (bool, error) {
	if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
		return false, nil
	}
	match, err := i.bctx.MatchFile("/", name)
	if err != nil {
		return false, err
	}
	return match, nil
}

func newBuildContext(source map[string]string, tags []string) *build.Context {

	tags = append(tags, "js", "netgo", "purego", "jsgo")

	b := &build.Context{
		GOARCH:        "js",     // Target architecture
		GOOS:          "darwin", // Target operating system
		GOROOT:        "goroot", // Go root
		GOPATH:        "gopath", // Go path
		InstallSuffix: "",       // Builder only: "min" or "".
		Compiler:      "gc",     // Compiler to assume when computing target paths
		BuildTags:     tags,     // Build tags
		CgoEnabled:    false,    // Builder only: detect `import "C"` to throw proper error
		ReleaseTags:   build.Default.ReleaseTags,

		IsDir:     func(path string) bool { panic("should not be called by includer") },
		HasSubdir: func(root, dir string) (rel string, ok bool) { panic("should not be called by includer") },
		ReadDir:   func(path string) ([]os.FileInfo, error) { panic("should not be called by includer") },

		// OpenFile opens a file (not a directory) for reading.
		// If OpenFile is nil, Import uses os.Open.
		OpenFile: func(path string) (io.ReadCloser, error) {
			_, name := filepath.Split(path)
			s, ok := source[name]
			if !ok {
				return nil, os.ErrNotExist
			}
			return ioutil.NopCloser(bytes.NewBuffer([]byte(s))), nil
		},
	}
	return b
}
