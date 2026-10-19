package markdown

import (
	"errors"
	"html/template"
	"strings"
	"testing"

	"github.com/nefdevelop/godian/internal/domain"
)

func TestIslandShortcode(t *testing.T) {
	var kind, name string
	var props map[string]string
	r := New(WithIslandResolver(func(k, n string, p map[string]string) (template.HTML, error) {
		kind, name, props = k, n, p
		return template.HTML("<div>ok</div>"), nil
	}))

	out, err := r.ToHTML("{{island counter start=3 load=visible}}")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<div>ok</div>") {
		t.Errorf("isla no renderizada: %s", out)
	}
	if kind != "island" || name != "counter" || props["start"] != "3" || props["load"] != "visible" {
		t.Errorf("resolver recibió kind=%q name=%q props=%v", kind, name, props)
	}
}

func TestPartialShortcodeQuoted(t *testing.T) {
	var name string
	var props map[string]string
	r := New(WithIslandResolver(func(k, n string, p map[string]string) (template.HTML, error) {
		name, props = n, p
		return template.HTML("<a></a>"), nil
	}))

	if _, err := r.ToHTML(`{{partial cta text="hola mundo" url="/docs/"}}`); err != nil {
		t.Fatal(err)
	}
	if name != "cta" || props["text"] != "hola mundo" || props["url"] != "/docs/" {
		t.Errorf("partial props = %v", props)
	}
}

func TestUnknownIslandErrors(t *testing.T) {
	r := New(WithIslandResolver(func(k, n string, p map[string]string) (template.HTML, error) {
		return "", errors.New("isla desconocida")
	}))
	if _, err := r.ToHTML("{{island nope}}"); err == nil {
		t.Error("se esperaba error de isla desconocida")
	}
}

func render(t *testing.T, src string) string {
	t.Helper()
	out, err := New().ToHTML(src)
	if err != nil {
		t.Fatalf("ToHTML() error = %v", err)
	}
	return string(out)
}

func TestWikilinks(t *testing.T) {
	out := render(t, "Ver [[code-showcase]] y [[code-showcase|el showcase]].")
	if !strings.Contains(out, `<a href="/blog/code-showcase/">code-showcase</a>`) {
		t.Errorf("wikilink simple no generado: %s", out)
	}
	if !strings.Contains(out, `<a href="/blog/code-showcase/">el showcase</a>`) {
		t.Errorf("wikilink con alias no generado: %s", out)
	}
}

func TestWikilinkAnchor(t *testing.T) {
	out := render(t, "Ver [[nota#seccion]].")
	if !strings.Contains(out, `<a href="/blog/nota/#seccion">nota</a>`) {
		t.Errorf("wikilink con ancla no generado: %s", out)
	}
}

func TestHighlight(t *testing.T) {
	out := render(t, "Texto ==resaltado== aquí.")
	if !strings.Contains(out, "<mark>resaltado</mark>") {
		t.Errorf("highlight no generado: %s", out)
	}
}

func TestEmbed(t *testing.T) {
	out := render(t, "![[foto.jpg]]")
	if !strings.Contains(out, `src="/foto.jpg"`) || !strings.Contains(out, `alt="foto.jpg"`) {
		t.Errorf("embed no generado: %s", out)
	}
}

func TestCallouts(t *testing.T) {
	out := render(t, "> [!note] Nota\n> Cuerpo del callout.")
	if !strings.Contains(out, `class="callout callout-note"`) {
		t.Errorf("callout note no generado: %s", out)
	}
	if !strings.Contains(out, `<div class="callout-title">Nota</div>`) {
		t.Errorf("título de callout incorrecto: %s", out)
	}
	if !strings.Contains(out, "Cuerpo del callout.") {
		t.Errorf("cuerpo del callout perdido: %s", out)
	}
}

func TestCalloutFoldable(t *testing.T) {
	out := render(t, "> [!important]- Colapsable\n> Oculto")
	if !strings.Contains(out, "<details class=\"callout callout-important\">") {
		t.Errorf("callout plegable no generado: %s", out)
	}
	if !strings.Contains(out, `<summary class="callout-title">Colapsable</summary>`) {
		t.Errorf("summary de callout plegable incorrecto: %s", out)
	}
}

func TestTableClass(t *testing.T) {
	src := "| a | b |\n|---|---|\n| 1 | 2 |\n\n{.sortable}\n"
	out := render(t, src)
	if !strings.Contains(out, `<table class="sortable">`) {
		t.Errorf("clase de tabla no aplicada: %s", out)
	}
	if strings.Contains(out, "{.sortable}") {
		t.Errorf("el párrafo de atributo no se eliminó: %s", out)
	}
}

func TestFigure(t *testing.T) {
	out := render(t, "![Un paisaje](/img/x.jpg)")
	if !strings.Contains(out, `<figure class="image-figure">`) {
		t.Errorf("figure no generada: %s", out)
	}
	if !strings.Contains(out, "<figcaption>Un paisaje</figcaption>") {
		t.Errorf("figcaption no generada: %s", out)
	}
}

func TestRawHTMLPreserved(t *testing.T) {
	out := render(t, "<details>\n<summary>Ver</summary>\n\nContenido\n\n</details>")
	if !strings.Contains(out, "<details>") || !strings.Contains(out, "<summary>Ver</summary>") {
		t.Errorf("HTML crudo no preservado: %s", out)
	}
}

