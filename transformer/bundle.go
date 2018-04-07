package transformer

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/jwowillo/gen/page"
)

// ErrNoPage is returned when a page.Page that doesn't exist is referenced.
var ErrNoPage = errors.New("no page.Page with path found")

// Bundle is a Transformer which bundles all referenced CSS and JS page.Pages
// referenced in a page.Page.
type Bundle struct {
	assets map[string][]byte
	err    error
}

// NewBundle that has access to references in the page.Pages.
func NewBundle(ps []page.Page) *Bundle {
	assets, err := findAssets(ps)
	return &Bundle{assets: assets, err: err}
}

// Transform the page.Page by replacing all referenced CSS and JS page.Pages with their
// page.Pages.
//
// Returns ErrNopage.Page if a referenced page.Page can't be found or if bundling fails.
func (t Bundle) Transform(p page.Page) (page.Page, error) {
	if t.err != nil {
		return nil, t.err
	}
	if filepath.Ext(p.Path()) != ".html" {
		return p, nil
	}
	bs, err := ioutil.ReadAll(p)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	for _, line := range bytes.Split(bs, []byte{'\n'}) {
		if err := t.bundleLine(buf, line); err != nil {
			return nil, err
		}
	}
	return page.New(p.Path(), buf), nil
}

func (t Bundle) bundleLine(buf *bytes.Buffer, bs []byte) error {
	path, ok := referencedAsset(bs)
	if !ok {
		buf.Write(append(bs, '\n'))
		return nil
	}
	asset, ok := t.assets[path]
	if !ok {
		return ErrNoPage
	}
	bundleAsset(buf, filepath.Ext(path), asset)
	return nil
}

var (
	jsPrefix  = []byte("script src=\"/")
	cssPrefix = []byte("link rel=\"stylesheet\" href=\"/")
	quote     = []byte("\"")
)

func findAssets(ps []page.Page) (map[string][]byte, error) {
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

func referencedAsset(bs []byte) (string, bool) {
	if path, ok := referencedJS(bs); ok {
		return path, true
	}
	if path, ok := referencedCSS(bs); ok {
		return path, true
	}
	return "", false
}

func referencedJS(line []byte) (string, bool) {
	return referenced(line, jsPrefix, quote)
}

func referencedCSS(line []byte) (string, bool) {
	return referenced(line, cssPrefix, quote)
}

func referenced(line, start, end []byte) (string, bool) {
	i := bytes.Index(line, start)
	if i == -1 {
		return "", false
	}
	path := line[i+len(start)-1:]
	j := bytes.Index(path, end)
	if j == -1 {
		return "", false
	}
	return string(path[:j]), true
}

func bundleAsset(buf *bytes.Buffer, ext string, bs []byte) {
	switch ext {
	case ".css":
		buf.WriteString("<style>\n")
		buf.Write(bs)
		buf.WriteString("</style>\n")
	case ".js":
		buf.WriteString("<script type='text/javascript'>\n")
		buf.Write(bs)
		buf.WriteString("</script>\n")
	}
}
