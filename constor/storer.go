// package constor (concurrent storer) for storing items into a services.Fileserver concurrently
package constor

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

	"github.com/dave/services"
	"github.com/dave/services/constor/constormsg"
)

/*
	pkg.jsgo.io (Pkg)
	-----------------
	<path>.<hash>.js        - deployed pkg / loader JS
	prelude.<hash>.js       - prelude JS
	<path>.<hash>.ax        - stripped archives
	/assets.zip             - assets zip
	<path>.<hash>.json      - package source bundle (for frizz.io)

	jsgo.io (Index)
	---------------
	<hash>.js               - index file deployed by play.jsgo.io
	<hash>/index.html       - index file deployed by play.jsgo.io

	<path>.js               - index file deployed by compile.jsgo.io
	<path>/index.html       - index file deployed by compile.jsgo.io
	<short-path>.js         - index file deployed by compile.jsgo.io
	<short-path>/index.html - index file deployed by compile.jsgo.io

	src.jsgo.io (Src)
	-----------------
	<hash>.json             - project shared by play.jsgo.io

	git.jsgo.io (Git)
    -----------------
    <repo-url>              - git repo archive (repo url is encoded with url.PathEscape)
*/
type Storer struct {
	fileserver services.Fileserver
	queue      chan Item
	wait       sync.WaitGroup
	unchanged  int32
	done       int32
	total      int32
	err        error
	send       func(services.Message)
}

func New(ctx context.Context, fileserver services.Fileserver, send func(services.Message), workers int) *Storer {
	s := &Storer{
		fileserver: fileserver,
		queue:      make(chan Item, 1000),
		wait:       sync.WaitGroup{},
		send:       send,
	}
	for i := 0; i < workers; i++ {
		go s.Worker(ctx)
	}
	return s
}

func (s *Storer) Close() {
	close(s.queue)
}

func (s *Storer) Wait() error {
	s.wait.Wait()
	if s.err != nil {
		return s.err
	}
	return nil
}

func (s *Storer) Worker(ctx context.Context) {
	for item := range s.queue {
		func() {
			defer s.wait.Done()
			if item.Wait != nil {
				defer item.Wait.Done()
			}
			overwrite := true
			cacheControl := "no-cache"
			if item.Immutable {
				overwrite = false
				cacheControl = "public,max-age=31536000,immutable"
			}
			saved, err := s.fileserver.Write(ctx, item.Bucket, item.Name, bytes.NewReader(item.Contents), overwrite, item.Mime, cacheControl)
			if err != nil {
				s.err = err
				return
			}
			if item.Count {
				if saved {
					atomic.AddInt32(&s.done, 1)
				} else {
					atomic.AddInt32(&s.unchanged, 1)
				}
			}
			if item.Done != nil {
				item.Done()
			}
			if item.Send {
				s.sendMessage()
			}
		}()
	}
}

func (s *Storer) sendMessage() {
	if s.send == nil {
		return
	}
	total := int(atomic.LoadInt32(&s.total))
	done := int(atomic.LoadInt32(&s.done))
	unchanged := int(atomic.LoadInt32(&s.unchanged))
	s.send(constormsg.Storing{Finished: done, Unchanged: unchanged, Remain: total - done - unchanged})
}

func (s *Storer) Add(item Item) {
	s.wait.Add(1)

	if item.Count {
		atomic.AddInt32(&s.total, 1)
	}
	if item.Send {
		s.sendMessage()
	}

	s.queue <- item

}

const (
	MimeJson = "application/json"
	MimeJs   = "application/javascript"
	MimeBin  = "application/octet-stream"
	MimeHtml = "text/html"
	MimeZip  = "application/zip"
)

type Item struct {
	Message   string
	Bucket    string
	Name      string
	Contents  []byte
	Mime      string
	Immutable bool
	Count     bool
	Wait      *sync.WaitGroup
	Send      bool
	Done      func()
}
