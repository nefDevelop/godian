package markdown

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	kindWiki          = ast.NewNodeKind("ObsidianWiki")
	kindEmbed         = ast.NewNodeKind("ObsidianEmbed")
	kindHighlight     = ast.NewNodeKind("ObsidianHighlight")
	kindShortcode     = ast.NewNodeKind("Shortcode")
	kindCalloutMarker = ast.NewNodeKind("CalloutMarker")
)

// calloutMarkerNode marca el inicio de un callout ("[!tipo]").
type calloutMarkerNode struct {
	ast.BaseInline
	Variant   string
	Foldable  bool
	Collapsed bool
	Raw       string
}

func (*calloutMarkerNode) Kind() ast.NodeKind { return kindCalloutMarker }
func (n *calloutMarkerNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Variant": n.Variant, "Raw": n.Raw}, nil)
}

// wikiNode es un wikilink [[destino|alias]].
type wikiNode struct {
	ast.BaseInline
	Target string
	Label  string
}

func (*wikiNode) Kind() ast.NodeKind { return kindWiki }
func (n *wikiNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Target": n.Target, "Label": n.Label}, nil)
}

// embedNode es un embed ![[recurso]].
type embedNode struct {
	ast.BaseInline
	Target string
}

func (*embedNode) Kind() ast.NodeKind { return kindEmbed }
func (n *embedNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Target": n.Target}, nil)
}

// highlightNode es texto ==resaltado==.
type highlightNode struct {
	ast.BaseInline
	Value string
}

func (*highlightNode) Kind() ast.NodeKind { return kindHighlight }
func (n *highlightNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Value": n.Value}, nil)
}

// shortcodeNode es un shortcode {{nombre args}}.
type shortcodeNode struct {
	ast.BaseInline
	Name  string
	Args  []string
	Inner string
	Raw   string
}

func (*shortcodeNode) Kind() ast.NodeKind { return kindShortcode }
func (n *shortcodeNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Name": n.Name, "Raw": n.Raw}, nil)
}

// nodeText extrae el texto plano de un nodo recorriendo sus hijos, sin usar
// la API deprecada ast.Node.Text.
func nodeText(n ast.Node, src []byte) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
			if t.SoftLineBreak() {
				b.WriteByte('\n')
			}
		case *ast.String:
			b.Write(t.Value)
		default:
			b.WriteString(nodeText(c, src))
		}
	}
	return b.String()
}

// obsidianParser reconoce [[wikilinks]], ![[embeds]] y ==highlight==.
type obsidianParser struct{}

func (p *obsidianParser) Trigger() []byte { return []byte{'[', '!', '='} }

func (p *obsidianParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	switch {
	case bytes.HasPrefix(line, []byte("![[")):
		return parseEmbed(block, line)
	case bytes.HasPrefix(line, []byte("[[")):
		return parseWiki(block, line)
	case bytes.HasPrefix(line, []byte("[!")):
		return parseCalloutMarker(block, line)
	case bytes.HasPrefix(line, []byte("==")):
		return parseHighlight(block, line)
	}
	return nil
}

func parseCalloutMarker(block text.Reader, line []byte) ast.Node {
	closeIdx := bytes.IndexByte(line, ']')
	if closeIdx < 3 {
		return nil
	}
	variant := strings.ToLower(strings.TrimSpace(string(line[2:closeIdx])))
	if variant == "" {
		return nil
	}
	markerEnd := closeIdx + 1
	foldable, collapsed := false, false
	if markerEnd < len(line) && line[markerEnd] == '-' {
		foldable, collapsed = true, true
		markerEnd++
	} else if markerEnd < len(line) && line[markerEnd] == '+' {
		foldable = true
		markerEnd++
	}
	block.Advance(markerEnd)
	return &calloutMarkerNode{
		Variant: variant, Foldable: foldable, Collapsed: collapsed,
		Raw: string(line[:markerEnd]),
	}
}

func parseEmbed(block text.Reader, line []byte) ast.Node {
	end := bytes.Index(line, []byte("]]"))
	if end < 0 || end <= 3 {
		return nil
	}
	target := string(bytes.TrimSpace(line[3:end]))
	if target == "" {
		return nil
	}
	block.Advance(end + 2)
	return &embedNode{Target: target}
}

func parseWiki(block text.Reader, line []byte) ast.Node {
	end := bytes.Index(line, []byte("]]"))
	if end < 0 || end <= 2 {
		return nil
	}
	inner := string(line[2:end])
	target, label := inner, inner
	if i := strings.IndexByte(inner, '|'); i >= 0 {
		target, label = inner[:i], inner[i+1:]
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return nil
	}
	block.Advance(end + 2)
	return &wikiNode{Target: target, Label: strings.TrimSpace(label)}
}

func parseHighlight(block text.Reader, line []byte) ast.Node {
	rest := line[2:]
	end := bytes.Index(rest, []byte("=="))
	if end <= 0 {
		return nil
	}
	value := string(rest[:end])
	block.Advance(2 + end + 2)
	return &highlightNode{Value: value}
}

// shortcodeParser reconoce {{nombre args}}.
type shortcodeParser struct{}

func (p *shortcodeParser) Trigger() []byte { return []byte{'{'} }

func (p *shortcodeParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if !bytes.HasPrefix(line, []byte("{{")) {
		return nil
	}
	end := bytes.Index(line, []byte("}}"))
	if end < 0 {
		return nil
	}
	inner := strings.TrimSpace(string(line[2:end]))
	if inner == "" {
		return nil
	}
	fields := strings.Fields(inner)
	block.Advance(end + 2)
	return &shortcodeNode{Name: fields[0], Args: fields[1:], Inner: inner, Raw: string(line[:end+2])}
}

// tokenize divide respetando comillas simples/dobles.
func tokenize(s string) []string {
	var out []string
	var b strings.Builder
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				b.WriteByte(c)
			}
		case c == '"' || c == '\'':
			quote = c
		case c == ' ' || c == '\t' || c == '\n':
			if b.Len() > 0 {
				out = append(out, b.String())
				b.Reset()
			}
		default:
			b.WriteByte(c)
		}
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

// parseProps convierte ["clave=valor", ...] en un mapa.
func parseProps(tokens []string) map[string]string {
	props := map[string]string{}
	for _, t := range tokens {
		if i := strings.IndexByte(t, '='); i > 0 {
			props[t[:i]] = t[i+1:]
		}
	}
	return props
}
