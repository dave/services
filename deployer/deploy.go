package deployer

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/dave/services/builder"
	"github.com/dave/services/builder/buildermsg"
	"github.com/dave/services/constor"
	"github.com/dave/services/constor/constormsg"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

// Deploy compiles and deploys path.
func (d *Deployer) Deploy(ctx context.Context, path string, index IndexType, minified map[bool]bool) (map[bool]*DeployOutput, error) {

	storer := constor.New(ctx, d.session.Fileserver, d.send, d.config.ConcurrentStorageUploads)
	defer storer.Close()

	d.send(buildermsg.Building{Starting: true})
	d.send(constormsg.Storing{Starting: true})

	wg := &sync.WaitGroup{}

	outputs := map[bool]*builder.CommandOutput{}
	mainHashes := map[bool][]byte{}
	indexHashes := map[bool][]byte{}

	var outer error

	do := func(min bool) {
		defer wg.Done()

		var err error
		var data *builder.PackageData

		data, outputs[min], err = d.compileAndStore(ctx, path, storer, min)
		if err != nil {
			outer = err
			return
		}

		d.send(buildermsg.Building{Message: "Loader"})

		mainHashes[min], err = d.genMain(ctx, storer, outputs[min], min)
		if err != nil {
			outer = err
			return
		}

		d.send(buildermsg.Building{Message: "Index"})

		tpl, err := d.getIndexTpl(data.Dir)
		if err != nil {
			outer = err
			return
		}

		indexHashes[min], err = d.genIndex(storer, tpl, path, mainHashes[min], min, index)
		if err != nil {
			outer = err
			return
		}
	}

	if minified[true] {
		// deploy the minified version
		wg.Add(1)
		go do(true)
	}

	if minified[false] {
		// deploy the non-minified version
		wg.Add(1)
		go do(false)
	}

	wg.Wait()

	if outer != nil {
		return nil, outer
	}

	d.send(buildermsg.Building{Done: true})

	if err := storer.Wait(); err != nil {
		return nil, err
	}

	d.send(constormsg.Storing{Done: true})

	out := map[bool]*DeployOutput{}
	for min := range outputs {
		out[min] = &DeployOutput{
			CommandOutput: outputs[min],
			MainHash:      mainHashes[min],
			IndexHash:     indexHashes[min],
		}
	}

	return out, nil

}

type IndexType int

const (
	HashIndex IndexType = iota
	PathIndex
)

type DeployOutput struct {
	*builder.CommandOutput
	MainHash, IndexHash []byte
}

func (d *Deployer) defaultOptions(min bool) *builder.Options {
	return &builder.Options{
		Temporary:   memfs.New(),
		Unvendor:    true,
		Initializer: true,
		Send:        d.send,
		Verbose:     true,
		Minify:      min,
		Standard:    d.index,
	}
}

func (d *Deployer) compileAndStore(ctx context.Context, path string, storer *constor.Storer, min bool) (*builder.PackageData, *builder.CommandOutput, error) {

	b := builder.New(d.session, d.defaultOptions(min))

	data, archive, err := b.BuildImportPath(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	if archive.Name != "main" {
		return nil, nil, fmt.Errorf("can't compile - %s is not a main package", path)
	}

	output, err := b.WriteCommandPackage(ctx, archive)
	if err != nil {
		return nil, nil, err
	}

	for _, po := range output.Packages {
		if !po.Store {
			continue
		}
		storer.Add(constor.Item{
			Message:   po.Path,
			Name:      fmt.Sprintf("%s.%x.js", po.Path, po.Hash),
			Contents:  po.Contents,
			Bucket:    d.config.PkgBucket,
			Mime:      constor.MimeJs,
			Count:     true,
			Immutable: true,
			Send:      true,
		})
	}

	return data, output, nil
}

func (d *Deployer) getIndexTpl(dir string) (*template.Template, error) {
	fs := d.session.Filesystem(dir)
	fname := filepath.Join(dir, "index.jsgo.html")
	_, err := fs.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return indexTemplate, nil
		}
		return nil, err
	}
	f, err := fs.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	tpl, err := template.New("main").Parse(string(b))
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

type IndexVars struct {
	Path   string
	Hash   string
	Script string
}

var indexTemplate = template.Must(template.New("main").Parse(`
<html>
	<head>
		<meta charset="utf-8">
	</head>
	<body id="wrapper">
		<span id="jsgo-progress-span"></span>
		<script>
			window.jsgoProgress = function(count, total) {
				if (count === total) {
					document.getElementById("jsgo-progress-span").style.display = "none";
				} else {
					document.getElementById("jsgo-progress-span").innerHTML = count + "/" + total;
				}
			}
		</script>
		<script src="{{ .Script }}"></script>
	</body>
</html>
`))

