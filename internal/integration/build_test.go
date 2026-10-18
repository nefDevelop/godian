// Package integration prueba el pipeline completo (build + verificación).
package integration

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nefdevelop/godian/internal/application"
	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/i18n"
	"github.com/nefdevelop/godian/internal/infrastructure/content"
	"github.com/nefdevelop/godian/internal/infrastructure/feed"
	"github.com/nefdevelop/godian/internal/infrastructure/icons"
	"github.com/nefdevelop/godian/internal/infrastructure/islands"
	"github.com/nefdevelop/godian/internal/infrastructure/markdown"
	"github.com/nefdevelop/godian/internal/infrastructure/output"
	"github.com/nefdevelop/godian/internal/infrastructure/render"
	"github.com/nefdevelop/godian/internal/infrastructure/theme"
	"github.com/nefdevelop/godian/internal/infrastructure/verify"
)

func TestBuildEndToEnd(t *testing.T) {
	out := t.TempDir()

	cfg := domain.DefaultConfig()
	cfg.Site = domain.SiteConfig{Title: "Test", URL: "https://example.com", Language: "es"}
	cfg.Build.ContentDir = "testdata/content"
	cfg.Build.OutputDir = out
	cfg.Build.AssetsDir = filepath.Join(out, "no-assets") // no existe: se ignora
	cfg.Collections = domain.DefaultCollections()
	cfg.Taxonomies = domain.DefaultTaxonomies()

	th, err := theme.Load("")
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	tmplFS, err := th.TemplateFS()
	if err != nil {
		t.Fatal(err)
	}
	reg := islands.New(th.FS())
	iconReg, err := icons.New()
	if err != nil {
		t.Fatalf("icons.New: %v", err)
	}
	if err := iconReg.LoadDir(th.FS(), "icons"); err != nil {
		t.Fatalf("icons.LoadDir: %v", err)
	}
	funcs := template.FuncMap{
		"t":    func(k string) string { return i18n.T("es", k) },
		"dict": dictFunc,
		"icon": func(name string, label ...string) (template.HTML, error) {
			return iconReg.HTML(name, firstString(label))
		},
		"island": func(name string, props ...map[string]string) (template.HTML, error) {
			return reg.Island(name, firstMap(props))
		},
		"partial": func(name string, props ...map[string]string) (template.HTML, error) {
			return reg.Partial(name, firstMap(props))
		},
	}
	renderer, err := render.New(tmplFS, funcs)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}
	reg.SetRenderer(renderer)
	assetsFS, err := th.AssetsFS()
	if err != nil {
		t.Fatal(err)
	}

	builder := application.NewBuild(
		content.NewLoader(cfg),
		markdown.New(markdown.WithIslandResolver(reg.Resolve), markdown.WithBlockResolver(reg.ResolveBlock)),
		renderer,
		output.NewWriter(out),
		feed.New(),
		application.Options{
			Site:        cfg.Site,
			Taxonomies:  cfg.Taxonomies,
			CSS:         th.Manifest.Assets.CSS,
			JS:          th.Manifest.Assets.JS,
			AssetsFS:    assetsFS,
			Minify:      true,
			Fingerprint: true,
		},
	)

	summary, err := builder.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.Published != 7 {
		t.Errorf("posts publicados = %d, want 7", summary.Published)
	}
	if summary.Scheduled != 1 {
		t.Errorf("posts programados = %d, want 1", summary.Scheduled)
	}
	if _, err := os.Stat(filepath.Join(out, "blog/futuro/index.html")); err == nil {
		t.Error("el post futuro no debería generarse")
	}

	for _, f := range []string{
		"index.html",
		"blog/index.html",
		"blog/2/index.html",
		"blog/hello/index.html",
		"blog/adios/index.html",
		"tags/index.html",
		"tags/go/index.html",
		"tags/go/rss.xml",
		"categories/dev/index.html",
		"authors/nef/index.html",
		"search/index.html",
		"search-index.json",
		"rss.xml",
		"atom.xml",
		"feed.json",
		"robots.txt",
		"sitemap.xml",
		"404.html",
		"docs/index.html",
		"about/index.html",
		"assets/islands/manifest.json",
		"assets/vendor/mermaid/mermaid.esm.min.mjs",
		"assets/vendor/katex/katex.min.css",
		"blog/viejo-hola/index.html",
	} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(f))); err != nil {
			t.Errorf("falta %s: %v", f, err)
		}
	}

	hello := readFile(t, filepath.Join(out, "blog/hello/index.html"))
	for _, want := range []string{"Hola mundo", "<mark>resaltado</mark>", `class="callout callout-note"`, `href="/blog/adios/"`, `class="code-block"`, `data-island="counter"`, `class="callout`, "Aviso"} {
		if !strings.Contains(hello, want) {
			t.Errorf("blog/hello no contiene %q", want)
		}
	}

	blog := readFile(t, filepath.Join(out, "blog/index.html"))
	if !strings.Contains(blog, `rel="next"`) {
		t.Errorf("blog/index.html sin rel=next (paginación)")
	}

	about := readFile(t, filepath.Join(out, "about/index.html"))
	for _, want := range []string{"Acerca de", `class="page-stack"`, "goldmark"} {
		if !strings.Contains(about, want) {
			t.Errorf("about/index.html no contiene %q", want)
		}
	}

	if problems, err := verify.Dist(out); err != nil {
		t.Fatalf("verify: %v", err)
	} else if len(problems) > 0 {
		t.Errorf("referencias rotas: %v", problems)
	}
}

func firstMap(maps []map[string]string) map[string]string {
	if len(maps) > 0 && maps[0] != nil {
		return maps[0]
	}
	return map[string]string{}
}

func firstString(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func dictFunc(values ...any) (map[string]string, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict: número impar de argumentos")
	}
	m := make(map[string]string, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		k, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: la clave debe ser string")
		}
		m[k] = fmt.Sprintf("%v", values[i+1])
	}
	return m, nil
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leyendo %s: %v", path, err)
	}
	return string(data)
}
