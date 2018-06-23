package fsprinter

import (
	"fmt"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v4"
)

// Print prints the filesystem to stdout for debugging
func Print(fs billy.Filesystem) {
	err := PrintDir(fs, "/")
	if err != nil {
		fmt.Println(err)
	}
}

func PrintDir(fs billy.Filesystem, dir string) error {
	fis, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		fpath := filepath.Join(dir, fi.Name())
		fmt.Println(fpath)
		if fi.IsDir() {
			if err := PrintDir(fs, fpath); err != nil {
				return err
			}
		}
	}
	return nil
}
