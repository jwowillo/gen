package transformer

import (
	"bytes"
	"mime"
	"path/filepath"
	"regexp"

	"github.com/jwowillo/gen/page"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/xml"
)

// Minify is a Transformer which minifies page.Pages.
type Minify struct {
	minifier *minify.M
}

// NewMinify which can HTML, CSS, JS, and XML files.
func NewMinify() *Minify {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	return &Minify{minifier: m}
}

// Transform the page.Page by minifying it if it is an HTML, CSS, JS, and XML file
// and skipping it otherwise.
//
// Returns an error if the page.Page couldn't be minified.
func (t Minify) Transform(p page.Page) (page.Page, error) {
	ext := mime.TypeByExtension(filepath.Ext(p.Path()))
	buf := &bytes.Buffer{}
	if err := t.minifier.Minify(ext, buf, p); err != nil {
		if err == minify.ErrNotExist {
			return p, nil
		}
		return nil, err
	}
	return page.New(p.Path(), buf), nil
}
