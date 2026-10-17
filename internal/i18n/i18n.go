// Package i18n contiene las traducciones de la interfaz del tema.
package i18n

// dictionaries mapea idioma → clave → texto.
var dictionaries = map[string]map[string]string{
	"en": {
		"nav.home":       "Home",
		"nav.blog":       "Blog",
		"nav.docs":       "Docs",
		"nav.tags":       "Tags",
		"nav.categories": "Categories",
		"nav.authors":    "Authors",
		"nav.series":     "Series",
		"nav.archive":    "Archive",
		"ui.search":      "Search",
		"ui.comments":    "Comments",
		"ui.toc":         "On this page",
		"ui.prev":        "← Prev",
		"ui.next":        "Next →",
		"ui.posts":       "post(s)",
		"ui.no_results":  "No results.",
		"ui.chapters":    "chapter(s)",
		"ui.min_read":    "min read",
		"ui.by":          "by",
	},
	"es": {
		"nav.home":       "Inicio",
		"nav.blog":       "Blog",
		"nav.docs":       "Documentación",
		"nav.tags":       "Etiquetas",
		"nav.categories": "Categorías",
		"nav.authors":    "Autores",
		"nav.series":     "Series",
		"nav.archive":    "Archivo",
		"ui.search":      "Buscar",
		"ui.comments":    "Comentarios",
		"ui.toc":         "En esta página",
		"ui.prev":        "← Anterior",
		"ui.next":        "Siguiente →",
		"ui.posts":       "entrada(s)",
		"ui.no_results":  "Sin resultados.",
		"ui.chapters":    "capítulo(s)",
		"ui.min_read":    "min de lectura",
		"ui.by":          "por",
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
