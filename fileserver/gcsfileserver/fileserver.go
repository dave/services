package gcsfileserver

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

func New(client *storage.Client, buckets []string) *Fileserver {
	f := new(Fileserver)
	f.client = client
	f.buckets = map[string]*storage.BucketHandle{}
	for _, b := range buckets {
		f.buckets[b] = client.Bucket(b)
	}
	return f
}

type Fileserver struct {
	client  *storage.Client
	buckets map[string]*storage.BucketHandle
}

func (f *Fileserver) Exists(ctx context.Context, bucket, name string) (bool, error) {
	return f.exists(ctx, f.buckets[bucket].Object(name))
}

func (f *Fileserver) exists(ctx context.Context, ob *storage.ObjectHandle) (bool, error) {
	_, err := ob.Attrs(ctx)
	if err == nil {
		// err == nil => file exists
		return true, nil
	}
	if err == storage.ErrObjectNotExist {
		// err == storage.ErrObjectNotExist => file doesn't exist
		return false, nil
	}
	// err != storage.ErrObjectNotExist => an error, so return the error
	return false, err
}

func (f *Fileserver) Write(ctx context.Context, bucket, name string, reader io.Reader, overwrite bool, contentType, cacheControl string) (saved bool, err error) {
	ob := f.buckets[bucket].Object(name)
	if !overwrite {
		exists, err := f.exists(ctx, ob)
		if err != nil {
			return false, err
		}
		if exists {
			return false, nil
		}
	}
	wc := ob.NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = contentType
	wc.CacheControl = cacheControl
	if _, err := io.Copy(wc, reader); err != nil {
		return false, err
	}
	return true, nil
}

func (f *Fileserver) Read(ctx context.Context, bucket, name string, writer io.Writer) (found bool, err error) {
	ob := f.buckets[bucket].Object(name)
	r, err := ob.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, err
	}
	if _, err := io.Copy(writer, r); err != nil {
		return false, err
	}
	return true, nil
}
