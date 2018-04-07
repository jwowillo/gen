// Package gen allows page.Pages to be transformed and written to a directory to
// generate a website.
//
// Some transformer.Transformers and page.Pages are provided.
package gen

//go:generate fill gopkg.in/jwowillo/pipe.v1 Item=page.Page

import (
	"io"
	"os"
	"path/filepath"

	"github.com/jwowillo/gen/page"
	"github.com/jwowillo/gen/transformer"

	"github.com/jwowillo/gen/internal/page.page/pipe.v1"
)

// Write the page.Pages to the directory after applying the
// transformer.Transformers in order.
//
// Returns an error if any of the transformer.Transformers couldn't by applied
// or any page.Pages couldn't be written.
func Write(out string, ps []page.Page, ts ...transformer.Transformer) []error {
	p, errs := makePipe(ts)
	f := pipe.ConsumerFunc(func(x page.Page) {
		if err := write(out, x); err != nil {
			errs = append(errs, err)
		}
	})
	pipe.ProcessAndConsume(p, f, ps...)
	return errs
}

// WriteWithDefaults writes the page.Pages to the out directory after applying
// the transformer.Bundle, transformer.Minify, and transformer.Gzip transforms.
//
// Returns an error if any transformer.Transformers couldn't be applied or any
// page.Pages couldn't be written.
func WriteWithDefaults(out string, ps []page.Page) []error {
	return Write(
		out, ps,
		transformer.NewBundle(ps),
		transformer.NewMinify(),
		transformer.NewGzip(),
	)
}

func write(out string, p page.Page) error {
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

func makePipe(ts []transformer.Transformer) (*pipe.Pipe, []error) {
	var errs []error
	ss := make([]pipe.Stage, len(ts))
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		ss[i] = pipe.StageFunc(func(x page.Page) page.Page {
			p, err := t.Transform(x)
			if err != nil {
				errs = append(errs, err)
			}
			return p
		})
	}
	return pipe.New(ss...), errs
}
