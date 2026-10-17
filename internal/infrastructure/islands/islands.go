// Package islands resuelve componentes ("islas") definidos como archivos del
// tema: un template SSR opcional (templates/islands/<name>.html), un módulo JS
// opcional (assets/islands/<name>.js) y un CSS opcional.
package islands

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"strings"
)

// TemplateRenderer ejecuta plantillas del tema por nombre.
type TemplateRenderer interface {
	Render(w io.Writer, name string, data any) error
}

// Registry conoce las islas disponibles en el tema.
type Registry struct {
	renderer  TemplateRenderer
	templates map[string]bool
	assets    map[string]bool
}

// New escanea el tema y construye el registro de islas.
func New(themeFS fs.FS) *Registry {
	r := &Registry{templates: map[string]bool{}, assets: map[string]bool{}}
	_ = fs.WalkDir(themeFS, "templates/islands", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".html") {
			return nil
		}
		r.templates[strings.TrimSuffix(path.Base(p), ".html")] = true
		return nil
	})
	_ = fs.WalkDir(themeFS, "assets/islands", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".js") {
			return nil
		}
		r.assets[strings.TrimSuffix(path.Base(p), ".js")] = true
		return nil
	})
	return r
}

// SetRenderer inyecta el renderer de plantillas (tras construirlo).
func (r *Registry) SetRenderer(tr TemplateRenderer) { r.renderer = tr }

// Has informa si existe la isla.
func (r *Registry) Has(name string) bool { return r.templates[name] || r.assets[name] }

// Island devuelve el HTML de una isla: SSR si hay template, si no un placeholder.
func (r *Registry) Island(name string, props map[string]string) (template.HTML, error) {
	if !validName(name) {
		return "", fmt.Errorf("isla inválida: %q", name)
	}
	load := props["load"]
	if load == "" {
		load = "idle"
	}
	clean := make(map[string]string, len(props))
	for k, v := range props {
		if k == "load" {
			continue
		}
		clean[k] = v
	}

	var inner string
	if r.templates[name] {
		if r.renderer == nil {
			return "", fmt.Errorf("isla %q: renderer no inicializado", name)
		}
		var buf strings.Builder
		if err := r.renderer.Render(&buf, "islands/"+name+".html", clean); err != nil {
			return "", fmt.Errorf("isla %q: %w", name, err)
		}
		inner = buf.String()
	}

	if !r.assets[name] {
		if inner == "" {
			return "", fmt.Errorf("isla desconocida %q (sin template ni js)", name)
		}
		return template.HTML(inner), nil // componente estático
	}
	return wrapper(name, clean, load, inner), nil
}

// Partial ejecuta un partial del tema (componente estático sin JS).
func (r *Registry) Partial(name string, props map[string]string) (template.HTML, error) {
	if !validName(name) {
		return "", fmt.Errorf("partial inválido: %q", name)
	}
	if r.renderer == nil {
		return "", fmt.Errorf("partial %q: renderer no inicializado", name)
	}
	var buf strings.Builder
	if err := r.renderer.Render(&buf, "partials/"+name+".html", props); err != nil {
		return "", fmt.Errorf("partial %q: %w", name, err)
	}
	return template.HTML(buf.String()), nil
}

// Resolve es la función que se inyecta en el pipeline de Markdown.
func (r *Registry) Resolve(kind, name string, props map[string]string) (template.HTML, error) {
	if kind == "partial" {
		return r.Partial(name, props)
	}
	return r.Island(name, props)
}

func wrapper(name string, props map[string]string, load, inner string) template.HTML {
	attrs := ` data-island="` + template.HTMLEscapeString(name) + `"`
	if len(props) > 0 {
		if data, err := json.Marshal(props); err == nil {
			attrs += ` data-props="` + template.HTMLEscapeString(string(data)) + `"`
		}
	}
	attrs += ` data-island-load="` + template.HTMLEscapeString(load) + `"`
	return template.HTML(`<div` + attrs + `>` + inner + `</div>`)
}

func validName(name string) bool {
	if name == "" || strings.ContainsAny(name, `/\\. `) {
		return false
	}
	for _, c := range name {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_'
		if !ok {
			return false
		}
	}
	return true
}
