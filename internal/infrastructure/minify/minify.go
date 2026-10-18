// Package minify reduce el tamaño de HTML, CSS y JS con tdewolff/minify
// (Go puro, sin cgo), evitando los errores típicos de un minificador casero
// (regex vs. división, ASI, cadenas con //, etc.).
package minify

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

// minifiers registra los minificadores por tipo MIME. El de HTML reutiliza los
// de CSS y JS para el contenido embebido en <style>/<script>.
var minifiers = func() *minify.M {
	m := minify.New()
	m.Add("text/html", &html.Minifier{KeepQuotes: true})
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("text/javascript", js.Minify)
	return m
}()

// HTML minifica HTML (y su CSS/JS embebido).
func HTML(in []byte) []byte { return minifyBytes("text/html", in) }

// CSS minifica CSS.
func CSS(in []byte) []byte { return minifyBytes("text/css", in) }

// JS minifica JavaScript con un parser real.
func JS(in []byte) []byte { return minifyBytes("application/javascript", in) }

// minifyBytes aplica el minificador del tipo MIME. Ante cualquier error devuelve
// la entrada sin cambios (nunca rompe el build por un fallo de minificado).
func minifyBytes(mimetype string, in []byte) []byte {
	out, err := minifiers.Bytes(mimetype, in)
	if err != nil {
		return in
	}
	return out
}
