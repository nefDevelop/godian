package feed

import (
	"strings"
	"testing"
	"time"

	"github.com/nefdevelop/godian/internal/domain"
)

func sample() (domain.SiteConfig, []domain.BlogPost) {
	site := domain.SiteConfig{Title: "T", URL: "https://x.com", Description: "D", Language: "en"}
	posts := []domain.BlogPost{{
		Title:       "Hola",
		Slug:        "hola",
		Description: "desc",
		Published:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		Tags:        []string{"go"},
	}}
	return site, posts
}

func TestAtom(t *testing.T) {
	site, posts := sample()
	b, err := New().Atom(site, posts)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, w := range []string{"<feed", "http://www.w3.org/2005/Atom", "https://x.com/blog/hola/", "Hola"} {
		if !strings.Contains(s, w) {
			t.Errorf("Atom no contiene %q", w)
		}
	}
}

func TestJSONFeed(t *testing.T) {
	site, posts := sample()
	b, err := New().JSONFeed(site, posts)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, w := range []string{"jsonfeed.org/version/1.1", `"url": "https://x.com/blog/hola/"`, "Hola"} {
		if !strings.Contains(s, w) {
			t.Errorf("JSONFeed no contiene %q", w)
		}
	}
}

func TestSitemapsSingle(t *testing.T) {
	site, _ := sample()
	files, err := New().Sitemaps(site, []string{"/", "/blog/hola/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("esperaba 1 fichero, obtuve %d", len(files))
	}
	s, ok := files["sitemap.xml"]
	if !ok {
		t.Fatal("falta sitemap.xml")
	}
	if !strings.Contains(string(s), "<urlset") {
		t.Errorf("sitemap.xml debería ser un urlset, obtuve %s", s)
	}
}

func TestSitemapsSplitProducesIndex(t *testing.T) {
	site, _ := sample()
	paths := []string{"/a/", "/b/", "/c/", "/d/", "/e/"}
	files, err := New().sitemaps(site, paths, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 {
		t.Fatalf("esperaba 4 ficheros (3 hijos + índice), obtuve %d", len(files))
	}
	index, ok := files["sitemap.xml"]
	if !ok {
		t.Fatal("falta sitemap.xml")
	}
	if !strings.Contains(string(index), "<sitemapindex") {
		t.Errorf("sitemap.xml debería ser un índice, obtuve %s", index)
	}
	for _, c := range []string{"sitemap-1.xml", "sitemap-2.xml", "sitemap-3.xml"} {
		if _, ok := files[c]; !ok {
			t.Errorf("falta %s", c)
		}
		if !strings.Contains(string(index), site.URL+"/"+c) {
			t.Errorf("el índice no referencia a %s", c)
		}
	}
	first := string(files["sitemap-1.xml"])
	if !strings.Contains(first, site.URL+"/a/") || !strings.Contains(first, site.URL+"/b/") {
		t.Errorf("sitemap-1.xml no contiene las dos primeras URLs: %s", first)
	}
}
