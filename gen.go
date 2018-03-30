// Package gen allows Pages to be transformed and written to a directory to
// generate a website.
//
// Some Transformers and Pages are provided.
package gen

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/jwowillo/pipe.v1"
)

// Write the Pages to the directory after applying the Transformers in order.
//
// Returns an error if any of the Transformers couldn't by applied or any Pages
// couldn't be written.
func Write(out string, ps []Page, ts ...Transformer) []error {
	p, errs := makePipe(ts)
	for _, page := range ps {
		p.Receive(page)
	}
	var wg sync.WaitGroup
	wg.Add(len(ps))
	for i := 0; i < len(ps); i++ {
		page := p.Deliver().(Page)
		go func() {
			if err := write(out, page); err != nil {
				errs = append(errs, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return errs
}

// WriteOnly writes the Pages to the out directory.
//
// Returns an error if any Pages couldn't be written.
func WriteOnly(out string, ps []Page) []error {
	return Write(out, ps)
}

// WriteWithDefaults writes the Pages to the out directory after applying the
// Bundle, Minify, and Gzip transforms.
//
// Returns an error if any Transformers couldn't be applied or any Pages
// couldn't be written.
func WriteWithDefaults(out string, ps []Page) []error {
	return Write(out, ps, NewBundle(ps), NewMinify(), NewGzip())
}

// write the Page to the directory.
func write(out string, p Page) error {
	full := filepath.Join(out, p.Path())
	if err := os.MkdirAll(filepath.Dir(full), os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, p)
	return err
}

func makePipe(ts []Transformer) (*pipe.Pipe, []error) {
	var errs []error
	ss := make([]pipe.Stage, len(ts))
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		ss[i] = pipe.StageFunc(func(x pipe.Item) pipe.Item {
			p, err := t.Transform(x.(Page))
			if err != nil {
				errs = append(errs, err)
			}
			return p
		})
	}
	return pipe.New(ss...), errs
}
