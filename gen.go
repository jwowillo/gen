// Package gen allows Pages to be transformed and written to a directory to
// generate a website.
//
// Some Transformers and Pages are provided.
package gen

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/jwowillo/pipe.v1"
)

// Write the Pages to the directory after applying the Transformers in order.
//
// Returns an error if any of the Transformers couldn't by applied or any Pages
// couldn't be written.
func Write(out string, ps []Page, ts ...Transformer) []error {
	p, errs := makePipe(ts)
	f := pipe.ConsumerFunc(func(x pipe.Item) {
		if err := write(out, x.(Page)); err != nil {
			errs = append(errs, err)
		}
	})
	xs := make([]pipe.Item, len(ps))
	for i, page := range ps {
		xs[i] = pipe.Item(page)
	}
	pipe.ProcessAndConsume(p, f, xs...)
	return errs
}

// WriteWithDefaults writes the Pages to the out directory after applying the
// Bundle, Minify, and Gzip transforms.
//
// Returns an error if any Transformers couldn't be applied or any Pages
// couldn't be written.
func WriteWithDefaults(out string, ps []Page) []error {
	return Write(out, ps, NewBundle(ps), NewMinify(), NewGzip())
}

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
