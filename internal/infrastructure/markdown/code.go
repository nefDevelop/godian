package markdown

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

var (
	titleAttrRe = regexp.MustCompile(`title="([^"]*)"`)
	braceRe     = regexp.MustCompile(`\{([^}]*)\}`)
)

type codeAttrs struct {
	lineNumbers bool
	collapsible bool
	title       string
	highlight   map[int]bool
}

// parseCodeAttrs interpreta el info string de un bloque de código:
// "js {3,6-7}", "js wrap collapsible title=\"utils.js\"".
func parseCodeAttrs(info, lang string) codeAttrs {
	var a codeAttrs
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(info), lang))
	if m := titleAttrRe.FindStringSubmatch(rest); m != nil {
		a.title = m[1]
	}
	if strings.Contains(rest, "collapsible") {
		a.collapsible = true
	}
	if m := braceRe.FindStringSubmatch(rest); m != nil {
		a.lineNumbers = true
		a.highlight = parseRanges(m[1])
	}
	return a
}

// parseRanges interpreta "3,6-7" como {3:true,6:true,7:true}.
func parseRanges(spec string) map[int]bool {
	out := map[int]bool{}
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.IndexByte(part, '-'); i > 0 {
			from, err1 := strconv.Atoi(strings.TrimSpace(part[:i]))
			to, err2 := strconv.Atoi(strings.TrimSpace(part[i+1:]))
			if err1 != nil || err2 != nil {
				continue
			}
			for n := from; n <= to; n++ {
				out[n] = true
			}
			continue
		}
		if n, err := strconv.Atoi(part); err == nil {
			out[n] = true
		}
	}
	return out
}

// highlightLines marca con la clase "hl" las líneas indicadas del HTML de Chroma.
func highlightLines(html string, lines map[int]bool) string {
	if len(lines) == 0 {
		return html
	}
	const marker = `<span class="line">`
	var b strings.Builder
	n := 0
	for {
		i := strings.Index(html, marker)
		if i < 0 {
			b.WriteString(html)
			break
		}
		n++
		b.WriteString(html[:i])
		if lines[n] {
			b.WriteString(`<span class="line hl">`)
		} else {
			b.WriteString(marker)
		}
		html = html[i+len(marker):]
	}
	return b.String()
}

func (r *customRenderer) renderFencedCode(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.FencedCodeBlock)
	lang := string(n.Language(source))
	info := ""
	if n.Info != nil {
		info = string(n.Info.Segment.Value(source))
	}
	attrs := parseCodeAttrs(info, lang)

	var code strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		code.Write(seg.Value(source))
	}

	// Mermaid se renderiza en cliente. Se usa <pre> para que el minificador
	// de HTML preserve el contenido (los saltos de línea importan).
	if lang == "mermaid" {
		_, _ = w.WriteString(`<pre class="mermaid">` + escape(code.String()) + `</pre>`)
		return ast.WalkContinue, nil
	}

	body := highlightCode(code.String(), lang, attrs.lineNumbers)
	body = highlightLines(body, attrs.highlight)

	_, _ = w.WriteString(`<div class="code-block">`)
	_, _ = w.WriteString(`<div class="code-head"><span class="code-lang">` + escape(lang) + `</span>`)
	if attrs.title != "" {
		_, _ = w.WriteString(`<span class="code-title">` + escape(attrs.title) + `</span>`)
	}
	_, _ = w.WriteString(`<button class="code-copy" type="button" aria-label="Copy code">Copy</button></div>`)
	if attrs.collapsible {
		summary := lang
		if attrs.title != "" {
			summary = attrs.title
		}
		_, _ = w.WriteString(`<details class="code-collapsible"><summary>` + escape(summary) + `</summary>`)
		_, _ = w.WriteString(body)
		_, _ = w.WriteString(`</details>`)
	} else {
		_, _ = w.WriteString(body)
	}
	_, _ = w.WriteString(`</div>`)
	return ast.WalkContinue, nil
}

func highlightCode(code, lang string, lineNumbers bool) string {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	formatter := html.New(
		html.WithClasses(true),
		html.WithLineNumbers(lineNumbers),
	)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return `<pre class="chroma"><code>` + escape(code) + `</code></pre>`
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, styles.Get("github"), iterator); err != nil {
		return `<pre class="chroma"><code>` + escape(code) + `</code></pre>`
	}
	return buf.String()
}
