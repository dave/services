package deployer

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/dave/services/builder"
	"github.com/dave/services/builder/buildermsg"
	"github.com/dave/services/constor"
	"github.com/dave/services/deployer/deployermsg"
	"github.com/gopherjs/gopherjs/compiler"
)

func (d *Deployer) Update(ctx context.Context, source map[string]map[string]string, cache map[string]string, min bool) error {

	storer := constor.New(ctx, d.session.Fileserver, d.send, d.config.ConcurrentStorageUploads)
	defer storer.Close()

	d.send(buildermsg.Building{Starting: true})

	b := builder.New(d.session, d.defaultOptions(min))

	index := deployermsg.ArchiveIndex{}
	done := map[string]bool{}

	b.Callback = func(archive *compiler.Archive) error {

		if done[archive.ImportPath] {
			return nil
		}

		done[archive.ImportPath] = true

		if archive.Name == "main" {
			return nil
		}

		if d.session.HasSource(archive.ImportPath) {
			// don't return anything if the package is in the source collection
			return nil
		}

		hashPair, standard := d.index[archive.ImportPath]
		var hash string
		var js []byte
		if standard {
			hash = hashPair[min]
		} else {
			var b []byte
			var err error
			js, b, err = builder.GetPackageCode(ctx, archive, min, true)
			if err != nil {
				return err
			}
			hash = fmt.Sprintf("%x", b)
		}

		var unchanged bool
		if cached, exists := cache[archive.ImportPath]; exists && cached == hash {
			unchanged = true
		}

		index[archive.ImportPath] = deployermsg.ArchiveIndexItem{
			Hash:      hash,
			Unchanged: unchanged,
		}

		if unchanged {
			// If the dependency is unchanged from the client cache, don't return it as a PlaygroundArchive
			// message
			return nil
		}

		if standard {
			d.send(deployermsg.Archive{
				Path:     archive.ImportPath,
				Hash:     hash,
				Standard: true,
			})
		} else {
			var count uint32
			done := func() {
				if atomic.AddUint32(&count, 1) == 2 {
					d.send(deployermsg.Archive{
						Path:     archive.ImportPath,
						Hash:     hash,
						Standard: false,
					})
				}
			}
			storer.Add(constor.Item{
				Message:   archive.Name,
				Name:      fmt.Sprintf("%s.%s.js", archive.ImportPath, hash), // Note: hash is a string
				Contents:  js,
				Bucket:    d.config.PkgBucket,
				Mime:      constor.MimeJs,
				Count:     true,
				Immutable: true,
				Send:      true,
				Done:      done,
			})
			buf := &bytes.Buffer{}
			if err := compiler.WriteArchive(StripArchive(archive), buf); err != nil {
				return err
			}
			storer.Add(constor.Item{
				Message:   "",
				Name:      fmt.Sprintf("%s.%s.ax", archive.ImportPath, hash), // Note: hash is a string
				Contents:  buf.Bytes(),
				Bucket:    d.config.PkgBucket,
				Mime:      constor.MimeBin,
				Count:     true,
				Immutable: true,
				Send:      true,
				Done:      done,
			})
		}
		return nil
	}

	if cachedPrelude, exists := cache["prelude"]; !exists || cachedPrelude != d.prelude[min] {
		// send the prelude first if it's not in the cache
		d.send(deployermsg.Archive{
			Path:     "prelude",
			Hash:     d.prelude[min],
			Standard: true,
		})
	}

	// All programs need runtime and it's dependencies
	if _, _, err := b.BuildImportPath(ctx, "runtime"); err != nil {
		return err
	}

	for path := range source {
		if _, _, err := b.BuildImportPath(ctx, path); err != nil {
			return err
		}
	}

	if err := storer.Wait(); err != nil {
		return err
	}

	d.send(index)

	d.send(buildermsg.Building{Done: true})

	return nil
}

func StripArchive(a *compiler.Archive) *compiler.Archive {
	out := &compiler.Archive{
		ImportPath: a.ImportPath,
		Name:       a.Name,
		Imports:    a.Imports,
		ExportData: a.ExportData,
		Minified:   a.Minified,
	}
	for _, d := range a.Declarations {
		// All that's needed in Declarations is FullName (https://github.com/gopherjs/gopherjs/blob/423bf76ba1888a53d4fe3c1a82991cdb019a52ad/compiler/package.go#L187-L191)
		out.Declarations = append(out.Declarations, &compiler.Decl{FullName: d.FullName, Blocking: d.Blocking})
	}
	return out
}
