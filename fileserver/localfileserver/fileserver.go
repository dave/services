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
)

func New(dir string, sites map[string]string) *Fileserver {
	for host, bucket := range sites {
		go http.ListenAndServe(host, pathEscape(http.FileServer(http.Dir(filepath.Join(dir, bucket)))))
	}
	return &Fileserver{
		dir: dir,
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

func (f *Fileserver) Write(ctx context.Context, bucket, name string, reader io.Reader, overwrite bool, contentType, cacheControl string) (saved bool, err error) {
	fdir := filepath.Join(f.dir, bucket)
	fpath := filepath.Join(f.dir, bucket, url.PathEscape(name))
	if !overwrite {
		// err == nil => file exists, return with saved == false
		// os.IsNotExist(err) => file doesn't exist, so continue and write file
		// !os.IsNotExist(err) => any other error, so return the error
		if _, err := os.Stat(fpath); err == nil {
			return false, nil
		} else if !os.IsNotExist(err) {
			return false, err
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
