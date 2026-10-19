// Package verify comprueba que las referencias del sitio generado resuelven:
// assets, scripts, imágenes y enlaces internos.
package verify

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reAsset  = regexp.MustCompile(`<(?:link|script|img)\b[^>]*\b(?:href|src)="([^"]+)"`)
	reAnchor = regexp.MustCompile(`<a\b[^>]*\bhref="([^"]+)"`)
)

// Dist recorre el directorio de salida y devuelve las referencias que no resuelven.
func Dist(dir string) ([]string, error) {
	return DistBase(dir, "")
}

// DistBase es como Dist pero tolera que las URLs internas lleven un base path
// (por ejemplo /godian) y lo elimina antes de resolver contra dist/.
func DistBase(dir, basePath string) ([]string, error) {
	seen := map[string]bool{}
	var problems []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		page, _ := filepath.Rel(dir, path)

		refs := []string{}
		for _, m := range reAsset.FindAllStringSubmatch(string(data), -1) {
			refs = append(refs, m[1])
		}
		for _, m := range reAnchor.FindAllStringSubmatch(string(data), -1) {
			refs = append(refs, m[1])
		}

		for _, ref := range refs {
			if skip(ref) {
				continue
			}
			if resolve(dir, ref, basePath) {
				continue
			}
			key := filepath.ToSlash(page) + " -> " + ref
			if !seen[key] {
				seen[key] = true
				problems = append(problems, key)
			}
		}
		return nil
	})

	return problems, err
}

func skip(ref string) bool {
	switch {
	case ref == "":
		return true
	case strings.HasPrefix(ref, "#"),
		strings.HasPrefix(ref, "http://"),
		strings.HasPrefix(ref, "https://"),
		strings.HasPrefix(ref, "//"),
		strings.HasPrefix(ref, "mailto:"),
		strings.HasPrefix(ref, "tel:"),
		strings.HasPrefix(ref, "data:"),
		strings.HasPrefix(ref, "javascript:"):
		return true
	}
	return false
}

func resolve(root, ref, basePath string) bool {
	ref = strings.SplitN(ref, "#", 2)[0]
	ref = strings.SplitN(ref, "?", 2)[0]
	if ref == "" {
		return true
	}
	if basePath != "" {
		if ref == basePath {
			ref = "/"
		} else if strings.HasPrefix(ref, basePath+"/") {
			ref = strings.TrimPrefix(ref, basePath)
		}
	}
	clean := strings.TrimPrefix(filepath.ToSlash(ref), "/")
	full := filepath.Join(root, filepath.FromSlash(clean))
	return exists(full) || exists(filepath.Join(full, "index.html")) || exists(full+".html")
}

func exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