func TestShortcode(t *testing.T) {
	out := render(t, "{{youtube dQw4w9WgXcQ}}")
	if !strings.Contains(out, `class="video-embed"`) || !strings.Contains(out, `data-youtube-id="dQw4w9WgXcQ"`) {
		t.Errorf("shortcode youtube no generado: %s", out)
	}
}

func TestMermaid(t *testing.T) {
	out := render(t, "```mermaid\ngraph LR\n  A-->B\n```")
	if !strings.Contains(out, `class="mermaid"`) {
		t.Errorf("mermaid no generado: %s", out)
	}
	if strings.Contains(out, "chroma") {
		t.Errorf("mermaid no debe pasar por Chroma: %s", out)
	}
}

func TestCodeHighlight(t *testing.T) {
	out := render(t, "```go {2}\nfunc main() {}\n```")
	if !strings.Contains(out, `class="code-block"`) || !strings.Contains(out, `class="chroma"`) {
		t.Errorf("bloque de código no resaltado: %s", out)
	}
	if !strings.Contains(out, `class="ln"`) {
		t.Errorf("line numbers no generados: %s", out)
	}
}

func TestBlockShortcode(t *testing.T) {
	var content string
	r := New(WithBlockResolver(func(kind, name string, props map[string]string, c template.HTML) (template.HTML, error) {
		content = string(c)
		return template.HTML("<section>" + string(c) + "</section>"), nil
	}))
	out, err := r.ToHTML("{{#partial note title=\"A\"}}\nHola **mundo**\n{{/partial}}")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<section>") || !strings.Contains(content, "<strong>mundo</strong>") {
		t.Errorf("bloque no renderizado: out=%s content=%s", out, content)
	}
	if strings.Contains(string(out), "data-godian-block") {
		t.Errorf("placeholder sin reemplazar: %s", out)
	}
}

func TestBlockInsideFenceIgnored(t *testing.T) {
	called := false
	r := New(WithBlockResolver(func(kind, name string, props map[string]string, c template.HTML) (template.HTML, error) {
		called = true
		return c, nil
	}))
	out, err := r.ToHTML("```\n{{#partial note}}\nx\n{{/partial}}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("no debe procesar bloques dentro de fences de código")
	}
	if !strings.Contains(string(out), "{{#partial note}}") {
		t.Errorf("se esperaba el literal: %s", out)
	}
}

func TestExternalLinkHook(t *testing.T) {
	ext := render(t, "[go](https://go.dev)")
	if !strings.Contains(ext, `target="_blank" rel="noopener noreferrer"`) {
		t.Errorf("enlace externo sin target/rel: %s", ext)
	}
	internal := render(t, "[docs](/docs/)")
	if strings.Contains(internal, "target=") {
		t.Errorf("enlace interno no debe llevar target: %s", internal)
	}
}

func TestUnknownShortcodeLiteral(t *testing.T) {
	out := render(t, "{{desconocido arg}}")
	if !strings.Contains(out, "{{desconocido arg}}") {
		t.Errorf("shortcode desconocido no se dejó literal: %s", out)
	}
}

func renderCfg(t *testing.T, cfg domain.RenderConfig, src string) string {
	t.Helper()
	out, err := New(WithRenderConfig(cfg)).ToHTML(src)
	if err != nil {
		t.Fatalf("ToHTML() error = %v", err)
	}
	return string(out)
}

func TestRenderConfigHeadingAnchorsOff(t *testing.T) {
	cfg := domain.DefaultRenderConfig()
	cfg.Headings.Anchors = false
	out := renderCfg(t, cfg, "# Hola")
	if strings.Contains(out, `class="anchor"`) {
		t.Errorf("no debería haber ancla: %s", out)
	}
	if !strings.Contains(out, `<h1 id="hola">`) {
		t.Errorf("el encabezado debería conservar el id: %s", out)
	}
}

func TestRenderConfigHeadingAnchorLabel(t *testing.T) {
	cfg := domain.DefaultRenderConfig()
	cfg.Headings.Anchor = "¶"
	out := renderCfg(t, cfg, "# Hola")
	// El símbolo va en un <span aria-hidden> con nombre accesible en el <a>.
	if !strings.Contains(out, `>¶</span></a>`) || !strings.Contains(out, `aria-label="`) {
		t.Errorf("etiqueta de ancla personalizada no aplicada: %s", out)
	}
}

func TestRenderConfigExternalLinkDisabled(t *testing.T) {
	cfg := domain.DefaultRenderConfig()
	cfg.ExternalLinks.Target = ""
	cfg.ExternalLinks.Rel = ""
	out := renderCfg(t, cfg, "[go](https://go.dev)")
	if strings.Contains(out, "target=") || strings.Contains(out, "rel=") {
		t.Errorf("enlace externo no debería llevar target/rel: %s", out)
	}
}

func TestRenderConfigImageAttrsAndFigureOff(t *testing.T) {
	cfg := domain.DefaultRenderConfig()
	cfg.Images.Loading = "eager"
	cfg.Images.Decoding = ""
	cfg.Images.Figure = false
	out := renderCfg(t, cfg, "![Un paisaje](/img/x.jpg)")
	if !strings.Contains(out, `loading="eager"`) {
		t.Errorf("loading configurable no aplicado: %s", out)
	}
	if strings.Contains(out, "decoding=") {
		t.Errorf("decoding debería omitirse: %s", out)
	}
	if strings.Contains(out, "<figure") {
		t.Errorf("figure desactivada pero presente: %s", out)
	}
}
