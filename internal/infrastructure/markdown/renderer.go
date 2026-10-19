package markdown

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type customRenderer struct {
	shortcodes      map[string]Shortcode
	images          func(src, alt string) template.HTML
	islands         func(kind, name string, props map[string]string) (template.HTML, error)
	imageShortcodes func(args map[string]string) (template.HTML, error)
	wikilinks       func(target string) (string, bool)
	embeds          func(target string) (template.HTML, bool)
	tr              func(key string) string
	render          domain.RenderConfig
}

func (r *customRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindWiki, r.renderWiki)
	reg.Register(kindEmbed, r.renderEmbed)
	reg.Register(kindHighlight, r.renderHighlight)
	reg.Register(kindShortcode, r.renderShortcode)
	reg.Register(kindCallout, r.renderCallout)
	reg.Register(kindFigure, r.renderFigure)
	reg.Register(kindCalloutMarker, r.renderMarker)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCode)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindLink, r.renderLink)
}

// renderLink aplica a los enlaces externos el target/rel configurados y avisa
// (para lectores de pantalla) de que abren en ventana nueva.
func (r *customRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	href := string(n.Destination)
	external := isExternal(href)
	if entering {
		_, _ = w.WriteString(`<a href="` + escape(href) + `"`)
		if external {
			if target := r.render.ExternalLinks.Target; target != "" {
				_, _ = w.WriteString(` target="` + escape(target) + `"`)
			}
			if rel := r.render.ExternalLinks.Rel; rel != "" {
				_, _ = w.WriteString(` rel="` + escape(rel) + `"`)
			}
		}
		if len(n.Title) > 0 {
			_, _ = w.WriteString(` title="` + escape(string(n.Title)) + `"`)
		}
		_, _ = w.WriteString(`>`)
	} else {
		if external && r.render.ExternalLinks.Target == "_blank" {
			hint := "opens in a new window"
			if r.tr != nil {
				hint = r.tr("ui.new_window")
			}
			_, _ = w.WriteString(`<span class="sr-only"> (` + escape(hint) + `)</span>`)
		}
		_, _ = w.WriteString(`</a>`)
	}
	return ast.WalkContinue, nil
}

func isExternal(href string) bool {
	return strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "//")
}

// renderHeading añade (según config) un ancla a cada encabezado.
func (r *customRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	id := attrString(n, "id")
	if entering {
		_, _ = w.WriteString("<h")
		_, _ = w.WriteString(strconv.Itoa(n.Level))
		if id != "" {
			_, _ = w.WriteString(` id="` + escape(id) + `"`)
		}
		_, _ = w.WriteString(">")
	} else {
		if id != "" && r.render.Headings.Anchors && r.render.Headings.Anchor != "" {
			// Enlace de ancla con nombre accesible (el símbolo es decorativo).
			label := "Enlace a esta sección"
			if r.tr != nil {
				label = r.tr("ui.anchor")
			}
			_, _ = w.WriteString(`<a class="anchor" href="#` + escape(id) + `" aria-label="` + escape(label) + `" tabindex="-1"><span aria-hidden="true">` + escape(r.render.Headings.Anchor) + `</span></a>`)
		}
		_, _ = w.WriteString("</h")
		_, _ = w.WriteString(strconv.Itoa(n.Level))
		_, _ = w.WriteString(">")
	}
	return ast.WalkContinue, nil
}

func attrString(n ast.Node, name string) string {
	v, ok := n.AttributeString(name)
	if !ok {
		return ""
	}
	switch t := v.(type) {
	case []byte:
		return string(t)
	case string:
		return t
	}
	return ""
}

func (r *customRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	src := string(n.Destination)
	alt := nodeText(n, source)
	if r.images != nil {
		if out := r.images(src, alt); out != "" {
			_, _ = w.WriteString(string(out))
			return ast.WalkContinue, nil
		}
	}
	_, _ = w.WriteString(`<img src="` + escape(src) + `" alt="` + escape(alt) + `"` + r.imgAttrs() + `>`)
	return ast.WalkContinue, nil
}

// imgAttrs devuelve los atributos de carga de las imágenes según config.
func (r *customRenderer) imgAttrs() string {
	var b strings.Builder
	if r.render.Images.Loading != "" {
		b.WriteString(` loading="` + escape(r.render.Images.Loading) + `"`)
	}
	if r.render.Images.Decoding != "" {
		b.WriteString(` decoding="` + escape(r.render.Images.Decoding) + `"`)
	}
	return b.String()
}

// renderMarker solo se alcanza si el marcador no se convirtió en callout.
func (r *customRenderer) renderMarker(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*calloutMarkerNode)
	_, _ = w.Write(util.EscapeHTML([]byte(n.Raw)))
	return ast.WalkContinue, nil
}

