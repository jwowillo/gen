package gen

import (
	"bytes"
	"html/template"
)

// Template is a Golang HTML template.
type Template struct {
	Page
}

// NewTemplate from template files tmpls with x injected into it at path p.
func NewTemplate(tmpls []string, x interface{}, p string) (*Template, error) {
	tmpl, err := template.ParseFiles(tmpls...)
	tmpl = tmpl.Funcs(template.FuncMap{
		"inc": func(x int) int { return x + 1 },
	})
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, x); err != nil {
		return nil, err
	}
	return &Template{NewPage(p, buf)}, nil
}
