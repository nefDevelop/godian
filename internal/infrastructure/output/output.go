// Package output escribe los artefactos generados en el directorio de salida.
package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Writer escribe archivos en un directorio raíz.
type Writer struct {
	dir string
}

// NewWriter crea un Writer sobre dir.
func NewWriter(dir string) *Writer {
	return &Writer{dir: dir}
}

// Write escribe data en relPath (relativo a la raíz), creando directorios.
func (w *Writer) Write(relPath string, data []byte) error {
	full := filepath.Join(w.dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

// Clean elimina y recrea el directorio raíz. Rechaza rutas peligrosas.
func (w *Writer) Clean() error {
	dir := strings.TrimSpace(w.dir)
	if dir == "" || dir == "." || dir == "/" || strings.HasPrefix(dir, "..") {
		return fmt.Errorf("clean: ruta de salida insegura %q", w.dir)
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}
