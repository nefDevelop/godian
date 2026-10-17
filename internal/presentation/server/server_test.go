package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotify(t *testing.T) {
	s := New(":0", t.TempDir(), nil, nil)
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()

	s.notify()

	select {
	case <-ch:
	default:
		t.Error("el cliente no recibió la notificación")
	}
}

func TestSignatureChanges(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	first := signature([]string{dir})

	// Añadir un archivo cambia la huella (nº de archivos + mtime).
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("2"), 0o644); err != nil {
		t.Fatal(err)
	}
	second := signature([]string{dir})

	if first == second {
		t.Errorf("signature no cambió: %q", first)
	}
}

func TestHandleStaticInjectsReload(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<body>hola</body>"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := New(":0", dir, nil, nil)

	rec := httptest.NewRecorder()
	s.handleStatic(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "__reload") {
		t.Errorf("no se inyectó el snippet de recarga: %s", rec.Body.String())
	}
}

func TestHandleStaticContentType(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "styles.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := New(":0", dir, nil, nil)

	rec := httptest.NewRecorder()
	s.handleStatic(rec, httptest.NewRequest("GET", "/styles.css", nil))

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("Content-Type = %q, want text/css", ct)
	}
}

func TestHandleStaticNotFound(t *testing.T) {
	s := New(":0", t.TempDir(), nil, nil)
	rec := httptest.NewRecorder()
	s.handleStatic(rec, httptest.NewRequest("GET", "/nope", nil))
	if rec.Code != 404 {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
