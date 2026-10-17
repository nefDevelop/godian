package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nefdevelop/godian/internal/domain"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func loaderFor(t *testing.T, dir string) *Loader {
	t.Helper()
	cfg := domain.DefaultConfig()
	cfg.Build.ContentDir = dir
	cfg.Collections = domain.DefaultCollections()
	return NewLoader(cfg)
}

func TestLoadPostsNestedSlugs(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "blog", "a.md"), "---\ntitle: A\npublished: 2026-01-01\ndescription: x\n---\ncuerpo")
	write(t, filepath.Join(dir, "blog", "sub", "b.md"), "---\ntitle: B\npublished: 2026-01-02\ndescription: y\n---\ncuerpo")

	posts, err := loaderFor(t, dir).LoadPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Fatalf("posts = %d, want 2", len(posts))
	}
	slugs := map[string]bool{}
	for _, p := range posts {
		slugs[p.Slug] = true
	}
	if !slugs["a"] || !slugs["sub/b"] {
		t.Errorf("slugs = %v, want a y sub/b", slugs)
	}
}

func TestDuplicateSlug(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "blog", "a.md"), "---\ntitle: A\n---\nx")
	write(t, filepath.Join(dir, "blog", "a.mdx"), "---\ntitle: A2\n---\nx")

	if _, err := loaderFor(t, dir).LoadPosts(); err == nil {
		t.Error("se esperaba error por slug duplicado")
	}
}

func TestMissingContentDirIsEmpty(t *testing.T) {
	c, err := loaderFor(t, filepath.Join(t.TempDir(), "no-existe")).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(c.Posts) != 0 || len(c.Docs) != 0 {
		t.Errorf("se esperaba contenido vacío, got %+v", c)
	}
}

func TestInvalidFrontmatter(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "blog", "bad.md"), "---\ntitle: [sin cerrar\n---\nx")

	if _, err := loaderFor(t, dir).LoadPosts(); err == nil {
		t.Error("se esperaba error de frontmatter inválido")
	}
}

func TestAliasesLoaded(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "blog", "a.md"),
		"---\ntitle: A\npublished: 2026-01-01\naliases: [/old-a/, /blog/viejo/]\n---\nx")

	posts, err := loaderFor(t, dir).LoadPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || len(posts[0].Aliases) != 2 {
		t.Errorf("Aliases = %+v, want 2", posts[0].Aliases)
	}
}

func TestHiddenFilesIgnored(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "blog", ".draft.md"), "---\ntitle: oculto\n---\nx")
	write(t, filepath.Join(dir, "blog", "ok.md"), "---\ntitle: ok\npublished: 2026-01-01\n---\nx")

	posts, err := loaderFor(t, dir).LoadPosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].Slug != "ok" {
		t.Errorf("posts = %+v, want solo 'ok'", posts)
	}
}
