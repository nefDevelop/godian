package images

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	p := New(assets, out, true)
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
