package gen

// Transformer creates a new Page from the given Page.
type Transformer interface {
	// Transform the Page.
	//
	// Returns an error if the transformation couldn't be applied.
	Transform(Page) (Page, error)
}

// TransformerFunc wraps functions into Transformers.
type TransformerFunc func(Page) (Page, error)

// Transform the Page by applying the function.
func (f TransformerFunc) Transform(p Page) (Page, error) {
	return f(p)
}
