package markdown

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	kindCallout = ast.NewNodeKind("Callout")
	kindFigure  = ast.NewNodeKind("Figure")

	attrRe = regexp.MustCompile(`^\{(\.?[\w][\w-]*)\}$`)
)

// calloutNode es un callout estilo Obsidian.
type calloutNode struct {
	ast.BaseBlock
	Variant   string
	Title     string
	Foldable  bool
	Collapsed bool
}

func (*calloutNode) Kind() ast.NodeKind { return kindCallout }
func (n *calloutNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Variant": n.Variant, "Title": n.Title}, nil)
}

// figureNode envuelve una imagen con su caption.
type figureNode struct {
	ast.BaseBlock
	Alt string
}

func (*figureNode) Kind() ast.NodeKind { return kindFigure }
func (n *figureNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Alt": n.Alt}, nil)
}

// attributeTransformer aplica atributos de párrafos "{.clase}" o "{attr}" al
// bloque anterior (tablas, párrafos, etc.).
type attributeTransformer struct{}

func (t *attributeTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	processAttributes(doc, reader.Source())
}

func processAttributes(parent ast.Node, src []byte) {
	for n := parent.FirstChild(); n != nil; {
		next := n.NextSibling()
		if p, ok := n.(*ast.Paragraph); ok {
			if m := attrRe.FindStringSubmatch(strings.TrimSpace(nodeText(p, src))); m != nil {
				if prev := prevMeaningful(p, src); prev != nil {
					attr := m[1]
					if strings.HasPrefix(attr, ".") {
						addClass(prev, attr[1:])
					} else {
						prev.SetAttributeString(attr, "")
					}
					parent.RemoveChild(parent, p)
				}
			}
		}
		n = next
	}
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		processAttributes(c, src)
	}
}

func prevMeaningful(n ast.Node, src []byte) ast.Node {
	for prev := n.PreviousSibling(); prev != nil; prev = prev.PreviousSibling() {
		if t, ok := prev.(*ast.Text); ok && strings.TrimSpace(string(t.Segment.Value(src))) == "" {
			continue
		}
		return prev
	}
	return nil
}

func addClass(n ast.Node, cls string) {
	var classes []string
	if v, ok := n.AttributeString("class"); ok {
		switch t := v.(type) {
		case string:
			classes = strings.Fields(t)
		case []string:
			classes = append(classes, t...)
		}
	}
	classes = append(classes, cls)
	n.SetAttributeString("class", strings.Join(classes, " "))
}

// figureTransformer envuelve párrafos que solo contienen una imagen.
type figureTransformer struct{}

func (t *figureTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	processFigures(doc, reader.Source())
}

func processFigures(parent ast.Node, src []byte) {
	for n := parent.FirstChild(); n != nil; {
		next := n.NextSibling()
		if p, ok := n.(*ast.Paragraph); ok {
			if img, ok := soleImage(p, src); ok {
				fig := &figureNode{Alt: nodeText(img, src)}
				p.RemoveChild(p, img)
				fig.AppendChild(fig, img)
				parent.ReplaceChild(parent, p, fig)
			}
		}
		n = next
	}
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		processFigures(c, src)
	}
}

func soleImage(p *ast.Paragraph, src []byte) (*ast.Image, bool) {
	var img *ast.Image
	for c := p.FirstChild(); c != nil; c = c.NextSibling() {
		switch n := c.(type) {
		case *ast.Image:
			if img != nil {
				return nil, false
			}
			img = n
		case *ast.Text:
			if strings.TrimSpace(string(n.Segment.Value(src))) != "" {
				return nil, false
			}
		default:
			return nil, false
		}
	}
	return img, img != nil
}

// calloutTransformer convierte blockquotes "[!tipo]" en callouts.
type calloutTransformer struct{}

func (t *calloutTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	processCallouts(doc, reader.Source())
}

func processCallouts(parent ast.Node, src []byte) {
	for n := parent.FirstChild(); n != nil; {
		next := n.NextSibling()
		if bq, ok := n.(*ast.Blockquote); ok {
			if cl := toCallout(bq, src); cl != nil {
				parent.ReplaceChild(parent, bq, cl)
			}
		}
		n = next
	}
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		processCallouts(c, src)
	}
}

func toCallout(bq *ast.Blockquote, src []byte) *calloutNode {
	p, ok := bq.FirstChild().(*ast.Paragraph)
	if !ok {
		return nil
	}
	marker, ok := p.FirstChild().(*calloutMarkerNode)
	if !ok {
		return nil
	}

	// El título es el resto de la primera línea tras el marcador.
	var title string
	if t, ok := marker.NextSibling().(*ast.Text); ok {
		val := string(t.Segment.Value(src))
		if nl := strings.IndexByte(val, '\n'); nl >= 0 {
			title = strings.TrimSpace(val[:nl])
			t.Segment = text.NewSegment(t.Segment.Start+nl+1, t.Segment.Stop)
		} else {
			title = strings.TrimSpace(val)
			p.RemoveChild(p, t)
		}
	}
	p.RemoveChild(p, marker)
	if title == "" {
		title = capitalize(marker.Variant)
	}
	if p.ChildCount() == 0 {
		bq.RemoveChild(bq, p)
	}

	cl := &calloutNode{Variant: marker.Variant, Title: title, Foldable: marker.Foldable, Collapsed: marker.Collapsed}
	for c := bq.FirstChild(); c != nil; {
		next := c.NextSibling()
		bq.RemoveChild(bq, c)
		cl.AppendChild(cl, c)
		c = next
	}
	return cl
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
