package theme

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmbedded(t *testing.T) {
	th, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if th.Manifest.Name == "" {
		t.Error("manifest sin nombre")
	}
}

func TestScaffold(t *testing.T) {
	out := filepath.Join(t.TempDir(), "mi-tema")
	if err := Scaffold(out); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"theme.yaml", "templates/home.html", "assets/css/styles.css"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(f))); err != nil {
			t.Errorf("falta %s: %v", f, err)
		}
	}
	th, err := Load(out)
	if err != nil {
		t.Fatalf("Load del tema copiado: %v", err)
	}
	if th.Manifest.Name != "default" {
		t.Errorf("manifest = %+v", th.Manifest)
	}
}

// Un directorio que existe solo en el tema en disco (p. ej. icons/) debe poder
// listarse; antes el overlay fallaba si faltaba en el embebido.
func TestOverlayReadDirMergesAndToleratesMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "icons"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "icons", "menu.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	th, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	entries, err := fs.ReadDir(th.FS(), "icons") // solo en disco
	if err != nil {
		t.Fatalf("ReadDir(icons): %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "menu.svg" {
		t.Errorf("icons = %v", entries)
	}

	if _, err := fs.ReadDir(th.FS(), "templates"); err != nil { // solo en embebido
		t.Errorf("ReadDir(templates): %v", err)
	}

	if _, err := fs.ReadDir(th.FS(), "no-existe"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("se esperaba ErrNotExist, got %v", err)
	}
}
