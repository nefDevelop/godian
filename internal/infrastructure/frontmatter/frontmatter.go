// Package frontmatter separa el frontmatter YAML del cuerpo de un documento
// Markdown y decodifica el frontmatter en la estructura indicada.
package frontmatter

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

var delimiter = []byte("---")

// Split separa el frontmatter del cuerpo. El frontmatter debe empezar en la
// primera línea con "---" y terminar con otra línea "---". Si no hay
// frontmatter, devuelve fm=nil y body con el contenido completo.
func Split(data []byte) (fm, body []byte, err error) {
	// Elimina BOM y normaliza el final de línea inicial.
	data = bytes.TrimPrefix(data, []byte("\uFEFF"))

	rest := data
	if i := bytes.IndexByte(rest, '\n'); i >= 0 {
		if !bytes.Equal(bytes.TrimRight(rest[:i], "\r"), delimiter) {
			return nil, data, nil
		}
		rest = rest[i+1:]
	} else {
		return nil, data, nil
	}

	// Busca la línea de cierre del frontmatter.
	for len(rest) > 0 {
		lineEnd := bytes.IndexByte(rest, '\n')
		line := rest
		if lineEnd >= 0 {
			line = rest[:lineEnd]
		}
		if bytes.Equal(bytes.TrimRight(line, "\r"), delimiter) {
			after := []byte(nil)
			if lineEnd >= 0 {
				after = rest[lineEnd+1:]
			}
			return fm, after, nil
		}
		if lineEnd < 0 {
			break
		}
		fm = append(fm, bytes.TrimRight(line, "\r")...)
		fm = append(fm, '\n')
		rest = rest[lineEnd+1:]
	}

	return nil, data, fmt.Errorf("frontmatter: delimitador de cierre no encontrado")
}

// Unmarshal decodifica el frontmatter YAML en v.
func Unmarshal(fm []byte, v any) error {
	if len(bytes.TrimSpace(fm)) == 0 {
		return nil
	}
	return yaml.Unmarshal(fm, v)
}
