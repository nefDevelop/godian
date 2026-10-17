// Package minify reduce el tamaño de HTML y CSS de forma conservadora.
// El JS se deja intacto: minificarlo de forma segura requiere un parser.
package minify

import (
	"bytes"
)

// JS devuelve el JavaScript sin cambios.
func JS(in []byte) []byte { return in }

var cssRawNext = map[byte]bool{'}': true, ';': true, ',': true, '{': true, ':': true}
var cssRawPrev = map[byte]bool{'{': true, '}': true, ';': true, ':': true, ',': true, '(': true}

// CSS elimina comentarios y espacios innecesarios, preservando cadenas.
func CSS(in []byte) []byte {
	var out bytes.Buffer
	i := 0
	for i < len(in) {
		c := in[i]
		if c == '/' && i+1 < len(in) && in[i+1] == '*' {
			j := bytes.Index(in[i+2:], []byte("*/"))
			if j < 0 {
				break
			}
			i += 2 + j + 2
			continue
		}
		if c == '"' || c == '\'' {
			out.WriteByte(c)
			i++
			for i < len(in) {
				out.WriteByte(in[i])
				if in[i] == '\\' && i+1 < len(in) {
					i++
					out.WriteByte(in[i])
					i++
					continue
				}
				if in[i] == c {
					i++
					break
				}
				i++
			}
			continue
		}
		if isSpace(c) {
			j := i
			for j < len(in) && isSpace(in[j]) {
				j++
			}
			var next byte
			if j < len(in) {
				next = in[j]
			}
			prev := lastByte(out.Bytes())
			if next == 0 || cssRawNext[next] || cssRawPrev[prev] {
				i = j
				continue
			}
			out.WriteByte(' ')
			i = j
			continue
		}
		out.WriteByte(c)
		i++
	}
	return bytes.TrimSpace(out.Bytes())
}

var rawElements = map[string]bool{
	"script": true, "style": true, "pre": true, "textarea": true,
}

// HTML elimina comentarios y colapsa espacios fuera de elementos crudos.
func HTML(in []byte) []byte {
	var out bytes.Buffer
	i := 0
	for i < len(in) {
		if in[i] == '<' {
			if bytes.HasPrefix(in[i:], []byte("<!--")) {
				j := bytes.Index(in[i+4:], []byte("-->"))
				if j < 0 {
					break
				}
				i += 4 + j + 3
				continue
			}
			j := bytes.IndexByte(in[i:], '>')
			if j < 0 {
				out.Write(in[i:])
				break
			}
			tag := in[i : i+j+1]
			out.Write(tag)
			i += j + 1

			name := tagName(tag)
			if rawElements[name] && !bytes.HasPrefix(tag, []byte("</")) && !bytes.HasSuffix(bytes.TrimSpace(tag), []byte("/>")) {
				idx := indexFold(in[i:], []byte("</"+name))
				if idx < 0 {
					out.Write(in[i:])
					i = len(in)
					continue
				}
				out.Write(in[i : i+idx])
				i += idx
			}
			continue
		}
		if isSpace(in[i]) {
			j := i
			for j < len(in) && isSpace(in[j]) {
				j++
			}
			out.WriteByte(' ')
			i = j
			continue
		}
		out.WriteByte(in[i])
		i++
	}
	return bytes.TrimSpace(out.Bytes())
}

func tagName(tag []byte) string {
	var b bytes.Buffer
	for i := 1; i < len(tag); i++ {
		c := tag[i]
		if c == '/' || c == '>' || isSpace(c) {
			continue
		}
		for ; i < len(tag); i++ {
			c = tag[i]
			if isSpace(c) || c == '>' || c == '/' {
				break
			}
			b.WriteByte(c)
		}
		break
	}
	return string(bytes.ToLower(b.Bytes()))
}

func indexFold(haystack, needle []byte) int {
	return bytes.Index(bytes.ToLower(haystack), bytes.ToLower(needle))
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}

func lastByte(b []byte) byte {
	if len(b) == 0 {
		return 0
	}
	return b[len(b)-1]
}
