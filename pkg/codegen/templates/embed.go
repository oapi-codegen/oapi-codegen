package templates

import (
	"embed"
	"text/template"
)

//go:embed *.tmpl chi/*.tmpl echo/*.tmpl
var templates embed.FS

// Parse parses all templates.
func Parse(t *template.Template) (*template.Template, error) {
	return t.ParseFS(templates, "*.tmpl", "chi/*.tmpl", "echo/*.tmpl")
}
