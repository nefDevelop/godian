package admin

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSite = `site:
  title: Sitio de prueba
  language: es
build:
  content_dir: content
  output_dir: dist
  assets_dir: assets
  i18n_dir: i18n
`

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	if err := os.WriteFile("site.yaml", []byte(testSite), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("content/blog", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("content/docs", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("content/authors", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("content/blog/hello.md", []byte(`---
title: Hola
published: 2026-01-15
description: Demo
tags: [go, web]
author: [Nef]
draft: false
links:
  website: https://example.com
---

# Hola

Texto.
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("content/docs/guia.md", []byte(`---
title: Guía
description: Guía de inicio
pubDate: 2026-01-10
---

# Guía
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("content/authors/nef.yaml", []byte("name: Nef\nbio: Dev\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	h, err := New("site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func localRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.RemoteAddr = "127.0.0.1:12345"
	return req
}

func TestDashboard(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, localRequest("GET", "/admin/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Panel de edición", "Blog", "Documentación", "Hola", "Nef", "/admin/theme/css/styles.css", "/admin/assets/admin.css"} {
		if !strings.Contains(body, want) {
			t.Errorf("falta %q en el dashboard", want)
		}
	}
}

func TestEdit(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, localRequest("GET", "/admin/edit?collection=blog&path="+url.QueryEscape("content/blog/hello.md"), nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Hola", "2026-01-15", "https://example.com"} {
		if !strings.Contains(body, want) {
			t.Errorf("falta %q en el editor", want)
		}
	}
}

func TestSaveRoundTrip(t *testing.T) {
	h := newTestHandler(t)
	fields := map[string]string{
		"collection":     "blog",
		"path":           "content/blog/hello.md",
		"title":          "Hola editado",
		"published":      "2026-01-16",
		"description":    "Descripción editada",
		"tags":           "go, web",
		"categories":     "guias",
		"author":         "Nef",
		"difficulty":     "beginner",
		"draft":          "on",
		"pinned":         "on",
		"order":          "2",
		"links.website":  "https://example.org",
		"downloads_json": `[{"id":"pdf","label":"PDF","url":"https://example.org/x.pdf"}]`,
		"aliases":        "viejo, antiguo",
		"body":           "# Hola editado\n\nNuevo cuerpo.\n",
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	_ = mw.Close()

	req := localRequest("POST", "/admin/save", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Admin-CSRF", h.csrf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "saved=1") {
		t.Errorf("Location = %q", loc)
	}

	data, err := os.ReadFile("content/blog/hello.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"title: Hola editado",
		"published: 2026-01-16",
		"description: Descripción editada",
		"order: 2",
		"draft: true",
		"pinned: true",
		"- guias",
		"- Nef",
		"links:",
		"website: https://example.org",
		"downloads:",
		"id: pdf",
		"# Hola editado",
		"Nuevo cuerpo.",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("falta %q en el archivo guardado:\n%s", want, text)
		}
	}
	if strings.Contains(text, `published: "2026-01-16"`) {
		t.Errorf("la fecha quedó entre comillas:\n%s", text)
	}
}

func TestCreateNew(t *testing.T) {
	h := newTestHandler(t)
	fields := map[string]string{
		"collection":  "blog",
		"filename":    "nuevo-post",
		"title":       "Nuevo post",
		"published":   "2026-02-01",
		"description": "Creado desde el admin",
		"body":        "# Nuevo post\n\nContenido.\n",
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	_ = mw.Close()

	req := localRequest("POST", "/admin/save", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Admin-CSRF", h.csrf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	data, err := os.ReadFile("content/blog/nuevo-post.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"title: Nuevo post", "published: 2026-02-01", "description: Creado desde el admin", "# Nuevo post", "Contenido."} {
		if !strings.Contains(text, want) {
			t.Errorf("falta %q en el archivo nuevo:\n%s", want, text)
		}
	}
}

func TestPathTraversalRejected(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, localRequest("GET", "/admin/edit?collection=blog&path="+url.QueryEscape("../../evil.md"), nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestPreview(t *testing.T) {
	h := newTestHandler(t)
	form := url.Values{"content": {"# Hola **mundo**"}}
	req := localRequest("POST", "/admin/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Admin-CSRF", h.csrf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<h1") || !strings.Contains(body, "<strong>mundo</strong>") {
		t.Errorf("preview inesperado: %s", body)
	}
}

func TestCSRFRequired(t *testing.T) {
	h := newTestHandler(t)
	form := url.Values{"content": {"# Hola"}}
	req := localRequest("POST", "/admin/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestThemeCSS(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, localRequest("GET", "/admin/theme/css/styles.css", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("Content-Type = %q", ct)
	}
	if !strings.Contains(rec.Body.String(), ":root") {
		t.Errorf("no parece CSS del tema")
	}
}

func TestUpload(t *testing.T) {
	h := newTestHandler(t)
	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "foto.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(png)
	_ = mw.Close()

	req := localRequest("POST", "/admin/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Admin-CSRF", h.csrf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var out map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out["url"], "/uploads/") {
		t.Fatalf("url = %q", out["url"])
	}
	name := strings.TrimPrefix(out["url"], "/uploads/")
	if _, err := os.Stat(filepath.Join("assets", "uploads", name)); err != nil {
		t.Errorf("no se guardó la imagen: %v", err)
	}
}

func TestUploadsPreferDist(t *testing.T) {
	h := newTestHandler(t)
	if err := os.MkdirAll("dist/uploads", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("assets/uploads", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("dist/uploads/foto.webp", []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("assets/uploads/foto.jpg", []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, localRequest("GET", "/uploads/foto.webp", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "webp" {
		t.Errorf("webp: status=%d body=%q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, localRequest("GET", "/uploads/foto.jpg", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "jpg" {
		t.Errorf("fallback: status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestLoopbackOnly(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest("GET", "/admin/", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
