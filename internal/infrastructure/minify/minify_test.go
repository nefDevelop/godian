package minify

import (
	"strings"
	"testing"
)

func TestCSS(t *testing.T) {
	in := "/* c */\nbody {\n  color : red ;\n  content: \"a  b\";\n}\n"
	out := string(CSS([]byte(in)))
	if strings.Contains(out, "/*") || strings.Contains(out, "\n") {
		t.Errorf("CSS() no minificó: %q", out)
	}
	if !strings.Contains(out, `content:"a  b"`) {
		t.Errorf("CSS() rompió una cadena: %q", out)
	}
	if !strings.Contains(out, "body{color:red") {
		t.Errorf("CSS() = %q", out)
	}
}

func TestHTML(t *testing.T) {
	in := "<!-- c -->\n<p>hola   mundo</p>\n<pre>a\n  b</pre>\n"
	out := string(HTML([]byte(in)))
	if strings.Contains(out, "<!--") {
		t.Errorf("HTML() no quitó comentarios: %q", out)
	}
	if !strings.Contains(out, "<pre>a\n  b</pre>") {
		t.Errorf("HTML() rompió <pre>: %q", out)
	}
	if !strings.Contains(out, "<p>hola mundo") {
		t.Errorf("HTML() no colapsó espacios: %q", out)
	}
	if strings.Contains(out, "hola   mundo") {
		t.Errorf("HTML() dejó espacios redundantes: %q", out)
	}
}

func FuzzCSS(f *testing.F) {
	f.Add([]byte("body{color:red}"))
	f.Add([]byte("/* c */ a{content:\"x y\"}"))
	f.Fuzz(func(t *testing.T, in []byte) {
		_ = CSS(in)
	})
}

func FuzzHTML(f *testing.F) {
	f.Add([]byte("<p>hola   mundo</p>"))
	f.Add([]byte("<pre>a\n b</pre><!--x-->"))
	f.Fuzz(func(t *testing.T, in []byte) {
		_ = HTML(in)
	})
}

func TestJS(t *testing.T) {
	in := []byte("var a = 1; // comentario\nvar url = \"http://x/y\";\nvar re = /a\\/\\/b/;\n")
	out := string(JS(in))
	if strings.Contains(out, "comentario") {
		t.Errorf("JS() no quitó el comentario: %q", out)
	}
	if !strings.Contains(out, `"http://x/y"`) {
		t.Errorf("JS() rompió una cadena con //: %q", out)
	}
	if !strings.Contains(out, `/a\/\/b/`) {
		t.Errorf("JS() rompió una expresión regular: %q", out)
	}
	if len(out) >= len(in) {
		t.Errorf("JS() no redujo el tamaño: %q", out)
	}
}

func FuzzJS(f *testing.F) {
	f.Add([]byte("var a = 1; // x\n"))
	f.Add([]byte("const s = `a${b}c`;"))
	f.Add([]byte("return\n{x:1}"))
	f.Fuzz(func(t *testing.T, in []byte) {
		_ = JS(in)
	})
}
