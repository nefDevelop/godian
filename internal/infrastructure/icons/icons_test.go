package icons

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestHTMLDecorative(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := r.HTML("play", "")
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	s := string(out)
	for _, want := range []string{
		`<svg class="icon icon-play"`,
		`viewBox="0 0 24 24"`,
		`aria-hidden="true"`,
		`stroke="currentColor"`,
		`<path d="`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("falta %q en %s", want, s)
		}
	}
	if strings.Contains(s, `role="img"`) {
		t.Errorf("un icono decorativo no debe llevar role=img: %s", s)
	}
}

func TestHTMLLabelled(t *testing.T) {
	r, _ := New()
	out, err := r.HTML("menu", "Abrir menú")
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `role="img"`) || !strings.Contains(s, `aria-label="Abrir menú"`) {
		t.Errorf("el icono significativo debe ser accesible: %s", s)
	}
	if !strings.Contains(s, `<title>Abrir menú</title>`) {
		t.Errorf("falta el <title>: %s", s)
	}
	if strings.Contains(s, "aria-hidden") {
		t.Errorf("un icono significativo no debe ser aria-hidden: %s", s)
	}
}

func TestUnknownIconWarnsAndPlaceholders(t *testing.T) {
	r, _ := New()
	var warns []string
	r.SetWarning(func(w string) { warns = append(warns, w) })
	out, err := r.HTML("no-existe-jamas", "")
	if err != nil {
		t.Fatalf("un icono desconocido no debe romper el build: %v", err)
	}
	if !strings.Contains(string(out), "icon-missing") {
		t.Errorf("se esperaba un placeholder visible: %s", out)
	}
	// El aviso se emite una sola vez aunque se use varias veces.
	_, _ = r.HTML("no-existe-jamas", "")
	if len(warns) != 1 {
		t.Errorf("avisos = %d, se esperaba 1: %v", len(warns), warns)
	}
}

func TestThemeIconsOverrideCore(t *testing.T) {
	r, _ := New()
	theme := fstest.MapFS{
		"icons/play.svg": &fstest.MapFile{Data: []byte(`<!-- custom -->` +
			`<svg width="24" height="24" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`)},
		"icons/brand.svg": &fstest.MapFile{Data: []byte(`<svg viewBox="0 0 24 24"><path d="M0 0h24v24H0z"/></svg>`)},
	}
	if err := r.LoadDir(theme, "icons"); err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	// play del tema sustituye al de Lucide.
	out, _ := r.HTML("play", "")
	if !strings.Contains(string(out), `<circle cx="12" cy="12" r="10"/>`) {
		t.Errorf("el icono del tema no sobrescribió al del core: %s", out)
	}
	if strings.Contains(string(out), "<path") {
		t.Errorf("no debería quedar el icono Lucide: %s", out)
	}

	// brand solo existe en el tema.
	if !r.Has("brand") {
		t.Error("Has(brand) debería ser true")
	}
	out, err := r.HTML("brand", "")
	if err != nil || !strings.Contains(string(out), `<path d="M0 0h24v24H0z"/>`) {
		t.Errorf("icono del tema: %v / %s", err, out)
	}
}

func TestLoadDirMissingIsFine(t *testing.T) {
	r, _ := New()
	if err := r.LoadDir(fstest.MapFS{}, "icons"); err != nil {
		t.Errorf("un dir icons/ ausente no debe fallar: %v", err)
	}
}

func TestInnerSVG(t *testing.T) {
	if _, err := innerSVG("<div>no svg</div>"); err == nil {
		t.Error("se esperaba error sin <svg>")
	}
	got, err := innerSVG("<svg a=\"1\"><path d=\"M0\"/></svg>")
	if err != nil || got != `<path d="M0"/>` {
		t.Errorf("innerSVG = %q, %v", got, err)
	}
}
