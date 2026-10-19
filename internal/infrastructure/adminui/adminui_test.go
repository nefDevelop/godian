package adminui

import (
	"strings"
	"testing"
)

func TestRenderIndex(t *testing.T) {
	out, err := RenderIndex(IndexData{
		Lang:       "es",
		SiteTitle:  "Sitio",
		AdminBase:  "/admin/",
		CSS:        []string{"/assets/css/styles.css"},
		AdminCSS:   "/assets/css/admin.css",
		ConnectSrc: "'self'",
	})
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	for _, want := range []string{
		`href="/admin/assets/admin.css"`,
		`href="/assets/css/admin.css"`,
		`src="/admin/assets/app.js"`,
		`src="/admin/vendor/js-yaml.min.js"`,
		"noindex,nofollow",
		"connect-src &#39;self&#39;",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("falta %q en el HTML:\n%s", want, html)
		}
	}
	if strings.Contains(html, "ZgotmplZ") {
		t.Errorf("hay una URL rechazada por html/template:\n%s", html)
	}
}

func TestRenderIndexBasePath(t *testing.T) {
	out, err := RenderIndex(IndexData{AdminBase: "/base/admin/", SiteTitle: "Sitio"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `src="/base/admin/assets/app.js"`) {
		t.Errorf("no se aplicó el base path:\n%s", out)
	}
}
