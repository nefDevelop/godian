// Package check valida el contenido (imágenes, series, enlaces, metadatos).
package check

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
)

var (
	mdImageRe  = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)`)
	embedRe    = regexp.MustCompile(`!\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	srcAttrRe  = regexp.MustCompile(`src="([^"]+)"`)
	imagesRe   = regexp.MustCompile(`images="([^"]*)"`)
	fenceRe    = regexp.MustCompile("(?s)```.*?```")
	codeSpanRe = regexp.MustCompile("`[^`]*`")
)

// stripCode elimina bloques y spans de código para no analizar ejemplos.
func stripCode(s string) string {
	s = fenceRe.ReplaceAllString(s, "")
	s = codeSpanRe.ReplaceAllString(s, "")
	return s
}

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".avif": true, ".svg": true,
}

// Run valida el contenido y devuelve la lista de problemas encontrados.
func Run(cfg domain.Config, c domain.Content) []string {
	var problems []string
	add := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }

	postSlugs := map[string]bool{}
	for _, p := range c.Posts {
		postSlugs[p.Slug] = true
	}

	// Series: capítulos existentes.
	for _, s := range c.Series {
		for _, ch := range s.Chapters {
			if !postSlugs[ch] {
				add("serie %q: capítulo %q no existe", s.Slug, ch)
			}
		}
	}

	// Imágenes referenciadas + alt.
	referenced := map[string]bool{}
	collect := func(body string) {
		body = stripCode(body)
		for _, m := range mdImageRe.FindAllStringSubmatch(body, -1) {
			if strings.TrimSpace(m[1]) == "" {
				add("imagen sin alt: %s", m[2])
			}
			referenced[normalizeRef(m[2])] = true
		}
		for _, m := range embedRe.FindAllStringSubmatch(body, -1) {
			referenced[normalizeRef(m[1])] = true
		}
		for _, m := range srcAttrRe.FindAllStringSubmatch(body, -1) {
			referenced[normalizeRef(m[1])] = true
		}
		for _, m := range imagesRe.FindAllStringSubmatch(body, -1) {
			for _, img := range strings.Split(m[1], ",") {
				referenced[normalizeRef(img)] = true
			}
		}
	}
	for _, p := range c.Posts {
		collect(p.Body)
		referenced[normalizeRef(p.Image)] = true
	}
	for _, d := range c.Docs {
		collect(d.Body)
		referenced[normalizeRef(d.HeroImage)] = true
	}
	for _, pg := range c.Pages {
		collect(pg.Body)
	}
	for _, s := range c.Series {
		referenced[normalizeRef(s.Cover)] = true
	}

	// Assets existentes.
	assets := map[string]bool{}
	_ = filepath.WalkDir(cfg.Build.AssetsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(cfg.Build.AssetsDir, p)
		if rerr != nil {
			return nil
		}
		assets["/"+filepath.ToSlash(rel)] = true
		return nil
	})

	for ref := range referenced {
		if ref == "" || isExternal(ref) || !isImageExt(ref) {
			continue
		}
		if !assets[ref] {
			add("imagen referenciada no encontrada: %s", ref)
		}
	}
	for a := range assets {
		if isImageExt(a) && !referenced[a] {
			add("imagen sin usar (huérfana): %s", a)
		}
	}

	// Metadatos mínimos.
	for _, p := range c.Posts {
		if strings.TrimSpace(p.Title) == "" {
			add("post %q sin title", p.Slug)
		}
	}
	for _, d := range c.Docs {
		if strings.TrimSpace(d.Title) == "" {
			add("doc %q sin title", d.Slug)
		}
	}

	sort.Strings(problems)
	return problems
}

func normalizeRef(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.SplitN(ref, "#", 2)[0]
	ref = strings.SplitN(ref, "?", 2)[0]
	if ref == "" {
		return ""
	}
	if !strings.HasPrefix(ref, "/") {
		ref = "/" + ref
	}
	return ref
}

func isExternal(ref string) bool {
	for _, p := range []string{"http://", "https://", "//", "mailto:", "data:", "tel:"} {
		if strings.HasPrefix(ref, p) {
			return true
		}
	}
	return false
}

func isImageExt(ref string) bool {
	return imageExts[strings.ToLower(filepath.Ext(ref))]
}
