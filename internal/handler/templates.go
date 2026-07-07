package handler

import (
	"fmt"
	"html/template"
	"io/fs"

	"github.com/lancer/log/internal/markdown"
	"github.com/lancer/log/web"
)

// TemplatesFromEmbed builds the per-page template set from the embedded web/templates FS.
func TemplatesFromEmbed() (*Templates, error) {
	tplFS, err := fs.Sub(web.TemplatesFS, "templates")
	if err != nil {
		return nil, err
	}
	funcs := FuncMap()
	for k, v := range MarkdownFuncs() {
		funcs[k] = v
	}
	return LoadTemplates(tplFS, []string{"layout.tmpl"}, map[string]string{
		"index": "index.tmpl", "post": "post.tmpl",
		"about": "about.tmpl", "archive": "archive.tmpl", "shelf": "shelf.tmpl", "notfound": "notfound.tmpl",
	}, funcs)
}

// Templates holds per-page parsed templates (each a clone of the base layout).
type Templates struct {
	base  *template.Template
	pages map[string]*template.Template
}

// LoadTemplates parses the base layout + partials once, then clones for each
// page template so per-page {{define "content"}} / {{define "title"}} don't clash.
func LoadTemplates(tplFS fs.FS, baseFiles []string, pageFiles map[string]string, funcs template.FuncMap) (*Templates, error) {
	// base
	var base *template.Template
	for _, f := range baseFiles {
		raw, err := fs.ReadFile(tplFS, f)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f, err)
		}
		if base == nil {
			base, err = template.New("base").Funcs(funcs).Parse(string(raw))
		} else {
			base, err = base.Parse(string(raw))
		}
		if err != nil {
			return nil, fmt.Errorf("parse base %s: %w", f, err)
		}
	}

	t := &Templates{base: base, pages: map[string]*template.Template{}}
	for name, f := range pageFiles {
		raw, err := fs.ReadFile(tplFS, f)
		if err != nil {
			return nil, fmt.Errorf("read page %s: %w", f, err)
		}
		clone, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("clone for %s: %w", name, err)
		}
		clone, err = clone.Parse(string(raw))
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", f, err)
		}
		t.pages[name] = clone
	}
	return t, nil
}

// Execute renders the named page. Each page clone's body is the page file's
// root text ({{template "layout" .}} + defines), executed via the "base" name.
func (t *Templates) Execute(wr interface{ Write([]byte) (int, error) }, name string, data any) error {
	tpl, ok := t.pages[name]
	if !ok {
		return fmt.Errorf("unknown page template %q", name)
	}
	return tpl.ExecuteTemplate(wr, "base", data)
}

// MarkdownFuncs exposes markdown helpers for templates that need raw HTML rendering.
func MarkdownFuncs() template.FuncMap {
	return template.FuncMap{
		"mdRender": func(s string) template.HTML {
			if s == "" {
				return ""
			}
			return template.HTML(markdown.Render(s))
		},
	}
}
