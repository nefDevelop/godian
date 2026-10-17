package application

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
)

var (
	reCodeFence  = regexp.MustCompile("(?s)```.*?```")
	reInlineHTML = regexp.MustCompile(`<[^>]+>`)
	reImage      = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
	reLink       = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	reWiki       = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
	reMarkers    = regexp.MustCompile("[#>*_`~]+")
	reSpaces     = regexp.MustCompile(`\s+`)
	reMath       = regexp.MustCompile(`\$\$?[^$]+\$\$?|\\\(.+?\\\)|\\\[.+?\\\]`)
)

// needsMath indica si el cuerpo contiene fórmulas (KaTeX).
func needsMath(body string) bool {
	return reMath.MatchString(body)
}

// plainText convierte Markdown en texto plano para el índice de búsqueda.
func plainText(md string) string {
	s := reCodeFence.ReplaceAllString(md, " ")
	s = reImage.ReplaceAllString(s, "$1")
	s = reLink.ReplaceAllString(s, "$1")
	s = reWiki.ReplaceAllStringFunc(s, func(m string) string {
		g := reWiki.FindStringSubmatch(m)
		if g[2] != "" {
			return g[2]
		}
		return g[1]
	})
	s = reInlineHTML.ReplaceAllString(s, " ")
	s = reMarkers.ReplaceAllString(s, " ")
	s = reSpaces.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) > 1200 {
		s = s[:1200]
	}
	return s
}

// pageSize es el número de posts por página en los listados.
const pageSize = 6

// cardOf construye la vista resumida de un post.
func cardOf(p domain.BlogPost) PostCard {
	return PostCard{
		Title:       p.Title,
		URL:         "/blog/" + p.Slug + "/",
		Description: p.Description,
		Date:        formatDate(p.Published),
		Tags:        p.Tags,
		ReadingTime: readingTime(p.Body),
	}
}

// readingMinutes estima el tiempo de lectura a 200 palabras por minuto.
func readingMinutes(body string) int {
	minutes := len(strings.Fields(body)) / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

// readingTime formatea el tiempo de lectura.
func readingTime(body string) string {
	return strconv.Itoa(readingMinutes(body)) + " min read"
}

// cardsOf convierte una lista de posts en tarjetas de vista.
func cardsOf(posts []domain.BlogPost) []PostCard {
	out := make([]PostCard, 0, len(posts))
	for _, p := range posts {
		out = append(out, cardOf(p))
	}
	return out
}

// firstN devuelve las primeras n tarjetas.
func firstN(cards []PostCard, n int) []PostCard {
	if len(cards) <= n {
		return cards
	}
	return cards[:n]
}

// countBy cuenta los posts que cumplen el predicado.
func countBy(posts []domain.BlogPost, match func(domain.BlogPost) bool) int {
	n := 0
	for _, p := range posts {
		if match(p) {
			n++
		}
	}
	return n
}

// sortStrings ordena una copia del slice.
func sortStrings(s []string) {
	sort.Strings(s)
}

// totalMinutes suma los minutos de lectura de los capítulos.
func totalMinutes(chapters []ChapterCard) int {
	total := 0
	for _, c := range chapters {
		total += c.Minutes
	}
	if total < 1 {
		total = 1
	}
	return total
}

// sortPinned ordena los posts con los fijados primero y luego por fecha desc.
func sortPinned(posts []domain.BlogPost) []domain.BlogPost {
	out := make([]domain.BlogPost, len(posts))
	copy(out, posts)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Pinned != out[j].Pinned {
			return out[i].Pinned
		}
		return out[i].Published.After(out[j].Published)
	})
	return out
}

// sortByDate ordena los posts por fecha descendente.
func sortByDate(posts []domain.BlogPost) []domain.BlogPost {
	out := make([]domain.BlogPost, len(posts))
	copy(out, posts)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Published.After(out[j].Published)
	})
	return out
}

// paginate divide items en páginas de tamaño size.
func paginate(posts []domain.BlogPost, size int) [][]domain.BlogPost {
	if len(posts) == 0 {
		return [][]domain.BlogPost{{}}
	}
	var pages [][]domain.BlogPost
	for start := 0; start < len(posts); start += size {
		end := start + size
		if end > len(posts) {
			end = len(posts)
		}
		pages = append(pages, posts[start:end])
	}
	return pages
}

