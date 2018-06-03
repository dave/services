package srcimporter

import (
	"bytes"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func NewBuildContext(source map[string]map[string]string, tags []string) *build.Context {

	tags = append(tags, "js", "netgo", "purego", "jsgo")

	b := &build.Context{
		GOARCH:        "amd64",  // Target architecture
		GOOS:          "darwin", // Target operating system
		GOROOT:        "goroot", // Go root
		GOPATH:        "gopath", // Go path
		InstallSuffix: "",       // Builder only: "min" or "".
		Compiler:      "gc",     // Compiler to assume when computing target paths
		BuildTags:     tags,     // Build tags
		CgoEnabled:    false,    // Builder only: detect `import "C"` to throw proper error
		ReleaseTags:   build.Default.ReleaseTags,

		IsDir: func(path string) bool {
			pkg := dir2pkg(path)
			if _, ok := source[pkg]; ok {
				return true
			}
			for p := range source {
				if strings.HasPrefix(p, pkg+"/") {
					return true
				}
			}
			return false
		},

		HasSubdir: func(root, dir string) (rel string, ok bool) {
			// copied from default implementation to prevent use of filepath.EvalSymlinks
			const sep = string(filepath.Separator)
			root = filepath.Clean(root)
			if !strings.HasSuffix(root, sep) {
				root += sep
			}
			dir = filepath.Clean(dir)
			if !strings.HasPrefix(dir, root) {
				return "", false
			}
			return filepath.ToSlash(dir[len(root):]), true
		},

		ReadDir: func(path string) ([]os.FileInfo, error) {
			pkg := dir2pkg(path)
			idx := len(strings.Split(pkg, "/")) // len(pkg parts) will be the index of the next part (which is the name of any subdir)
			files, ok := source[pkg]
			if !ok {
				return nil, os.ErrNotExist
			}
			var fis []os.FileInfo
			for p := range source {
				if strings.HasPrefix(p, pkg+"/") {
					fis = append(fis, file{name: strings.Split(p, "/")[idx], dir: true})
				}
			}
			for name, contents := range files {
				fis = append(fis, file{name: name, length: len(contents)})
			}
			return fis, nil
		},

		// OpenFile opens a file (not a directory) for reading.
		// If OpenFile is nil, Import uses os.Open.
		OpenFile: func(path string) (io.ReadCloser, error) {
			dir, name := filepath.Split(path)
			files, ok := source[dir2pkg(dir)]
			if !ok {
				return nil, os.ErrNotExist
			}
			contents, ok := files[name]
			if !ok {
				return nil, os.ErrNotExist
			}
			return ioutil.NopCloser(bytes.NewBuffer([]byte(contents))), nil
		},
	}
	return b
}

func dir2pkg(dir string) string {
	const sep = string(filepath.Separator)
	dir = strings.TrimPrefix(dir, sep) // trim leading slash
	parts := strings.Split(dir, sep)   // split into parts
	if len(parts) <= 2 {
		return ""
	}
	dir = strings.Join(parts[2:], sep) // dirs[0] == "gopath", dirs[1] == "src"
	return strings.Trim(filepath.ToSlash(dir), "/")
}

type file struct {
	name   string
	length int
	dir    bool
}

func (f file) Name() string {
	return f.name
}

func (f file) Size() int64 {
	return int64(f.length)
}

func (f file) Mode() os.FileMode {
	if f.dir {
		return os.ModeDir
	} else {
		return 0
	}
}

func (f file) ModTime() time.Time {
	return time.Time{}
}

func (f file) IsDir() bool {
	return f.dir
}

func (file) Sys() interface{} {
	return nil
}
