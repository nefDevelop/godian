package application

import (
	"strings"
	"testing"
	"time"

	"github.com/nefdevelop/godian/internal/domain"
)

func TestNeedsMath(t *testing.T) {
	if !needsMath("Ecuación: $e^{i\\pi} + 1 = 0$") {
		t.Error("needsMath() no detectó $...$")
	}
	if !needsMath("$$\n\\sum x\n$$") {
		t.Error("needsMath() no detectó $$...$$")
	}
	if needsMath("Sin matemáticas aquí.") {
		t.Error("needsMath() falso positivo")
	}
}

func TestPlainText(t *testing.T) {
	got := plainText("# Título\n\nTexto **negrita** con [enlace](http://x) y `code`.\n\n```go\nfmt.Println()\n```")
	for _, bad := range []string{"#", "**", "```", "http://x"} {
		if strings.Contains(got, bad) {
			t.Errorf("plainText() conserva %q: %s", bad, got)
		}
	}
	if !strings.Contains(got, "enlace") {
		t.Errorf("plainText() perdió el texto del enlace: %s", got)
	}
}

func TestNavFromLinks(t *testing.T) {
	links := navFromLinks([]domain.NavLink{
		{Label: "A", URL: "/a", Children: []domain.NavLink{{Label: "B", URL: "/b"}}},
	})
	if len(links) != 1 || len(links[0].Children) != 1 || links[0].Children[0].Label != "B" {
		t.Errorf("navFromLinks = %+v", links)
	}
}

func TestResolveTaxonomy(t *testing.T) {
	cfg := domain.TaxonomyConfig{
		Name: "tags", Field: "tags", Hierarchical: true, Strict: true,
		Terms: []domain.TermConfig{
			{Name: "guia", Description: "Guías.", Aliases: []string{"guide"}, Weight: 5},
			{Name: "planned", Description: "Sin posts."},
		},
	}
	posts := []domain.BlogPost{
		{Tags: []string{"guide", "a/b"}},
		{Tags: []string{"code"}},
	}

	rt, warnings := resolveTaxonomy(cfg, posts)

	// El alias "guide" se normaliza a "guia".
	if !rt.postHas(posts[0], "guia") {
		t.Error("postHas(guia) = false, want true (alias guide)")
	}
	// Jerárquico: "a/b" pertenece a "a" y a "a/b".
	if !rt.postHas(posts[0], "a") || !rt.postHas(posts[0], "a/b") {
		t.Error("postHas jerárquico incorrecto")
	}
	if rt.postHas(posts[1], "a") {
		t.Error("postHas(a) en otro post = true, want false")
	}

	// strict avisa de "code", "a" y "a/b" (no declarados).
	found := false
	for _, w := range warnings {
		if strings.Contains(w, `"code"`) {
			found = true
		}
	}
	if !found {
		t.Errorf("faltan avisos de términos no declarados: %v", warnings)
	}

	// "guia" conserva su descripción; "planned" existe aunque no tenga posts.
	var guia *resolvedTerm
	for i := range rt.terms {
		if rt.terms[i].Name == "guia" {
			guia = &rt.terms[i]
		}
	}
	if guia == nil || guia.Description != "Guías." {
		t.Errorf("término guia sin descripción: %+v", guia)
	}
}

func TestPaginate(t *testing.T) {
	posts := make([]domain.BlogPost, 13)
	pages := paginate(posts, 6)
	if len(pages) != 3 {
		t.Fatalf("paginate() devolvió %d páginas, want 3", len(pages))
	}
	if len(pages[0]) != 6 || len(pages[1]) != 6 || len(pages[2]) != 1 {
		t.Fatalf("tamaños de página = %d,%d,%d", len(pages[0]), len(pages[1]), len(pages[2]))
	}

	empty := paginate(nil, 6)
	if len(empty) != 1 || len(empty[0]) != 0 {
		t.Fatalf("paginate(nil) = %v, want una página vacía", empty)
	}
}

func TestNewPagination(t *testing.T) {
	p := newPagination("/blog/", 1, 3)
	if p.HasPrev() {
		t.Error("página 1 no debe tener prev")
	}
	if !p.HasNext() || p.NextURL != "/blog/2/" {
		t.Errorf("NextURL = %q, want /blog/2/", p.NextURL)
	}

	mid := newPagination("/blog/", 2, 3)
	if mid.PrevURL != "/blog/" {
		t.Errorf("PrevURL = %q, want /blog/", mid.PrevURL)
	}
	if mid.NextURL != "/blog/3/" {
		t.Errorf("NextURL = %q, want /blog/3/", mid.NextURL)
	}

	last := newPagination("/blog/", 3, 3)
	if last.HasNext() {
		t.Error("última página no debe tener next")
	}
}

func TestReadingMinutes(t *testing.T) {
	if got := readingMinutes("hola mundo"); got != 1 {
		t.Errorf("readingMinutes() = %d, want 1", got)
	}
}

func TestSortPinned(t *testing.T) {
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	posts := []domain.BlogPost{
		{Title: "newer", Published: newer},
		{Title: "pinned", Published: older, Pinned: true},
	}
	got := sortPinned(posts)
	if got[0].Title != "pinned" {
		t.Errorf("sortPinned() primero = %q, want pinned", got[0].Title)
	}
	if got[1].Title != "newer" {
		t.Errorf("sortPinned() segundo = %q, want newer", got[1].Title)
	}
}
