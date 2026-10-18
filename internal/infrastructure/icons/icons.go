// Package icons resuelve iconos SVG para plantillas y Markdown.
//
// El catálogo por defecto es Lucide (embebido en lucide.json, licencia ISC en
// LICENSE). Un tema puede añadir sus propios iconos en icons/*.svg, que
// sobrescriben el catálogo por nombre.
package icons

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"
)

// lucideVersion es la versión vendored del catálogo (ver LICENSE).
const lucideVersion = "1.59.0"

//go:embed lucide.json
var lucideJSON []byte

// svgNode es un elemento del icono (path, circle, line…) con sus atributos.
type svgNode struct {
	tag   string
	attrs map[string]string
}

// iconDef es un icono resuelto: nodos estructurados (Lucide) o markup interno
// literal (iconos del tema, tal cual vienen en el .svg).
type iconDef struct {
	nodes []svgNode
	raw   string
}

// Registry conoce los iconos del core (Lucide) y los extra de un tema.
type Registry struct {
	core    map[string]iconDef
	extra   map[string]iconDef
	warn    func(string)
	mu      sync.Mutex
	missing map[string]bool
}

// New carga el catálogo Lucide embebido.
func New() (*Registry, error) {
	var raw map[string][][]any
	if err := json.Unmarshal(lucideJSON, &raw); err != nil {
		return nil, fmt.Errorf("iconos: catálogo Lucide v%s inválido: %w", lucideVersion, err)
	}
	core := make(map[string]iconDef, len(raw))
	for name, def := range raw {
		nodes := make([]svgNode, 0, len(def))
		for _, el := range def {
			if len(el) != 2 {
				continue
			}
			tag, _ := el[0].(string)
			attrs, _ := el[1].(map[string]any)
			if tag == "" {
				continue
			}
			a := make(map[string]string, len(attrs))
			for k, v := range attrs {
				a[k] = fmt.Sprint(v)
			}
			nodes = append(nodes, svgNode{tag: tag, attrs: a})
		}
		core[name] = iconDef{nodes: nodes}
	}
	if len(core) == 0 {
		return nil, errors.New("iconos: el catálogo Lucide quedó vacío")
	}
	return &Registry{
		core:    core,
		extra:   map[string]iconDef{},
		missing: map[string]bool{},
	}, nil
}

// SetWarning registra el canal de avisos (iconos desconocidos). Opcional.
func (r *Registry) SetWarning(fn func(string)) { r.warn = fn }

// LoadDir carga los iconos propios del tema desde los *.svg de dir. Los del
// tema tienen prioridad sobre el catálogo Lucide. Si dir no existe, no pasa nada.
func (r *Registry) LoadDir(fsys fs.FS, dir string) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("iconos del tema: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".svg" {
			continue
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, e.Name()))
		if err != nil {
			return fmt.Errorf("iconos del tema: %w", err)
		}
		inner, err := innerSVG(string(data))
		if err != nil {
			return fmt.Errorf("icono %q: %w", e.Name(), err)
		}
		r.extra[strings.TrimSuffix(e.Name(), ".svg")] = iconDef{raw: inner}
	}
	return nil
}

// Has informa si el icono existe (catálogo o tema).
func (r *Registry) Has(name string) bool {
	if _, ok := r.extra[name]; ok {
		return true
	}
	_, ok := r.core[name]
	return ok
}

// HTML devuelve el <svg> del icono. Si label está vacío el icono es decorativo
// (aria-hidden); si no, es significativo (role="img" + <title>). Un nombre
// desconocido devuelve un placeholder visible y dispara un aviso, sin romper el build.
func (r *Registry) HTML(name, label string) (template.HTML, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("icono: falta el nombre")
	}
	def, ok := r.extra[name]
	if !ok {
		def, ok = r.core[name]
	}
	if !ok {
		r.markMissing(name)
		return template.HTML(placeholder(name)), nil
	}
	return template.HTML(render(name, def, label)), nil
}

// markMissing avisa una sola vez por nombre desconocido.
func (r *Registry) markMissing(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.missing[name] {
		return
	}
	r.missing[name] = true
	if r.warn != nil {
		r.warn(fmt.Sprintf("icono desconocido %q (¿está en Lucide o en icons/ del tema?)", name))
	}
}

// render compone el <svg> de forma determinista (atributos ordenados).
func render(name string, def iconDef, label string) string {
	var b strings.Builder
	b.WriteString(`<svg class="icon icon-`)
	b.WriteString(classToken(name))
	b.WriteString(`" xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"`)
	b.WriteString(` fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"`)
	if label != "" {
		b.WriteString(` role="img" aria-label="`)
		b.WriteString(html.EscapeString(label))
		b.WriteString(`"`)
	} else {
		b.WriteString(` aria-hidden="true"`)
	}
	b.WriteString(`>`)
	if label != "" {
		b.WriteString("<title>")
		b.WriteString(html.EscapeString(label))
		b.WriteString("</title>")
	}
	if def.raw != "" {
		b.WriteString(def.raw)
	} else {
		for _, n := range def.nodes {
			b.WriteString("<")
			b.WriteString(n.tag)
			keys := make([]string, 0, len(n.attrs))
			for k := range n.attrs {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				b.WriteString(" ")
				b.WriteString(k)
				b.WriteString(`="`)
				b.WriteString(html.EscapeString(n.attrs[k]))
				b.WriteString(`"`)
			}
			b.WriteString("/>")
		}
	}
	b.WriteString("</svg>")
	return b.String()
}

// placeholder sustituye a un icono desconocido, visible y localizable.
func placeholder(name string) string {
	safe := html.EscapeString(name)
	return `<span class="icon-missing" title="icono desconocido: ` + safe + `">` + safe + `</span>`
}

// innerSVG extrae el contenido interno de un .svg (sin la etiqueta <svg>),
// que es lo único que se reutiliza dentro de nuestro wrapper.
func innerSVG(s string) (string, error) {
	start := strings.Index(s, "<svg")
	if start < 0 {
		return "", errors.New("sin etiqueta <svg>")
	}
	gt := strings.IndexByte(s[start:], '>')
	if gt < 0 {
		return "", errors.New("etiqueta <svg> sin cerrar")
	}
	end := strings.LastIndex(s, "</svg>")
	if end < start+gt {
		return "", errors.New("sin </svg>")
	}
	return strings.TrimSpace(s[start+gt+1 : end]), nil
}

// classToken deja el nombre del icono en algo seguro para una clase CSS.
func classToken(name string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, name)
}
