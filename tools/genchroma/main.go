// Command genchroma regenera el CSS de resaltado (Chroma) del tema por defecto.
//
// Uso: go run ./tools/genchroma
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

const outputPath = "internal/infrastructure/theme/default/assets/css/chroma.css"

func main() {
	f := html.New(html.WithClasses(true))

	var light, dark strings.Builder
	if err := f.WriteCSS(&light, styles.Get("github")); err != nil {
		fmt.Fprintln(os.Stderr, "light:", err)
		os.Exit(1)
	}
	if err := f.WriteCSS(&dark, styles.Get("github-dark")); err != nil {
		fmt.Fprintln(os.Stderr, "dark:", err)
		os.Exit(1)
	}

	var out strings.Builder
	out.WriteString("/* Generado por tools/genchroma. No editar a mano. */\n")
	out.WriteString(light.String())
	out.WriteString("\n")
	out.WriteString(scopeDark(dark.String()))

	if err := os.WriteFile(outputPath, []byte(out.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("escrito:", outputPath)
}

// scopeDark prefija cada selector con ".dark " para el modo oscuro explícito.
func scopeDark(css string) string {
	var b strings.Builder
	for _, line := range strings.Split(css, "\n") {
		switch {
		case line == "":
		case strings.Contains(line, "*/ "):
			i := strings.Index(line, "*/ ")
			line = line[:i+3] + ".dark " + line[i+3:]
		case strings.HasPrefix(line, ".") && strings.Contains(line, "{"):
			line = ".dark " + line
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
