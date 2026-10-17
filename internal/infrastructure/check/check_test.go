package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nefdevelop/godian/internal/domain"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "used.png"))
	writeFile(t, filepath.Join(dir, "orphan.png"))

	cfg := domain.DefaultConfig()
	cfg.Build.AssetsDir = dir

	c := domain.Content{
		Posts: []domain.BlogPost{
			{Slug: "a", Title: "A", Body: "![alt](/used.png)\n\n![ ](/missing.png)"},
		},
		Series: []domain.Series{{Slug: "s", Chapters: []string{"nope"}}},
	}

	joined := strings.Join(Run(cfg, c), "\n")
	for _, want := range []string{
		"no encontrada: /missing.png",
		"/orphan.png",
		`capítulo "nope"`,
		"sin alt",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("falta %q en:\n%s", want, joined)
		}
	}
}

func TestRunIgnoresCode(t *testing.T) {
	cfg := domain.DefaultConfig()
	cfg.Build.AssetsDir = t.TempDir()
	c := domain.Content{
		Posts: []domain.BlogPost{{Slug: "a", Title: "A", Body: "```\n![x](/nope.png)\n```\n`{{image src=\"/a.jpg\"}}`"}},
	}
	if p := Run(cfg, c); len(p) != 0 {
		t.Errorf("no debe analizar código: %v", p)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
