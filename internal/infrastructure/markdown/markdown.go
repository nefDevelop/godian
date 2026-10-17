// Package markdown convierte Markdown a HTML con extensiones propias:
// wikilinks y embeds estilo Obsidian, ==highlight==, callouts, atributos,
// figuras y shortcodes.
package markdown

import (
	"bytes"
	"html/template"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Shortcode genera HTML a partir de sus argumentos.
type Shortcode func(args []string) (template.HTML, error)

// Option configura el Renderer.
type Option func(*Renderer)

// WithShortcode registra un shortcode con nombre.
func WithShortcode(name string, fn Shortcode) Option {
	return func(r *Renderer) { r.shortcodes[name] = fn }
}

// WithImageResolver sustituye el render de las imágenes (p. ej. por <picture>).
func WithImageResolver(fn func(src, alt string) template.HTML) Option {
	return func(r *Renderer) { r.imageResolver = fn }
}

// WithIslandResolver resuelve los shortcodes {{island}} y {{partial}}.
func WithIslandResolver(fn func(kind, name string, props map[string]string) (template.HTML, error)) Option {
	return func(r *Renderer) { r.islandResolver = fn }
}

// Renderer convierte documentos Markdown en HTML.
type Renderer struct {
	md             goldmark.Markdown
	shortcodes     map[string]Shortcode
	imageResolver  func(src, alt string) template.HTML
	islandResolver func(kind, name string, props map[string]string) (template.HTML, error)
}

// New construye un Renderer con las extensiones del tema.
func New(opts ...Option) *Renderer {
	r := &Renderer{shortcodes: map[string]Shortcode{}}
	for _, o := range opts {
		o(r)
	}
	r.shortcodes["youtube"] = youtubeShortcode

	r.md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithInlineParsers(
				// Prioridad menor que el parser de enlaces (200) para que
				// [[wikilinks]] y [!callouts] no se interpreten como enlaces.
				util.Prioritized(&obsidianParser{}, 150),
				util.Prioritized(&shortcodeParser{}, 150),
			),
			parser.WithASTTransformers(
				util.Prioritized(&attributeTransformer{}, 100),
				util.Prioritized(&figureTransformer{}, 200),
				util.Prioritized(&calloutTransformer{}, 300),
			),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			// Prioridad menor que el renderer por defecto (1000): goldmark
			// registra en orden inverso, así que gana el número más bajo.
			renderer.WithNodeRenderers(util.Prioritized(&customRenderer{shortcodes: r.shortcodes, images: r.imageResolver, islands: r.islandResolver}, 500)),
		),
	)
	return r
}

// Render convierte Markdown a HTML y extrae los encabezados para el TOC.
func (r *Renderer) Render(src string) (template.HTML, []domain.Heading, error) {
	source := []byte(src)
	doc := r.md.Parser().Parse(text.NewReader(source))

	var buf bytes.Buffer
	if err := r.md.Renderer().Render(&buf, source, doc); err != nil {
		return "", nil, err
	}
	return template.HTML(buf.String()), collectHeadings(doc, source), nil
}

// ToHTML renderiza el Markdown de entrada a HTML (sin TOC).
func (r *Renderer) ToHTML(src string) (template.HTML, error) {
	html, _, err := r.Render(src)
	return html, err
}

func collectHeadings(doc ast.Node, source []byte) []domain.Heading {
	var out []domain.Heading
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		id := ""
		if v, ok := h.AttributeString("id"); ok {
			switch t := v.(type) {
			case []byte:
				id = string(t)
			case string:
				id = t
			}
		}
		out = append(out, domain.Heading{Level: h.Level, ID: id, Text: nodeText(h, source)})
		return ast.WalkContinue, nil
	})
	return out
}

// youtubeShortcode embebe un iframe responsivo de YouTube.
func youtubeShortcode(args []string) (template.HTML, error) {
	if len(args) == 0 {
		return "", nil
	}
	id := template.HTMLEscapeString(args[0])
	return template.HTML(`<div class="video-embed"><iframe src="https://www.youtube.com/embed/` + id +
		`" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen loading="lazy"></iframe></div>`), nil
}
