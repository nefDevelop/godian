package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

// block es un shortcode de bloque: {{#island nombre k=v}} … {{/island}}.
type block struct {
	kind  string
	name  string
	props map[string]string
	inner string
}

var (
	// Las props pueden contener valores entre comillas con espacios (p. ej. label="Ver más"),
	// así que se captura el resto de la línea hasta el cierre y se delega el parseo a tokenize.
	blockStartRe = regexp.MustCompile(`^\{\{#\s*(island|partial)\s+([A-Za-z0-9_-]+)(.*)\}\}$`)
	blockEndRe   = regexp.MustCompile(`^\{\{/\s*(island|partial)\s*\}\}$`)
)

// extractBlocks sustituye cada shortcode de bloque por un marcador HTML y
// devuelve el texto procesado junto con los bloques encontrados (soporta anidamiento).
func extractBlocks(src string) (string, []block) {
	lines := strings.Split(src, "\n")
	var out []string
	var blocks []block
	depth := 0
	inFence := false
	var cur block
	var inner []string

	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if depth == 0 {
			if isFence(t) {
				inFence = !inFence
				out = append(out, ln)
				continue
			}
			if inFence {
				out = append(out, ln)
				continue
			}
			if m := blockStartRe.FindStringSubmatch(t); m != nil {
				depth = 1
				cur = block{kind: m[1], name: m[2], props: parseProps(tokenize(m[3]))}
				inner = nil
				continue
			}
			out = append(out, ln)
			continue
		}
		// Dentro de un bloque: acumular (los anidados se procesan en la recursión).
		if blockStartRe.MatchString(t) {
			depth++
			inner = append(inner, ln)
			continue
		}
		if blockEndRe.MatchString(t) {
			depth--
			if depth == 0 {
				cur.inner = strings.Join(inner, "\n")
				out = append(out, placeholderFor(len(blocks)))
				blocks = append(blocks, cur)
				continue
			}
			inner = append(inner, ln)
			continue
		}
		inner = append(inner, ln)
	}

	// Bloque sin cerrar: se trata como texto literal.
	if depth > 0 {
		out = append(out, inner...)
	}
	return strings.Join(out, "\n"), blocks
}

// isFence indica si la línea abre o cierra un bloque de código ``` o ~~~.
func isFence(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

// placeholderFor devuelve el marcador HTML de un bloque.
func placeholderFor(i int) string {
	return fmt.Sprintf(`<div data-godian-block="%d"></div>`, i)
}
