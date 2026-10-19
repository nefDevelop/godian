package application

import (
	"strings"
	"testing"
)

func TestInjectChunksMultiple(t *testing.T) {
	html := []byte(`<html><head></head><body><div data-css="article-ux lists"></div></body></html>`)
	chunks := map[string]string{
		"article-ux": "/assets/css/chunks/article-ux.css",
		"lists":      "/assets/css/chunks/lists.css",
	}
	out := string(injectChunks(html, chunks, nil))
	for _, want := range []string{
		`<link rel="stylesheet" href="/assets/css/chunks/article-ux.css">`,
		`<link rel="stylesheet" href="/assets/css/chunks/lists.css">`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("salida sin %q: %s", want, out)
		}
	}
	if got := strings.Count(out, "lists.css"); got != 1 {
		t.Errorf("lists.css inyectado %d veces, want 1", got)
	}
}

func TestInjectChunksDependencies(t *testing.T) {
	html := []byte(`<html><head></head><body><div data-css="article-ux"></div></body></html>`)
	chunks := map[string]string{
		"article-ux": "/a.css",
		"lists":      "/l.css",
	}
	deps := map[string][]string{"article-ux": {"lists"}}

	out := string(injectChunks(html, chunks, deps))
	for _, want := range []string{`href="/a.css"`, `href="/l.css"`} {
		if !strings.Contains(out, want) {
			t.Errorf("salida sin %q: %s", want, out)
		}
	}
	if strings.Index(out, "/a.css") > strings.Index(out, "/l.css") {
		t.Errorf("la dependencia debe ir después del chunk que la declara: %s", out)
	}
}

func TestInjectChunksDependenciesCycle(t *testing.T) {
	html := []byte(`<html><head></head><body><div data-css="a"></div></body></html>`)
	chunks := map[string]string{"a": "/a.css", "b": "/b.css"}
	deps := map[string][]string{"a": {"b"}, "b": {"a"}}

	out := string(injectChunks(html, chunks, deps))
	if got := strings.Count(out, "/a.css"); got != 1 {
		t.Errorf("a.css inyectado %d veces, want 1", got)
	}
	if got := strings.Count(out, "/b.css"); got != 1 {
		t.Errorf("b.css inyectado %d veces, want 1", got)
	}
}

func TestInjectChunksUnknownIgnored(t *testing.T) {
	html := []byte(`<html><head></head><body><div data-css="missing"></div></body></html>`)
	out := string(injectChunks(html, map[string]string{"lists": "/l.css"}, nil))
	if strings.Contains(out, "<link") {
		t.Errorf("no debería inyectar nada: %s", out)
	}
}
