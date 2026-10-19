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
	out := string(injectChunks(html, chunks))
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

func TestInjectChunksUnknownIgnored(t *testing.T) {
	html := []byte(`<html><head></head><body><div data-css="missing"></div></body></html>`)
	out := string(injectChunks(html, map[string]string{"lists": "/l.css"}))
	if strings.Contains(out, "<link") {
		t.Errorf("no debería inyectar nada: %s", out)
	}
}
