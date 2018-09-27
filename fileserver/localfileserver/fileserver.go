package localfileserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

func New(dir string, sites []string, host, bucket map[string]string) *Fileserver {
	expanded, err := homedir.Expand(dir)
	if err != nil {
		panic(err)
	}
	for _, site := range sites {
		go http.ListenAndServe(host[site], pathEscape(http.FileServer(http.Dir(filepath.Join(expanded, bucket[site])))))
	}
	return &Fileserver{
		dir: expanded,
	}
}

func pathEscape(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = "/" + url.PathEscape(strings.TrimPrefix(r.URL.Path, "/"))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r2)
	})
}

type Fileserver struct {
	dir string
}

func (f *Fileserver) Exists(ctx context.Context, bucket, name string) (bool, error) {
	return f.exists(ctx, filepath.Join(f.dir, bucket, url.PathEscape(name)))
}

func (f *Fileserver) exists(ctx context.Context, fpath string) (bool, error) {
	_, err := os.Stat(fpath)
	if err == nil {
		// err == nil => file exists
		return true, nil
	}
	if os.IsNotExist(err) {
		// os.IsNotExist(err) => file doesn't exist
		return false, nil
	}
	// !os.IsNotExist(err) => any other error, so return the error
	return false, err
}

func (f *Fileserver) Write(ctx context.Context, bucket, name string, reader io.Reader, overwrite bool, contentType, cacheControl string) (saved bool, err error) {
	fdir := filepath.Join(f.dir, bucket)
	fpath := filepath.Join(f.dir, bucket, url.PathEscape(name))
	if !overwrite {
		exists, err := f.exists(ctx, fpath)
		if err != nil {
			return false, err
		}
		if exists {
			return false, nil
		}
	}
	if err := os.MkdirAll(fdir, 0777); err != nil {
		return false, err
	}
	fmt.Println("creating", fpath)
	file, err := os.Create(fpath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return false, err
	}
	return true, nil
}

func (f *Fileserver) Read(ctx context.Context, bucket, name string, writer io.Writer) (found bool, err error) {
	fpath := filepath.Join(f.dir, bucket, url.PathEscape(name))
	file, err := os.Open(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer file.Close()
	if _, err := io.Copy(writer, file); err != nil {
		return false, err
	}
	return true, nil
}
