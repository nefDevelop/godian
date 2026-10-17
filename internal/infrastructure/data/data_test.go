package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "links.yaml"), "- {label: A, url: /a}\n")
	writeFile(t, filepath.Join(dir, "cfg.json"), `{"x": 1}`)
	writeFile(t, filepath.Join(dir, "team.csv"), "name,role\nAna,dev\n")

	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m["links"]; !ok {
		t.Error("falta clave links")
	}
	if cfg, ok := m["cfg"].(map[string]any); !ok || cfg["x"] == nil {
		t.Errorf("cfg = %#v", m["cfg"])
	}
	team, ok := m["team"].([]map[string]string)
	if !ok || len(team) != 1 || team[0]["name"] != "Ana" {
		t.Errorf("team = %#v", m["team"])
	}
}

func TestLoadMissingDir(t *testing.T) {
	m, err := Load(filepath.Join(t.TempDir(), "nope"))
	if err != nil || len(m) != 0 {
		t.Errorf("Load(dir inexistente) = %v, %v", m, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
