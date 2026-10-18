package theme

import (
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
