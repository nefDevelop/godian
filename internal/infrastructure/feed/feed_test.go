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
