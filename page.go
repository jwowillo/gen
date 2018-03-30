package gen

import (
	"bytes"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
)

// Page is an io.Reader with a path describing its location.
type Page interface {
	Path() string
	io.Reader
}

// NewPage where the io.Reader is at the path.
func NewPage(path string, rd io.Reader) Page {
	return &page{path: path, Reader: rd}
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

// Assets corresponding to all files in the dir directory.
//
// Returns an error if there were any problems reading the directory or files.
func Assets(dir string) ([]Page, error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	ps := make([]Page, 0, len(fs))
	for _, f := range fs {
		a, err := NewAsset(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, err
		}
		ps = append(ps, a)
	}
	return ps, nil
}

// Asset is a static file.
type Asset struct {
	Page
}

// NewAsset from static file at the path that will be accessed from its file
// name.
//
// Returns an error if the file couldn't be read.
func NewAsset(p string) (*Asset, error) {
	bs, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return &Asset{NewPage("/"+path.Base(p), bytes.NewBuffer(bs))}, nil
}
