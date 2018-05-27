package copier

import (
	"io"
	"os"
	"path/filepath"

	"time"

	"fmt"

	billy "gopkg.in/src-d/go-billy.v4"
)

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(src, dest string, srcFs, dstFs billy.Filesystem) error {
	info, err := srcFs.Stat(src)
	if err != nil {
		return err
	}
	return copyInternal(src, dest, info, srcFs, dstFs, nil)
}

type filterFunc func(name string, dir bool) bool

// Filter copies src to dest, doesn't matter if src is a directory or a file, and includes a filter
// function to exclude files / dirs
func Filter(src, dest string, srcFs, dstFs billy.Filesystem, filter filterFunc) error {
	info, err := srcFs.Stat(src)
	if err != nil {
		return err
	}
	return copyInternal(src, dest, info, srcFs, dstFs, filter)
}

// "info" must be given here, NOT nil.
func copyInternal(src, dest string, info os.FileInfo, srcFs, dstFs billy.Filesystem, filter filterFunc) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	if info.IsDir() {
		return dcopy(src, dest, info, srcFs, dstFs, filter)
	}
	return fcopy(src, dest, info, srcFs, dstFs, filter)
}

func fcopy(src, dest string, info os.FileInfo, srcFs, dstFs billy.Filesystem, filter filterFunc) error {

	if filter != nil && !filter(src, false) {
		return nil
	}

	f, err := dstFs.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	var s billy.File
	done := make(chan struct{})
	go func() {
		s, err = srcFs.Open(src)
		close(done)
	}()
	select {
	case <-done:
		// nothing
	case <-time.After(time.Second):
		return fmt.Errorf("timed out opening %s", src)
	}
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dcopy(src, dest string, info os.FileInfo, srcFs, dstFs billy.Filesystem, filter filterFunc) error {

	if filter != nil && !filter(src, true) {
		return nil
	}

	if err := dstFs.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	infos, err := srcFs.ReadDir(src)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if err := copyInternal(
			filepath.Join(src, info.Name()),
			filepath.Join(dest, info.Name()),
			info,
			srcFs,
			dstFs,
			filter,
		); err != nil {
			return err
		}
	}

	return nil
}
