// Package gen allows Pages to be transformed and written to a directory to
// generate a website.
//
// Some common Page types are provided.
package gen

import (
	"io"
	"os"
	"path/filepath"
)

// Page is an io.Reader with a path describing its location.
type Page interface {
	Path() string
	io.Reader
}

// page with path and io.Reader.
type page struct {
	path string
	io.Reader
}

// Path of the Page.
func (p page) Path() string {
	return p.path
}

// NewPage where the io.Reader rd is at the path.
func NewPage(path string, rd io.Reader) Page {
	return &page{path: path, Reader: rd}
}

// Write the Pages to the out directory after applying the Transforms in order.
//
// Returns an error if any of the Transforms couldn't by applied or any Pages
// couldn't be written.
func Write(out string, ts []Transform, ps []Page) error {
	ps, err := apply(compose(ts), ps)
	if err != nil {
		return err
	}
	return WriteOnly(out, ps)
}

// WriteOnly writes the Pages to the out directory and applies no Transforms.
//
// This is useful for debugging Pages.
//
// Returns an error if any Pages couldn't be written.
func WriteOnly(out string, ps []Page) error {
	for _, p := range ps {
		full := filepath.Join(out, p.Path())
		if err := os.MkdirAll(
			filepath.Dir(full),
			os.ModePerm,
		); err != nil {
			return err
		}
		f, err := os.Create(full)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(f, p); err != nil {
			return err
		}
	}
	return nil
}

// Transform accepts one Page and returns a transformed Page or an error if the
// Transform couldn't be applied.
type Transform func(Page) (Page, error)

// apply the Transform to all the Pages and return an error if any fails.
func apply(t Transform, ps []Page) ([]Page, error) {
	as := make([]Page, 0, len(ps))
	var err error
	for _, p := range ps {
		p, err = t(p)
		if err != nil {
			return nil, err
		}
		as = append(as, p)
	}
	return as, nil
}

// compose all the Transforms together.
//
// Any error caused during any Transform causes all to fail and the error to be
// returned.
func compose(ts []Transform) Transform {
	if len(ts) == 0 {
		return func(p Page) (Page, error) {
			return p, nil
		}
	}
	return func(p Page) (Page, error) {
		var err error
		for _, t := range ts {
			p, err = t(p)
			if err != nil {
				return nil, err
			}
		}
		return p, nil
	}
}
