package images

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nefdevelop/godian/internal/domain"
)

func TestProcessAndPicture(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 800, 400))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	f, err := os.Create(filepath.Join(assets, "pic.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()

	out := filepath.Join(dir, "dist")
	p := New(assets, out, true, nil)
	if err := p.Process(); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	for _, name := range []string{"pic.png", "pic.webp", "pic-480.webp"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("falta %s: %v", name, err)
		}
	}

	html := string(p.Picture("/pic.png", "Hola"))
	if !strings.Contains(html, "<picture>") || !strings.Contains(html, "pic-480.webp 480w") {
		t.Errorf("Picture() = %s", html)
	}
	if !strings.Contains(html, `width="800" height="400"`) {
		t.Errorf("Picture() sin dimensiones: %s", html)
	}

	plain := string(p.Picture("/nope.jpg", "x"))
	if !strings.Contains(plain, "<img") || strings.Contains(plain, "<picture>") {
		t.Errorf("Picture() desconocido = %s", plain)
	}
}

func TestLQIPAndPreset(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))
	img.Set(0, 0, color.RGBA{0, 128, 255, 255})
	f, err := os.Create(filepath.Join(assets, "pic.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()

	out := filepath.Join(dir, "dist")
	presets := map[string]domain.ImagePreset{
		"thumb": {Width: 200, Height: 100, Fit: "fill", Focal: "center"},
	}
	p := New(assets, out, true, presets)
	if err := p.Process(); err != nil {
		t.Fatalf("Process: %v", err)
	}

	// LQIP: data URI en el <img>.
	if html := string(p.Picture("/pic.png", "x")); !strings.Contains(html, "background-image:url(data:image/webp;base64") {
		t.Errorf("Picture sin LQIP: %s", html)
	}

	// Preset generado y referenciado.
	if _, err := os.Stat(filepath.Join(out, "pic-thumb.webp")); err != nil {
		t.Errorf("falta pic-thumb.webp: %v", err)
	}
	preset, ok := p.Preset("/pic.png", "x", "thumb")
	if !ok || !strings.Contains(string(preset), "pic-thumb.webp") {
		t.Errorf("Preset() = %q, %v", preset, ok)
	}
	if _, ok := p.Preset("/pic.png", "x", "nope"); ok {
		t.Error("Preset inexistente no debe existir")
	}
}
