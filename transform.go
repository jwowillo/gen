package gen

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"mime"
	"path/filepath"
	"regexp"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/xml"
)

// Minify the Page if it has a CSS, HTML, JavaScript, or XML file extension.
//
// All other Pages are skipped.
//
// Returns an error if the Page couldn't be minified.
func Minify(p Page) (Page, error) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	ext := filepath.Ext(p.Path())
	if ext != ".xml" && ext != ".css" && ext != ".js" && ext != ".html" {
		return p, nil
	}
	t := mime.TypeByExtension(ext)
	buf := &bytes.Buffer{}
	if err := m.Minify(t, buf, p); err != nil {
		return nil, err
	}
	return NewPage(p.Path(), buf), nil
}

// Gzip the Page.
//
// Returns an error if the Page couldn't be gzipped.
func Gzip(p Page) (Page, error) {
	bs, err := ioutil.ReadAll(p)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	zw := gzip.NewWriter(buf)
	if _, err := zw.Write(bs); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return NewPage(p.Path(), buf), nil
}
