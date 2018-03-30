package gen

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// Gzip is a Transformer which gzips Pages.
type Gzip struct{}

// NewGzip Transformer.
func NewGzip() Gzip {
	return Gzip{}
}

// Transform returns the gzipped Page.
//
// Returns an error if the Page couldn't be gzipped.
func (t Gzip) Transform(p Page) (Page, error) {
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
	return NewPage(p.Path(), buf), nil
}
