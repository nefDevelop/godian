package islands

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/fstest"
)

type fakeRenderer struct{ templates map[string]string }

func (f fakeRenderer) Render(w io.Writer, name string, data any) error {
	body, ok := f.templates[name]
	if !ok {
		return fmt.Errorf("template %q no existe", name)
	}
	m, _ := data.(map[string]string)
	body = strings.ReplaceAll(body, "{{start}}", m["start"])
	body = strings.ReplaceAll(body, "{{text}}", m["text"])
	_, _ = io.WriteString(w, body)
	return nil
}

func newTestRegistry() *Registry {
	fsys := fstest.MapFS{
		"templates/islands/counter.html": {Data: []byte("<p>{{start}}</p>")},
		"templates/partials/cta.html":    {Data: []byte("<a>{{text}}</a>")},
		"assets/islands/counter.js":      {Data: []byte("export function mount(){}")},
		"assets/islands/tabs.js":         {Data: []byte("export function mount(){}")},
	}
	r := New(fsys)
	r.SetRenderer(fakeRenderer{templates: map[string]string{
		"islands/counter.html": "<p>{{start}}</p>",
		"partials/cta.html":    "<a>{{text}}</a>",
	}})
	return r
}

func TestIslandSSRAndWrapper(t *testing.T) {
	r := newTestRegistry()
	out, err := r.Island("counter", map[string]string{"start": "3", "load": "visible"})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{`data-island="counter"`, `data-island-load="visible"`, `data-props=`, `<p>3</p>`} {
		if !strings.Contains(s, want) {
			t.Errorf("falta %q en %s", want, s)
		}
	}
	if strings.Contains(s, `"load"`) {
		t.Error("load no debe ir en data-props")
	}
}

func TestIslandJSOnly(t *testing.T) {
	r := newTestRegistry()
	out, err := r.Island("tabs", map[string]string{"items": "a,b"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `data-island="tabs"`) {
		t.Errorf("placeholder incorrecto: %s", out)
	}
}

func TestIslandUnknown(t *testing.T) {
	r := newTestRegistry()
	if _, err := r.Island("nope", nil); err == nil {
		t.Error("se esperaba error de isla desconocida")
	}
}

func TestPartial(t *testing.T) {
	r := newTestRegistry()
	out, err := r.Partial("cta", map[string]string{"text": "Hola"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<a>Hola</a>") {
		t.Errorf("partial = %s", out)
	}
}

func TestInvalidName(t *testing.T) {
	r := newTestRegistry()
	if _, err := r.Island("../evil", nil); err == nil {
		t.Error("se esperaba error de nombre inválido")
	}
}
