package application

import (
	"strings"
	"testing"
)

func TestRewriteHTMLBase(t *testing.T) {
	in := `<a href="/blog/x/">x</a>` +
		`<img src="/assets/a.png" srcset="/a-480.webp 480w, /a-768.webp 768w">` +
		`<meta http-equiv="refresh" content="0; url=/blog/y/">` +
		`<a href="https://example.com/x">e</a>` +
		`<a href="//cdn.example.com/x">c</a>` +
		`<a href="/godian/blog/z/">z</a>`

	out := string(rewriteHTMLBase([]byte(in), "/godian"))
	for _, want := range []string{
		`href="/godian/blog/x/"`,
		`src="/godian/assets/a.png"`,
		`srcset="/godian/a-480.webp 480w, /godian/a-768.webp 768w"`,
		`content="0; url=/godian/blog/y/"`,
		`href="https://example.com/x"`,
		`href="//cdn.example.com/x"`,
		`href="/godian/blog/z/"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q en %s", want, out)
		}
	}
	if strings.Contains(out, "/godian/godian") {
		t.Errorf("doble prefijo: %s", out)
	}
}

func TestRewriteCSSBase(t *testing.T) {
	in := `.a{background:url(/assets/a.png)}` +
		`.b{background:url("/assets/b.png")}` +
		`.c{background:url(data:image/png;base64,x)}` +
		`.d{background:url(../x.png)}`
	out := string(rewriteCSSBase([]byte(in), "/godian"))
	for _, want := range []string{
		`url(/godian/assets/a.png)`,
		`url("/godian/assets/b.png")`,
		`url(data:image/png;base64,x)`,
		`url(../x.png)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q en %s", want, out)
		}
	}
}

func TestRewriteBasePathNoop(t *testing.T) {
	in := []byte(`<a href="/blog/x/">x</a>`)
	if got := string(rewriteBasePath("index.html", in, "")); got != string(in) {
		t.Errorf("sin base no debe cambiar: %s", got)
	}
	if got := string(rewriteBasePath("app.js", in, "/godian")); got != string(in) {
		t.Errorf("JS no se reescribe aquí: %s", got)
	}
}
