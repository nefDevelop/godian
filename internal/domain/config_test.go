package domain

import "testing"

func TestNormalizeBasePath(t *testing.T) {
	cases := map[string]string{
		"":         "",
		"/":        "",
		"godian":   "/godian",
		"/godian/": "/godian",
		" /blog/ ": "/blog",
	}
	for in, want := range cases {
		if got := NormalizeBasePath(in); got != want {
			t.Errorf("NormalizeBasePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeSite(t *testing.T) {
	s := SiteConfig{URL: "https://example.com/godian/"}
	NormalizeSite(&s)
	if s.BasePath != "/godian" || s.URL != "https://example.com/godian" {
		t.Errorf("derivado = %+v", s)
	}

	s = SiteConfig{URL: "https://example.com", BasePath: "/blog/"}
	NormalizeSite(&s)
	if s.BasePath != "/blog" || s.URL != "https://example.com/blog" {
		t.Errorf("explícito = %+v", s)
	}

	s = SiteConfig{URL: "https://example.com"}
	NormalizeSite(&s)
	if s.BasePath != "" || s.URL != "https://example.com" {
		t.Errorf("raíz = %+v", s)
	}
}
