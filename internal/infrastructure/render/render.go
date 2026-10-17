// Package render ejecuta las plantillas del tema (html/template).
package render

import (
	"html/template"
	"io"
	"io/fs"
	"strings"
)

// Renderer mantiene el conjunto de plantillas del tema.
type Renderer struct {
	tmpl *template.Template
}

// New carga todas las plantillas .html del sistema de archivos indicado.
// funcs permite inyectar funciones (p. ej. el traductor "t").
func New(templateFS fs.FS, funcs template.FuncMap) (*Renderer, error) {
	t := template.New("").Funcs(funcs)
	err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		data, err := fs.ReadFile(templateFS, path)
		if err != nil {
			return err
		}
		_, err = t.New(path).Parse(string(data))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: t}, nil
}

// Render ejecuta la plantilla indicada escribiendo el resultado en w.
func (r *Renderer) Render(w io.Writer, name string, data any) error {
	return r.tmpl.ExecuteTemplate(w, name, data)
}