func (r *customRenderer) renderWiki(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*wikiNode)
	label := n.Label
	if label == "" || label == n.Target {
		label = n.Target
		if i := strings.Index(label, "#"); i >= 0 {
			label = label[:i]
		}
	}
	target, anchor := n.Target, ""
	if i := strings.Index(target, "#"); i >= 0 {
		anchor = target[i+1:]
		target = target[:i]
	}
	href := "/blog/" + slugify(target) + "/"
	if anchor != "" {
		href += "#" + slugify(anchor)
	}
	if r.wikilinks != nil {
		if u, ok := r.wikilinks(n.Target); ok {
			href = u
		}
	}
	_, _ = w.WriteString(`<a href="` + href + `">`)
	_, _ = w.Write(util.EscapeHTML([]byte(label)))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

func (r *customRenderer) renderEmbed(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*embedNode)
	// Transclusión de nota si el resolver la reconoce.
	if r.embeds != nil {
		if html, ok := r.embeds(n.Target); ok {
			_, _ = w.WriteString(string(html))
			return ast.WalkContinue, nil
		}
	}
	// Si no, imagen (vía el resolver de imágenes si está disponible).
	if r.images != nil {
		_, _ = w.WriteString(string(r.images("/"+n.Target, n.Target)))
		return ast.WalkContinue, nil
	}
	esc := escape(n.Target)
	_, _ = w.WriteString(`<img class="embed" src="/` + esc + `" alt="` + esc + `"` + r.imgAttrs() + `>`)
	return ast.WalkContinue, nil
}

func (r *customRenderer) renderHighlight(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*highlightNode)
	_, _ = w.WriteString(`<mark>`)
	_, _ = w.Write(util.EscapeHTML([]byte(n.Value)))
	_, _ = w.WriteString(`</mark>`)
	return ast.WalkContinue, nil
}

func (r *customRenderer) renderShortcode(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*shortcodeNode)

	// Imagen con preset: {{image src=… preset=… alt=…}}.
	if n.Name == "image" && r.imageShortcodes != nil {
		out, err := r.imageShortcodes(parseProps(tokenize(n.Inner)[1:]))
		if err != nil {
			return ast.WalkContinue, err
		}
		_, _ = w.WriteString(string(out))
		return ast.WalkContinue, nil
	}

	// Islas y componentes: {{island nombre k=v}} / {{partial nombre k=v}}.
	if (n.Name == "island" || n.Name == "partial") && r.islands != nil {
		tokens := tokenize(n.Inner)
		if len(tokens) < 2 {
			return ast.WalkContinue, fmt.Errorf("shortcode %q: falta el nombre", n.Name)
		}
		out, err := r.islands(n.Name, tokens[1], parseProps(tokens[2:]))
		if err != nil {
			return ast.WalkContinue, err
		}
		_, _ = w.WriteString(string(out))
		return ast.WalkContinue, nil
	}

	fn, ok := r.shortcodes[n.Name]
	if !ok {
		_, _ = w.Write(util.EscapeHTML([]byte(n.Raw)))
		return ast.WalkContinue, nil
	}
	out, err := fn(n.Args)
	if err != nil {
		_, _ = w.Write(util.EscapeHTML([]byte(n.Raw)))
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(string(out))
	return ast.WalkContinue, nil
}

func (r *customRenderer) renderCallout(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*calloutNode)
	if entering {
		title := `<div class="callout-title">` + escape(n.Title) + `</div>`
		if n.Foldable {
			open := ""
			if !n.Collapsed {
				open = " open"
			}
			_, _ = w.WriteString(`<details class="callout callout-` + n.Variant + `"` + open + `>`)
			_, _ = w.WriteString(`<summary class="callout-title">` + escape(n.Title) + `</summary>`)
		} else {
			_, _ = w.WriteString(`<div class="callout callout-` + n.Variant + `">`)
			_, _ = w.WriteString(title)
		}
		_, _ = w.WriteString(`<div class="callout-content">`)
	} else {
		_, _ = w.WriteString(`</div>`)
		if n.Foldable {
			_, _ = w.WriteString(`</details>`)
		} else {
			_, _ = w.WriteString(`</div>`)
		}
	}
	return ast.WalkContinue, nil
}

func (r *customRenderer) renderFigure(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*figureNode)
	if entering {
		_, _ = w.WriteString(`<figure class="image-figure">`)
	} else {
		if n.Alt != "" {
			_, _ = w.WriteString(`<figcaption>` + escape(n.Alt) + `</figcaption>`)
		}
		_, _ = w.WriteString(`</figure>`)
	}
	return ast.WalkContinue, nil
}

func escape(s string) string {
	return string(util.EscapeHTML([]byte(s)))
}

func slugify(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), "-")
}
