package gen

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"mime"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/xml"
)

var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("application/javascript", js.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
}

// Minify the Pages with CSS, HTML, JavaScript, or XML file extensions.
//
// All other Pages are skipped.
//
// Returns an error if a Page couldn't be minified.
func Minify(p Page) (Page, error) {
	ext := filepath.Ext(p.Path())
	if ext != ".xml" && ext != ".css" && ext != ".js" &&
		ext != ".html" {
		return p, nil
	}
	t := mime.TypeByExtension(ext)
	buf := &bytes.Buffer{}
	if err := minifier.Minify(t, buf, p); err != nil {
		return nil, err
	}
	return NewPage(p.Path(), buf), nil
}

// Gzip the Pages.
//
// Returns an error if any Page couldn't be gzipped.
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

// ErrNoPage is returned when a Page that doesn't exist is referenced.
var ErrNoPage = errors.New("no Page with path found")

// MakeBundle all Pages by moving JS and CSS Pages into HTML Pages where they are
// references.
//
// Returns an error if any referenced file couldn't be found or if any Page
// couldn't be read.
func MakeBundle(ps []Page) Transform {
	assets, err := findAssetPages(ps)
	if err != nil {
		return func(p Page) (Page, error) {
			return nil, err
		}
	}
	return func(p Page) (Page, error) {
		if filepath.Ext(p.Path()) != ".html" {
			return p, nil
		}
		buf := &bytes.Buffer{}
		bs, err := ioutil.ReadAll(p)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(bs, []byte{'\n'}) {
			if err := bundleLine(buf, line, assets); err != nil {
				return nil, err
			}
		}
		return NewPage(p.Path(), buf), nil
	}
}

func bundleLine(buf *bytes.Buffer, bs []byte, assets map[string][]byte) error {
	path, ok := referencedAsset(bs)
	if !ok {
		_, err := buf.Write(append(bs, '\n'))
		return err
	}
	asset, ok := assets[path]
	if !ok {
		return ErrNoPage
	}
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

func parallelize(ps []Page, f func(Page) (Page, error)) ([]Page, error) {
	var out []Page
	var mux sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(ps))
	var err error
	for _, p := range ps {
		go func(p Page) {
			defer wg.Done()
			p, err = f(p)
			if err != nil {
				return
			}
			if p == nil {
				return
			}
			mux.Lock()
			defer mux.Unlock()
			out = append(out, p)
		}(p)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return out, nil
}
