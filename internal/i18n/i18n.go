// Package i18n contiene las traducciones de la interfaz y su carga desde archivos.
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// dictionaries mapea idioma → clave → texto (valores por defecto embebidos).
var dictionaries = map[string]map[string]string{
	"en": {
		"nav.home":            "Home",
		"nav.blog":            "Blog",
		"nav.docs":            "Docs",
		"nav.tags":            "Tags",
		"nav.categories":      "Categories",
		"nav.authors":         "Authors",
		"nav.series":          "Series",
		"nav.archive":         "Archive",
		"ui.search":           "Search",
		"ui.comments":         "Comments",
		"ui.toc":              "On this page",
		"ui.prev":             "← Prev",
		"ui.next":             "Next →",
		"ui.posts":            "post(s)",
		"ui.no_results":       "No results.",
		"ui.chapters":         "chapter(s)",
		"ui.min_read":         "min read",
		"ui.by":               "by",
		"ui.skip":             "Skip to content",
		"ui.rss":              "RSS",
		"ui.sitemap":          "Sitemap",
		"ui.share":            "Share",
		"ui.copy_link":        "Copy link",
		"ui.related":          "Related",
		"ui.backlinks":        "Backlinks",
		"ui.featured":         "Featured",
		"ui.stats":            "Stats",
		"ui.links":            "Links",
		"ui.all_posts":        "All posts",
		"ui.previous":         "← Previous",
		"ui.chapter":          "Chapter",
		"ui.website":          "Website",
		"ui.not_found":        "Page not found",
		"ui.no_series":        "No series yet.",
		"taxonomy.tags":       "Tags",
		"taxonomy.categories": "Categories",
	},
	"es": {
		"nav.home":            "Inicio",
		"nav.blog":            "Blog",
		"nav.docs":            "Documentación",
		"nav.tags":            "Etiquetas",
		"nav.categories":      "Categorías",
		"nav.authors":         "Autores",
		"nav.series":          "Series",
		"nav.archive":         "Archivo",
		"ui.search":           "Buscar",
		"ui.comments":         "Comentarios",
		"ui.toc":              "En esta página",
		"ui.prev":             "← Anterior",
		"ui.next":             "Siguiente →",
		"ui.posts":            "entrada(s)",
		"ui.no_results":       "Sin resultados.",
		"ui.chapters":         "capítulo(s)",
		"ui.min_read":         "min de lectura",
		"ui.by":               "por",
		"ui.skip":             "Saltar al contenido",
		"ui.rss":              "RSS",
		"ui.sitemap":          "Mapa del sitio",
		"ui.share":            "Compartir",
		"ui.copy_link":        "Copiar enlace",
		"ui.related":          "Relacionados",
		"ui.backlinks":        "Enlaces entrantes",
		"ui.featured":         "Destacado",
		"ui.stats":            "Estadísticas",
		"ui.links":            "Enlaces",
		"ui.all_posts":        "Todas las entradas",
		"ui.previous":         "← Anterior",
		"ui.chapter":          "Capítulo",
		"ui.website":          "Sitio web",
		"ui.not_found":        "Página no encontrada",
		"ui.no_series":        "Aún no hay series.",
		"taxonomy.tags":       "Etiquetas",
		"taxonomy.categories": "Categorías",
	},
}

// T devuelve la traducción de key en lang, cayendo a inglés y luego a key.
func T(lang, key string) string {
	if d, ok := dictionaries[lang]; ok {
		if v, ok := d[key]; ok {
			return v
		}
	}
	if v, ok := dictionaries["en"][key]; ok {
		return v
	}
	return key
}

// Keys devuelve todas las claves conocidas (unión de los diccionarios base).
func Keys() []string {
	set := map[string]struct{}{}
	for _, d := range dictionaries {
		for k := range d {
			set[k] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Builtin devuelve una copia del diccionario base del idioma (o inglés).
func Builtin(lang string) map[string]string {
	src := dictionaries["en"]
	if d, ok := dictionaries[lang]; ok {
		src = d
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// HasBuiltin indica si el idioma tiene diccionario embebido.
func HasBuiltin(lang string) bool {
	_, ok := dictionaries[lang]
	return ok
}

// Load carga i18n/<lang>.{json,yaml} sobre el diccionario base y devuelve avisos
// si falta el archivo o alguna clave requerida.
func Load(dir, lang string, required []string) (map[string]string, []string) {
	dict := Builtin(lang)
	var warnings []string

	var fileMap map[string]string
	path, found := findFile(dir, lang)
	if found {
		if data, err := os.ReadFile(path); err == nil {
			fileMap = map[string]string{}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".json":
				_ = json.Unmarshal(data, &fileMap)
			case ".yaml", ".yml":
				_ = yaml.Unmarshal(data, &fileMap)
			}
			for k, v := range fileMap {
				dict[k] = v
			}
		} else {
			warnings = append(warnings, "i18n: no se pudo leer "+path+": "+err.Error())
		}
	} else if dirExists(dir) {
		warnings = append(warnings, fmt.Sprintf("i18n: no hay %s; se usan los textos por defecto", filepath.Join(dir, lang+".json")))
	} else if !HasBuiltin(lang) {
		warnings = append(warnings, fmt.Sprintf("i18n: no hay traducción para %q; se usa inglés", lang))
	}

	// Conjunto de claves que deben existir (base + claves requeridas, p. ej. taxonomy.<name>).
	req := map[string]struct{}{}
	for _, k := range Keys() {
		req[k] = struct{}{}
	}
	for _, k := range required {
		if k != "" {
			req[k] = struct{}{}
		}
	}

	// Si hay archivo externo, debe estar completo; si no, se compara con los valores por defecto.
	var missing []string
	for k := range req {
		var ok bool
		if fileMap != nil {
			_, ok = fileMap[k]
		} else {
			_, ok = dict[k]
		}
		if !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		warnings = append(warnings, fmt.Sprintf("i18n: faltan %d claves en %q: %s", len(missing), lang, strings.Join(missing, ", ")))
	}
	return dict, warnings
}

func findFile(dir, lang string) (string, bool) {
	if dir == "" {
		return "", false
	}
	for _, ext := range []string{".json", ".yaml", ".yml"} {
		p := filepath.Join(dir, lang+ext)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func dirExists(dir string) bool {
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}
