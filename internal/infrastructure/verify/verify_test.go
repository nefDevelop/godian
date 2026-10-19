package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDist(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.html"),
		`<link rel="stylesheet" href="/assets/css/x.css">`+
			`<a href="/blog/post/">ok</a>`+
			`<a href="/blog/missing/">bad</a>`+
			`<a href="https://example.com">ext</a>`+
			`<a href="#top">anchor</a>`)
	mustWrite(t, filepath.Join(dir, "assets/css/x.css"), "body{}")
	mustWrite(t, filepath.Join(dir, "blog/post/index.html"), "ok")

	problems, err := Dist(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Fatalf("Dist() = %v, want 1 problema", problems)
	}
	if got := problems[0]; got != "index.html -> /blog/missing/" {
		t.Errorf("problema = %q", got)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDistBase(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.html"),
		`<link rel="stylesheet" href="/godian/assets/css/x.css">`+
			`<a href="/godian/blog/post/">ok</a>`+
			`<a href="/godian/blog/missing/">bad</a>`)
	mustWrite(t, filepath.Join(dir, "assets/css/x.css"), "body{}")
	mustWrite(t, filepath.Join(dir, "blog/post/index.html"), "ok")

	problems, err := DistBase(dir, "/godian")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Fatalf("DistBase() = %v, want 1 problema", problems)
	}
	if got := problems[0]; got != "index.html -> /godian/blog/missing/" {
		t.Errorf("problema = %q", got)
	}
}
