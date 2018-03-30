package gen

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// ErrNoPage is returned when a Page that doesn't exist is referenced.
var ErrNoPage = errors.New("no Page with path found")

// Bundle is a Transformer which bundles all referenced CSS and JS Pages
// referenced in a Page.
type Bundle struct {
	assets map[string][]byte
	err    error
}

// NewBundle that has access to references in the Pages.
func NewBundle(ps []Page) *Bundle {
	assets, err := findAssets(ps)
	return &Bundle{assets: assets, err: err}
}

// Transform the Page by replacing all referenced CSS and JS Pages with their
// Pages.
//
// Returns ErrNoPage if a referenced Page can't be found or if bundling fails.
func (t Bundle) Transform(p Page) (Page, error) {
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
	return NewPage(p.Path(), buf), nil
}

// bundleLine converts all lines that reference Pages into their referenced
// assets and stores them back in the bytes.Buffer.
//
// Returns ErrNoPage if a referenced Page can't be found.
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
	// jsPrefix is the prefix of included JS scripts.
	jsPrefix = []byte("script src=\"/")
	// cssPrefix is the prefix of included CSS scripts.
	cssPrefix = []byte("link rel=\"stylesheet\" href=\"/")
	// quote is the end of an included script.
	quote = []byte("\"")
)

// findAssets referenced in all the Pages.
//
// Return an error if any Page couldn't be read.
func findAssets(ps []Page) (map[string][]byte, error) {
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

// referencedAsset returns the path of the asset referenced in bs and true if
// one is referenced or an empty string and false if one isn't referenced.
func referencedAsset(bs []byte) (string, bool) {
	if path, ok := referencedJS(bs); ok {
		return path, true
	}
	if path, ok := referencedCSS(bs); ok {
		return path, true
	}
	return "", false
}

// referencedJS returns the referenced JS Page's path and returns true if one is
// referenced or an empty string and false if one isn't referenced.
func referencedJS(line []byte) (string, bool) {
	return referenced(line, jsPrefix, quote)
}

// referencedCSS returns the referenced CSS Page's path and returns true if one
// is referenced or an empty string and false if one isn't referenced.
func referencedCSS(line []byte) (string, bool) {
	return referenced(line, cssPrefix, quote)
}

// referenced returns the referenced Page's path and returns true if one is
// referenced or an empty string and false if one isn't referenced.
//
// The reference is determined by a string surrounded by the start and end
// bytes.
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

// bundleAsset with extension and content to the bytes.Buffer.
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
