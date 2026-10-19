package application

import (
	"path"
	"regexp"
	"strings"
)

// basePathOutput reescribe las URLs root-absolute al escribir la salida.
// Se aplica a HTML y CSS; el resto de formatos se prefijan en sus generadores.
type basePathOutput struct {
	out  Output
	base string
}

func (w basePathOutput) Write(rel string, data []byte) error {
	return w.out.Write(rel, rewriteBasePath(rel, data, w.base))
}

func (w basePathOutput) Clean() error {
	return w.out.Clean()
}

var (
	reBaseAttr        = regexp.MustCompile(`(\s(?:href|src|action|poster|data-src)=")(/[^"]*)`)
	reBaseSrcset      = regexp.MustCompile(`(\ssrcset=")([^"]*)`)
	reBaseMetaRefresh = regexp.MustCompile(`(content="[^"]*\burl=)(/[^"]*)`)
	reBaseCSSURL      = regexp.MustCompile(`url\((['"]?)(/[^)'"]*)(['"]?)\)`)
)

// rewriteBasePath prefija las URLs internas de HTML y CSS.
func rewriteBasePath(rel string, data []byte, base string) []byte {
	if base == "" {
		return data
	}
	switch strings.ToLower(path.Ext(rel)) {
	case ".html", ".htm":
		return rewriteHTMLBase(data, base)
	case ".css":
		return rewriteCSSBase(data, base)
	default:
		return data
	}
}

func rewriteHTMLBase(data []byte, base string) []byte {
	s := reBaseAttr.ReplaceAllStringFunc(string(data), func(m string) string {
		sub := reBaseAttr.FindStringSubmatch(m)
		return sub[1] + prefixRoot(sub[2], base)
	})
	s = reBaseSrcset.ReplaceAllStringFunc(s, func(m string) string {
		sub := reBaseSrcset.FindStringSubmatch(m)
		return sub[1] + prefixSrcset(sub[2], base)
	})
	s = reBaseMetaRefresh.ReplaceAllStringFunc(s, func(m string) string {
		sub := reBaseMetaRefresh.FindStringSubmatch(m)
		return sub[1] + prefixRoot(sub[2], base)
	})
	return []byte(s)
}

func rewriteCSSBase(data []byte, base string) []byte {
	return []byte(reBaseCSSURL.ReplaceAllStringFunc(string(data), func(m string) string {
		sub := reBaseCSSURL.FindStringSubmatch(m)
		quote := sub[1]
		if quote == "" {
			quote = sub[3]
		}
		return "url(" + quote + prefixRoot(sub[2], base) + quote + ")"
	}))
}

func prefixRoot(v, base string) string {
	if v == "" || !strings.HasPrefix(v, "/") || strings.HasPrefix(v, "//") {
		return v
	}
	if v == base || strings.HasPrefix(v, base+"/") {
		return v
	}
	if v == "/" {
		return base + "/"
	}
	return base + v
}

// prefixSrcset prefija cada URL de un srcset sin tocar los descriptores.
func prefixSrcset(v, base string) string {
	parts := strings.Split(v, ",")
	for i, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		fields[0] = prefixRoot(fields[0], base)
		parts[i] = strings.Join(fields, " ")
	}
	return strings.Join(parts, ", ")
}
