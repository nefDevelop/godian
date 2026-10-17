package application

import (
	"testing"

	"github.com/nefdevelop/godian/internal/domain"
)

func TestLinkIndex(t *testing.T) {
	l := NewLinkIndex()
	posts := []domain.BlogPost{{Slug: "hello", Title: "Hello", Body: "ver [[guide]]"}}
	docs := []domain.Doc{{Slug: "guias/guide", Title: "Guide", Body: "contenido"}}
	l.Build(posts, docs)

	// Wikilink cross-collection resuelto por basename.
	if u, ok := l.Wikilink("guide"); !ok || u != "/docs/guias/guide/" {
		t.Errorf("Wikilink(guide) = %q, %v", u, ok)
	}
	// Backlink desde el post hacia el doc.
	bl := l.Backlinks("guide")
	if len(bl) != 1 || bl[0].URL != "/blog/hello/" {
		t.Errorf("Backlinks(guide) = %+v", bl)
	}
	// Transclusión.
	if html, ok := l.Embed("guide"); !ok || html == "" {
		t.Errorf("Embed(guide) = %q, %v", html, ok)
	}
	// Desconocido.
	if _, ok := l.Wikilink("nope"); ok {
		t.Error("Wikilink(nope) no debería resolver")
	}
}