func (d *Deployer) genIndex(storer *constor.Storer, tpl *template.Template, path string, loaderHash []byte, min bool, index IndexType) ([]byte, error) {

	v := IndexVars{
		Path:   path,
		Hash:   fmt.Sprintf("%x", loaderHash),
		Script: fmt.Sprintf("%s://%s/%s.%x.js", d.config.PkgProtocol, d.config.PkgHost, path, loaderHash),
	}

	buf := &bytes.Buffer{}
	sha := sha1.New()

	if err := tpl.Execute(io.MultiWriter(buf, sha), v); err != nil {
		return nil, err
	}

	indexHash := sha.Sum(nil)

	if index == HashIndex {
		storer.Add(constor.Item{
			Message:   "Index",
			Name:      fmt.Sprintf("%x", indexHash),
			Contents:  buf.Bytes(),
			Bucket:    d.config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     true,
			Immutable: true,
			Send:      true,
		})
		storer.Add(constor.Item{
			Message:   "",
			Name:      fmt.Sprintf("%x/index.html", indexHash),
			Contents:  buf.Bytes(),
			Bucket:    d.config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     true,
			Immutable: true,
			Send:      true,
		})
	} else {
		fullpath := path
		if !min {
			fullpath = fmt.Sprintf("%s$max", path)
		}
		shortpath := strings.TrimPrefix(fullpath, "github.com/")

		storer.Add(constor.Item{
			Message:   "Index",
			Name:      shortpath,
			Contents:  buf.Bytes(),
			Bucket:    d.config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     false,
			Immutable: false,
		})
		storer.Add(constor.Item{
			Message:   "",
			Name:      fmt.Sprintf("%s/index.html", shortpath),
			Contents:  buf.Bytes(),
			Bucket:    d.config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     false,
			Immutable: false,
		})

		if shortpath != fullpath {
			storer.Add(constor.Item{
				Message:   "",
				Name:      fullpath,
				Contents:  buf.Bytes(),
				Bucket:    d.config.IndexBucket,
				Mime:      constor.MimeHtml,
				Count:     false,
				Immutable: false,
			})
			storer.Add(constor.Item{
				Message:   "",
				Name:      fmt.Sprintf("%s/index.html", fullpath),
				Contents:  buf.Bytes(),
				Bucket:    d.config.IndexBucket,
				Mime:      constor.MimeHtml,
				Count:     false,
				Immutable: false,
			})
		}
	}

	return indexHash, nil

}

func (d *Deployer) genMain(ctx context.Context, storer *constor.Storer, output *builder.CommandOutput, min bool) ([]byte, error) {

	preludeHash := d.prelude[min]
	pkgs := []PkgJson{
		{
			// Always include the prelude dummy package first
			Path: "prelude",
			Hash: preludeHash,
		},
	}
	for _, po := range output.Packages {
		pkgs = append(pkgs, PkgJson{
			Path: po.Path,
			Hash: fmt.Sprintf("%x", po.Hash),
		})
	}

	pkgJson, err := json.Marshal(pkgs)
	if err != nil {
		return nil, err
	}

	m := MainVars{
		PkgProtocol: d.config.PkgProtocol,
		PkgHost:     d.config.PkgHost,
		Path:        output.Path,
		Json:        string(pkgJson),
	}

	buf := &bytes.Buffer{}
	var tmpl *template.Template
	if min {
		tmpl = mainTemplateMinified
	} else {
		tmpl = mainTemplate
	}
	if err := tmpl.Execute(buf, m); err != nil {
		return nil, err
	}

	s := sha1.New()
	if _, err := s.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	hash := s.Sum(nil)

	var message string
	if min {
		message = "Loader (minified)"
	} else {
		message = "Loader (un-minified)"
	}
	storer.Add(constor.Item{
		Message:   message,
		Name:      fmt.Sprintf("%s.%x.js", output.Path, hash),
		Contents:  buf.Bytes(),
		Bucket:    d.config.PkgBucket,
		Mime:      constor.MimeJs,
		Count:     true,
		Immutable: true,
		Send:      true,
	})

	return hash, nil
}

type MainVars struct {
	Path        string
	Json        string
	PkgHost     string
	PkgProtocol string
}

type PkgJson struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// minify with https://skalman.github.io/UglifyJS-online/

var mainTemplateMinified = template.Must(template.New("main").Parse(
	`"use strict";var $mainPkg,$load={};!function(){for(var n=0,t=0,e={{ .Json }},o=(document.getElementById("log"),function(){n++,window.jsgoProgress&&window.jsgoProgress(n,t),n==t&&function(){for(var n=0;n<e.length;n++)$load[e[n].path]();$mainPkg=$packages["{{ .Path }}"],$synthesizeMethods(),$packages.runtime.$init(),$go($mainPkg.$init,[]),$flushConsole()}()}),a=function(n){t++;var e=document.createElement("script");e.src=n,e.onload=o,e.onreadystatechange=o,document.head.appendChild(e)},s=0;s<e.length;s++)a("{{ .PkgProtocol }}://{{ .PkgHost }}/"+e[s].path+"."+e[s].hash+".js")}();`,
))
var mainTemplate = template.Must(template.New("main").Parse(`"use strict";
var $mainPkg;
var $load = {};
(function(){
	var count = 0;
	var total = 0;
	var path = "{{ .Path }}";
	var info = {{ .Json }};
	var log = document.getElementById("log");
	var finished = function() {
		for (var i = 0; i < info.length; i++) {
			$load[info[i].path]();
		}
		$mainPkg = $packages[path];
		$synthesizeMethods();
		$packages["runtime"].$init();
		$go($mainPkg.$init, []);
		$flushConsole();
	}
	var done = function() {
		count++;
		if (window.jsgoProgress) { window.jsgoProgress(count, total); }
		if (count == total) { finished(); }
	}
	var get = function(url) {
		total++;
		var tag = document.createElement('script');
		tag.src = url;
		tag.onload = done;
		tag.onreadystatechange = done;
		document.head.appendChild(tag);
	}
	for (var i = 0; i < info.length; i++) {
		get("{{ .PkgProtocol }}://{{ .PkgHost }}/" + info[i].path + "." + info[i].hash + ".js");
	}
})();`))
