package gen

import (
	"bytes"
	"compress/gzip"
	"errors"
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

// AllTransformations is a list of all implemented Transformations.
var AllTransformations = []Transform{Bundle, Minify, Gzip}

// Minify the Pages with CSS, HTML, JavaScript, or XML file extensions.
//
// All other Pages are skipped.
//
// Returns an error if a Page couldn't be minified.
func Minify(ps []Page) ([]Page, error) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	out := make([]Page, 0, len(ps))
	for _, p := range ps {
		ext := filepath.Ext(p.Path())
		if ext != ".xml" && ext != ".css" && ext != ".js" &&
			ext != ".html" {
			out = append(out, p)
			continue
		}
		t := mime.TypeByExtension(ext)
		buf := &bytes.Buffer{}
		if err := m.Minify(t, buf, p); err != nil {
			return nil, err
		}
		out = append(out, NewPage(p.Path(), buf))
	}
	return out, nil
}

// Gzip the Pages.
//
// Returns an error if any Page couldn't be gzipped.
func Gzip(ps []Page) ([]Page, error) {
	out := make([]Page, 0, len(ps))
	for _, p := range ps {
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
		out = append(out, NewPage(p.Path(), buf))
	}
	return out, nil
}

// ErrNoPage is returned when a Page that doesn't exist is referenced.
var ErrNoPage = errors.New("no Page with path found")

// Bundle all Pages by moving JS and CSS Pages into HTML Pages where they are
// references.
//
// Returns an error if any referenced file couldn't be found or if any Page
// couldn't be read.
func Bundle(ps []Page) ([]Page, error) {
	var out []Page
	assets, err := findAssetPages(ps)
	if err != nil {
		return nil, err
	}
	bundled := make(map[string]interface{})
	for _, p := range ps {
		if filepath.Ext(p.Path()) != ".html" {
			continue
		}
		buf := &bytes.Buffer{}
		bs, err := ioutil.ReadAll(p)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(bs, []byte{'\n'}) {
			if err := bundleLine(
				buf, line,
				assets, bundled,
			); err != nil {
				return nil, err
			}
		}
		bundled[p.Path()] = struct{}{}
		out = append(out, NewPage(p.Path(), buf))
	}
	for _, p := range ps {
		if _, ok := bundled[p.Path()]; ok {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func bundleLine(
	buf *bytes.Buffer, bs []byte,
	assets map[string][]byte, bundled map[string]interface{},
) error {
	path, ok := referencedAsset(bs)
	if !ok {
		_, err := buf.Write(append(bs, '\n'))
		return err
	}
	asset, ok := assets[path]
	if !ok {
		return ErrNoPage
	}
	bundled[path] = struct{}{}
	return writePage(buf, path, asset)
}

func referencedAsset(bs []byte) (string, bool) {
	re1 := regexp.MustCompile("script src=\"(.+)\"")
	re2 := regexp.MustCompile("link rel=\"stylesheet\" href=\"(.+)\"")
	re1Matches := re1.FindAllStringSubmatch(string(bs), -1)
	re2Matches := re2.FindAllStringSubmatch(string(bs), -1)
	if len(re1Matches) != 0 {
		path := re1Matches[0][1]
		if path[0] != '/' {
			return "", false
		}
		return path, true
	}
	if len(re2Matches) != 0 {
		path := re2Matches[0][1]
		if path[0] != '/' {
			return "", false
		}
		return path, true
	}
	return "", false
}

func findAssetPages(ps []Page) (map[string][]byte, error) {
	assets := make(map[string][]byte)
	for _, p := range ps {
		ext := filepath.Ext(p.Path())
		if ext != ".js" && ext != ".css" {
			continue
		}
		bs, err := ioutil.ReadAll(p)
		if err != nil {
			return nil, err
		}
		assets[p.Path()] = bs
	}
	return assets, nil
}

func writePage(buf *bytes.Buffer, path string, bs []byte) error {
	ext := filepath.Ext(path)
	if ext == ".css" {
		if _, err := buf.WriteString("<style>\n"); err != nil {
			return err
		}
	}
	if ext == ".js" {
		if _, err := buf.WriteString(
			"<script type='text/javascript'>\n",
		); err != nil {
			return err
		}
	}
	if _, err := buf.Write(bs); err != nil {
		return err
	}
	if ext == ".css" {
		if _, err := buf.WriteString("</style>\n"); err != nil {
			return err
		}
	}
	if ext == ".js" {
		if _, err := buf.WriteString("</script>\n"); err != nil {
			return err
		}
	}
	return nil
}
