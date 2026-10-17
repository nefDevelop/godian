// Package data carga datos estructurados (JSON/YAML/CSV) para usarlos en plantillas.
package data

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load recorre dir y devuelve un mapa clave → valor (clave = ruta sin extensión).
func Load(dir string) (map[string]any, error) {
	out := map[string]any{}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return out, nil
	}
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			return rerr
		}
		key := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
		raw, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		v, perr := parse(filepath.Ext(p), raw)
		if perr != nil {
			return fmt.Errorf("data %s: %w", rel, perr)
		}
		out[key] = v
		return nil
	})
	return out, err
}

func parse(ext string, raw []byte) (any, error) {
	switch strings.ToLower(ext) {
	case ".json":
		var v any
		return v, json.Unmarshal(raw, &v)
	case ".yaml", ".yml":
		var v any
		return v, yaml.Unmarshal(raw, &v)
	case ".csv":
		return parseCSV(raw)
	default:
		return string(raw), nil
	}
}

func parseCSV(raw []byte) ([]map[string]string, error) {
	rows, err := csv.NewReader(bytes.NewReader(raw)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	header := rows[0]
	out := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		m := map[string]string{}
		for i, h := range header {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}
