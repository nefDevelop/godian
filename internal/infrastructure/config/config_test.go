package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil || !strings.Contains(err.Error(), "no se encontró") {
		t.Fatalf("err = %v, want 'no se encontró'", err)
	}
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "site.yaml")
	if err := os.WriteFile(path, []byte("site:\n  title: X\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Site.Title != "X" {
		t.Errorf("title = %q", cfg.Site.Title)
	}
	if cfg.Build.ContentDir != "content" || !cfg.Build.Clean {
		t.Errorf("defaults no aplicados: %+v", cfg.Build)
	}
	if len(cfg.Taxonomies) != 2 || len(cfg.Collections) != 5 {
		t.Errorf("defaults de taxonomías/colecciones no aplicados")
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "site.yaml")
	if err := os.WriteFile(path, []byte("site: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("se esperaba error de YAML inválido")
	}
}

func TestLoadRenderDefaultsAndOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "site.yaml")
	yml := "site:\n  title: X\nrender:\n  headings:\n    anchors: false\n"
	if err := os.WriteFile(path, []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Render.Headings.Anchors {
		t.Error("anchors debería ser false (override)")
	}
	if cfg.Render.Headings.Anchor != "#" {
		t.Errorf("anchor_label por defecto = %q", cfg.Render.Headings.Anchor)
	}
	if cfg.Render.ExternalLinks.Target != "_blank" || cfg.Render.ExternalLinks.Rel != "noopener noreferrer" {
		t.Errorf("external_links por defecto = %+v", cfg.Render.ExternalLinks)
	}
	img := cfg.Render.Images
	if img.Loading != "lazy" || img.Decoding != "async" || !img.Figure {
		t.Errorf("images por defecto = %+v", img)
	}
}
