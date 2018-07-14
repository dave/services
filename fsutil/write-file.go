package fsutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v4"
)

func WriteFile(fs billy.Filesystem, fpath string, perm os.FileMode, contents interface{}) error {

	dir, _ := filepath.Split(fpath)
	if err := fs.MkdirAll(dir, 0777); err != nil {
		return err
	}

	f, err := fs.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	var r io.Reader

	switch contents := contents.(type) {
	case string:
		r = bytes.NewBufferString(contents)
	case []byte:
		r = bytes.NewBuffer(contents)
	case io.Reader:
		r = contents
	default:
		return fmt.Errorf("unsupported contents type %T", contents)
	}

	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}
