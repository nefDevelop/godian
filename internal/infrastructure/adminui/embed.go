// Package adminui contiene el admin estático que se publica con el sitio.
// No usa npm: los assets y las librerías vendorizadas van embebidos.
package adminui

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"strings"
)

//go:embed static
var staticFS embed.FS

// IndexData son los datos que el build inyecta en el shell del admin.
type IndexData struct {
	Lang       string
	SiteTitle  string
	AdminBase  string
	Favicon    string
	CSS        []string
	AdminCSS   string
	ConnectSrc string
}

// FS devuelve el árbol static/ (index.html, assets/ y vendor/).
func FS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}

// RenderIndex genera admin/index.html con el CSS del tema y una CSP concreta.
func RenderIndex(data IndexData) ([]byte, error) {
	if data.Lang == "" {
		data.Lang = "en"
	}
	if data.AdminBase == "" {
		data.AdminBase = "/admin/"
	}
	if !strings.HasSuffix(data.AdminBase, "/") {
		data.AdminBase += "/"
	}
	if data.SiteTitle == "" {
		data.SiteTitle = "godian"
	}
	connect := data.ConnectSrc
	if connect == "" {
		connect = "'self'"
	}
	csp := fmt.Sprintf(
		"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; font-src 'self' data: https:; "+
			"connect-src %s; base-uri 'self'; "+
			"object-src 'none'; form-action 'self'", connect)

	raw, err := fs.ReadFile(staticFS, "static/index.html")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("admin").Parse(string(raw))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{
		"Lang":      data.Lang,
		"SiteTitle": data.SiteTitle,
		"AdminBase": data.AdminBase,
		"Favicon":   data.Favicon,
		"CSS":       data.CSS,
		"AdminCSS":  data.AdminCSS,
		"CSP":       csp,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
