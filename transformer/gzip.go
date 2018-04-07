package transformer

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"github.com/jwowillo/gen/page"
)

// Gzip is a Transformer which gzips page.Pages.
type Gzip struct{}

// NewGzip Transformer.
func NewGzip() Gzip {
	return Gzip{}
}

// Transform returns the gzipped page.Page.
//
// Returns an error if the page.Page couldn't be gzipped.
func (t Gzip) Transform(p page.Page) (page.Page, error) {
	bs, err := ioutil.ReadAll(p)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	zw := gzip.NewWriter(buf)
	defer zw.Close()
	if _, err := zw.Write(bs); err != nil {
		return nil, err
	}
	return page.New(p.Path(), buf), nil
}
