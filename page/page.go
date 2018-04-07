package page

import "io"

// Page is an io.Reader with a path describing its location.
type Page interface {
	Path() string
	io.Reader
}

// New Page where the io.Reader is at the path.
func New(path string, rd io.Reader) Page {
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
