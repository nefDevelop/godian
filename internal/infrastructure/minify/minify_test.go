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
	if !strings.Contains(out, "<p>hola mundo</p>") {
		t.Errorf("HTML() no colapsó espacios: %q", out)
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

func TestJSNoop(t *testing.T) {
	in := []byte("var a = 1; // x\n")
	if string(JS(in)) != string(in) {
		t.Error("JS() debe devolver el contenido sin cambios")
	}
}
