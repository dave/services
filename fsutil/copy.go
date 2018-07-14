package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/src-d/go-billy.v4"
)

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(dstFs billy.Filesystem, dest string, srcFs billy.Filesystem, src string) error {
	info, err := srcFs.Stat(src)
	if err != nil {
		return err
	}
	return copyInternal(src, dest, info, srcFs, dstFs, nil)
}

type filterFunc func(name string, dir bool) bool

// Filter copies src to dest, doesn't matter if src is a directory or a file, and includes a filter
// function to exclude files / dirs
func Filter(dstFs billy.Filesystem, dest string, srcFs billy.Filesystem, src string, filter filterFunc) error {
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

func fcopy(srcpath, destpath string, srcinfo os.FileInfo, srcfs, destfs billy.Filesystem, filter filterFunc) error {

	if filter != nil && !filter(srcpath, false) {
		return nil
	}

	dir, _ := filepath.Split(destpath)
	if err := destfs.MkdirAll(dir, 0777); err != nil {
		return err
	}
	dst, err := destfs.OpenFile(destpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcinfo.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	var src billy.File
	done := make(chan struct{})
	go func() {
		src, err = srcfs.Open(srcpath)
		close(done)
	}()
	select {
	case <-done:
		// nothing
	case <-time.After(time.Second):
		return fmt.Errorf("timed out opening %s", srcpath)
	}
	if err != nil {
		return err
	}
	defer src.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
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
