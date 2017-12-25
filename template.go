package gen

import (
	"bytes"
	"html/template"
	"path/filepath"
)

// Template is a Golang HTML template.
type Template struct {
	Page
}

// NewTemplate from template files tmpls with x injected into it at path p.
func NewTemplate(tmpls []string, x interface{}, p string) (*Template, error) {
	base := filepath.Base(tmpls[0])
	tmpl, err := template.New(base).Funcs(template.FuncMap{
		"inc": func(x int) int { return x + 1 },
	}).ParseFiles(tmpls...)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(buf, base, x); err != nil {
		return nil, err
	}
	return &Template{NewPage(p, buf)}, nil
}
