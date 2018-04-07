package transformer

import "github.com/jwowillo/gen/page"

// Transformer creates a new page.Page from the given page.Page.
type Transformer interface {
	// Transform the page.Page.
	//
	// Returns an error if the transformation couldn't be applied.
	Transform(page.Page) (page.Page, error)
}

// Func wraps functions into Transformers.
type Func func(page.Page) (page.Page, error)

// Transform the page.Page by applying the function.
func (f Func) Transform(p page.Page) (page.Page, error) {
	return f(p)
}