// hasAuthor indica si un post pertenece a un autor.
func hasAuthor(p domain.BlogPost, slug string) bool {
	for _, a := range p.Authors {
		if a == slug {
			return true
		}
	}
	return false
}

// filterPosts devuelve los posts que cumplen el predicado, ordenados por fecha.
func filterPosts(posts []domain.BlogPost, match func(domain.BlogPost) bool) []domain.BlogPost {
	var out []domain.BlogPost
	for _, p := range posts {
		if match(p) {
			out = append(out, p)
		}
	}
	return sortByDate(out)
}

// tocItems convierte los encabezados en entradas de la tabla de contenidos.
func tocItems(headings []domain.Heading) []TOCItem {
	var out []TOCItem
	for _, h := range headings {
		if h.Level < 2 || h.Level > 3 || h.ID == "" {
			continue
		}
		out = append(out, TOCItem{Level: h.Level, ID: h.ID, Text: h.Text, URL: "#" + h.ID})
	}
	return out
}

// postTerms extrae los términos de un post según el campo de la taxonomía.
func postTerms(p domain.BlogPost, field string) []string {
	switch field {
	case "tags":
		return p.Tags
	case "categories":
		return p.Categories
	default:
		return nil
	}
}

// canonicalize normaliza un término aplicando alias.
func canonicalize(raw string, aliases map[string]string) string {
	if c, ok := aliases[raw]; ok {
		return c
	}
	return raw
}

// resolvedTerm es un término con sus metadatos.
type resolvedTerm struct {
	Name        string
	Description string
	Weight      int
}

// resolvedTaxonomy es una taxonomía con sus términos ya resueltos.
type resolvedTaxonomy struct {
	cfg     domain.TaxonomyConfig
	aliases map[string]string
	terms   []resolvedTerm
}

// postHas indica si un post pertenece al término indicado.
func (rt resolvedTaxonomy) postHas(p domain.BlogPost, name string) bool {
	for _, raw := range postTerms(p, rt.cfg.Field) {
		canon := canonicalize(raw, rt.aliases)
		if canon == name || (rt.cfg.Hierarchical && strings.HasPrefix(canon, name+"/")) {
			return true
		}
	}
	return false
}

// resolveTaxonomy fusiona los términos declarados en la config con los
// descubiertos en los posts, aplica alias y ordena por peso y nombre.
func resolveTaxonomy(cfg domain.TaxonomyConfig, posts []domain.BlogPost) (resolvedTaxonomy, []string) {
	aliases := map[string]string{}
	for _, t := range cfg.Terms {
		for _, a := range t.Aliases {
			aliases[a] = t.Name
		}
	}

	meta := map[string]resolvedTerm{}
	for _, t := range cfg.Terms {
		meta[t.Name] = resolvedTerm{Name: t.Name, Description: t.Description, Weight: t.Weight}
	}

	discovered := map[string]struct{}{}
	for _, p := range posts {
		for _, raw := range postTerms(p, cfg.Field) {
			discovered[canonicalize(raw, aliases)] = struct{}{}
		}
	}

	var warnings []string
	if cfg.Strict {
		names := make([]string, 0, len(discovered))
		for n := range discovered {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			if _, ok := meta[n]; !ok {
				warnings = append(warnings, fmt.Sprintf("taxonomía %q: término no declarado %q", cfg.Name, n))
			}
		}
	}

	for n := range discovered {
		if _, ok := meta[n]; !ok {
			meta[n] = resolvedTerm{Name: n}
		}
	}

	terms := make([]resolvedTerm, 0, len(meta))
	for _, t := range meta {
		terms = append(terms, t)
	}
	sort.Slice(terms, func(i, j int) bool {
		if terms[i].Weight != terms[j].Weight {
			return terms[i].Weight < terms[j].Weight
		}
		return terms[i].Name < terms[j].Name
	})

	return resolvedTaxonomy{cfg: cfg, aliases: aliases, terms: terms}, warnings
}
