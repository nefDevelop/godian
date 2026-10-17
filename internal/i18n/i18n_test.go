package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestT(t *testing.T) {
	if got := T("es", "nav.home"); got != "Inicio" {
		t.Errorf("T(es, nav.home) = %q, want Inicio", got)
	}
	if got := T("en", "nav.home"); got != "Home" {
		t.Errorf("T(en, nav.home) = %q, want Home", got)
	}
	if got := T("fr", "nav.home"); got != "Home" {
		t.Errorf("T(fr, nav.home) = %q, want Home (fallback)", got)
	}
	if got := T("en", "no.such.key"); got != "no.such.key" {
		t.Errorf("T(unknown) = %q, want la clave", got)
	}
}

func TestLoadExternalComplete(t *testing.T) {
	dir := t.TempDir()
	m := map[string]string{}
	for _, k := range Keys() {
		m[k] = "x"
	}
	data, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(dir, "es.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, warns := Load(dir, "es", nil); len(warns) != 0 {
		t.Errorf("archivo completo no debe avisar: %v", warns)
	}
}

func TestLoadExternalIncomplete(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "es.json"), []byte(`{"nav.home":"Inicio"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, warns := Load(dir, "es", nil)
	if !strings.Contains(strings.Join(warns, "\n"), "faltan") {
		t.Errorf("debe avisar de claves faltantes: %v", warns)
	}
}

func TestLoadMissingTaxonomyKey(t *testing.T) {
	// Sin carpeta i18n y con una taxonomía nueva (taxonomy.capitulos) → aviso.
	_, warns := Load(filepath.Join(t.TempDir(), "nope"), "es", []string{"taxonomy.capitulos"})
	if !strings.Contains(strings.Join(warns, "\n"), "taxonomy.capitulos") {
		t.Errorf("debe avisar de taxonomy.capitulos: %v", warns)
	}
}
