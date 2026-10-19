package application

import (
	"html/template"
	"regexp"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
)

// noteRef es una nota indexada para wikilinks/transclusión.
type noteRef struct {
	URL   string
	Title string
	Body  string
}

// LinkIndex resuelve wikilinks entre colecciones, transclusiones y backlinks.
type LinkIndex struct {
	notes     map[string]noteRef
	aliases   map[string]string
	backlinks map[string][]PostNav
	renderer  MarkdownRenderer
}

// NewLinkIndex crea un índice vacío.
func NewLinkIndex() *LinkIndex {
	return &LinkIndex{
		notes:     map[string]noteRef{},
		aliases:   map[string]string{},
		backlinks: map[string][]PostNav{},
	}
}

// SetRenderer inyecta el renderer usado para transcluir notas.
func (l *LinkIndex) SetRenderer(r MarkdownRenderer) { l.renderer = r }

func (l *LinkIndex) add(key, url, title, body string, aliases []string) {
	if key == "" {
		return
	}
	if _, exists := l.notes[key]; exists {
		return
	}
	l.notes[key] = noteRef{URL: url, Title: title, Body: body}
	for _, a := range aliases {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if _, ok := l.aliases[a]; !ok {
			l.aliases[a] = key
		}
	}
}

func (l *LinkIndex) resolve(target string) (string, noteRef, bool) {
	target = strings.TrimSpace(target)
	if r, ok := l.notes[target]; ok {
		return target, r, true
	}
	if k, ok := l.aliases[target]; ok {
		return k, l.notes[k], true
	}
	return "", noteRef{}, false
}

// Wikilink resuelve [[target]] o [[target#ancla]] a una URL.
func (l *LinkIndex) Wikilink(target string) (string, bool) {
	base, anchor := splitAnchor(target)
	if base == "" && anchor != "" {
		return "#" + anchorSlug(anchor), true
	}
	if _, r, ok := l.resolve(base); ok {
		if anchor != "" {
			return r.URL + "#" + anchorSlug(anchor), true
		}
		return r.URL, true
	}
	return "", false
}

// Embed resuelve ![[target]] como transclusión de nota (o false si no es nota).
func (l *LinkIndex) Embed(target string) (template.HTML, bool) {
	base, _ := splitAnchor(target)
	if _, r, ok := l.resolve(base); ok {
		var body string
		if l.renderer != nil {
			if out, _, err := l.renderer.Render(r.Body); err == nil {
				body = string(out)
			}
		}
		return template.HTML(`<div class="transclusion">` + body + `</div>`), true
	}
	return "", false
}

// Backlinks devuelve las notas que enlazan a target.
func (l *LinkIndex) Backlinks(target string) []PostNav {
	if k, _, ok := l.resolve(target); ok {
		return l.backlinks[k]
	}
	return nil
}

var wikilinkScanRe = regexp.MustCompile(`!?\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)

// Build indexa posts y docs y calcula backlinks.
func (l *LinkIndex) Build(posts []domain.BlogPost, docs []domain.Doc) {
	for _, p := range posts {
		l.add(p.Slug, "/blog/"+p.Slug+"/", p.Title, p.Body, []string{baseName(p.Slug)})
	}
	for _, d := range docs {
		l.add(d.Slug, "/docs/"+d.Slug+"/", d.Title, d.Body, []string{baseName(d.Slug)})
	}
	for key, ref := range l.notes {
		for _, m := range wikilinkScanRe.FindAllStringSubmatch(ref.Body, -1) {
			base, _ := splitAnchor(m[1])
			if tk, _, ok := l.resolve(base); ok && tk != key {
				l.backlinks[tk] = append(l.backlinks[tk], PostNav{Title: ref.Title, URL: ref.URL})
			}
		}
	}
}

// splitAnchor separa "nota#seccion" en ("nota", "seccion").
func splitAnchor(target string) (string, string) {
	target = strings.TrimSpace(target)
	if i := strings.Index(target, "#"); i >= 0 {
		return strings.TrimSpace(target[:i]), strings.TrimSpace(target[i+1:])
	}
	return target, ""
}

// anchorSlug normaliza el texto del ancla al id de encabezado (minúsculas y guiones).
func anchorSlug(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), "-")
}

func baseName(slug string) string {
	if i := strings.LastIndex(slug, "/"); i >= 0 {
		return slug[i+1:]
	}
	return slug
}
