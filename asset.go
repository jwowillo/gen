package gen

import (
	"bytes"
	"io/ioutil"
	"path"
	"path/filepath"
)

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

// NewAsset from static file at path p that will be accessed from only its file
// name.
//
// Returns an error if the file couldn't be read.
func NewAsset(p string) (*Asset, error) {
	bs, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return &Asset{NewPage(path.Base(p), bytes.NewBuffer(bs))}, nil
}
