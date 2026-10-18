// Package application contiene los casos de uso del generador.
package application

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/i18n"
	"github.com/nefdevelop/godian/internal/infrastructure/minify"
)

// ContentLoader es el puerto de carga de contenido.
type ContentLoader interface {
	Load() (domain.Content, error)
}

// MarkdownRenderer convierte Markdown en HTML y extrae sus encabezados.
type MarkdownRenderer interface {
	Render(src string) (template.HTML, []domain.Heading, error)
}

// PageRenderer ejecuta plantillas del tema.
type PageRenderer interface {
	Render(w io.Writer, name string, data any) error
}

// Output escribe artefactos en el directorio de salida.
type Output interface {
	Write(relPath string, data []byte) error
	Clean() error
}

// ImageProcessor prepara las imágenes (variantes WebP) antes de renderizar.
type ImageProcessor interface {
	Process() error
}

// OGGenerator genera las imágenes Open Graph.
type OGGenerator interface {
	Generate(site domain.SiteConfig, posts []domain.BlogPost) error
}

// FeedGenerator construye RSS, Atom, JSON Feed y sitemap.
type FeedGenerator interface {
	RSS(site domain.SiteConfig, posts []domain.BlogPost) ([]byte, error)
	Atom(site domain.SiteConfig, posts []domain.BlogPost) ([]byte, error)
	JSONFeed(site domain.SiteConfig, posts []domain.BlogPost) ([]byte, error)
	Sitemaps(site domain.SiteConfig, paths []string) (map[string][]byte, error)
}

// Options agrupa parámetros de presentación del build.
type Options struct {
	Site          domain.SiteConfig
	Taxonomies    []domain.TaxonomyConfig
	Giscus        domain.GiscusConfig
	Nav           []domain.NavLink
	Social        []domain.SocialLink
	Support       []domain.SocialLink
	Analytics     domain.AnalyticsConfig
	Data          map[string]any
	Links         *LinkIndex
	I18n          map[string]string
	CSS           []string
	JS            []string
	AssetsFS      fs.FS
	Images        ImageProcessor
	OG            OGGenerator
	IncludeDrafts bool
	Minify        bool
	Fingerprint   bool
	Clean         bool

	// Fuentes autohospedadas (descargadas en el build): CSS de @font-face y
	// ficheros (ruta relativa al output → datos).
	FontsCSS  []byte
	FontFiles map[string][]byte
}

// Build orquesta la construcción del sitio estático.
type Build struct {
	content ContentLoader
	md      MarkdownRenderer
	render  PageRenderer
	output  Output
	feeds   FeedGenerator
	opts    Options
	nav     []Link
	css     []string
	js      []string
	now     time.Time
}

// NewBuild construye el caso de uso con sus puertos.
func NewBuild(content ContentLoader, md MarkdownRenderer, render PageRenderer, out Output, feeds FeedGenerator, opts Options) *Build {
	nav := navFor(opts.I18n, opts.Site.Language)
	if len(opts.Nav) > 0 {
		nav = navFromLinks(opts.Nav)
	}
	return &Build{content: content, md: md, render: render, output: out, feeds: feeds, opts: opts, nav: nav, css: opts.CSS, js: opts.JS}
}

// Run carga el contenido, renderiza todas las páginas y escribe la salida.
func (b *Build) Run() (Summary, error) {
	b.now = time.Now()

	c, err := b.content.Load()
	if err != nil {
		return Summary{}, err
	}
	if b.opts.Links != nil {
		b.opts.Links.Build(c.Posts, c.Docs)
	}

	// Limpia la salida para evitar artefactos obsoletos.
	if b.opts.Clean {
		if err := b.output.Clean(); err != nil {
			return Summary{}, err
		}
	}

	// Prepara imágenes (variantes WebP) antes de renderizar.
	if b.opts.Images != nil {
		if err := b.opts.Images.Process(); err != nil {
			return Summary{}, err
		}
	}

	// Prepara assets (fingerprint + minify) antes de renderizar.
	if b.opts.AssetsFS != nil {
		if err := b.prepareAssets(); err != nil {
			return Summary{}, err
		}
	}

	posts := b.published(c.Posts)

	var paths []string
	paths = append(paths, "/")

	// Páginas individuales de post.
	postPaths, err := b.renderPosts(posts, c.Series)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, postPaths...)

	// Home (primeras 6 entradas, con fijados primero).
	home := b.newPage(b.opts.Site.Title, b.opts.Site.Description)
	home.Posts = firstN(cardsOf(sortPinned(posts)), pageSize)
	home.Featured = featuredPost(posts)
	home.HomeTags = b.homeTerms(posts, "tags")
	home.HomeCats = b.homeTerms(posts, "categories")
	home.Stats = &SiteStats{Posts: len(posts), Tags: len(home.HomeTags), Categories: len(home.HomeCats)}
	if err := b.writeTemplate("home", "index.html", home); err != nil {
		return Summary{}, err
	}

	// Página 404.
	notFound := b.newPage("404", "")
	if err := b.writeTemplate("404", "404.html", notFound); err != nil {
		return Summary{}, err
	}

	// Blog paginado.
	blogPaths, err := b.renderList("blog", "/blog/", "Blog", "Thoughts, notes, and everything in between.", sortPinned(posts), nil)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, blogPaths...)

	// Taxonomías (tags, categorías, etc.) definidas en la config.
	taxPaths, warnings, err := b.renderTaxonomies(posts)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, taxPaths...)

	// Búsqueda (índice + página).
	if err := b.renderSearch(posts); err != nil {
		return Summary{}, err
	}
	paths = append(paths, "/search/")

	// Índice de fechas de contenido (para la isla calendar con source="site").
	if err := b.writeContentDates(posts, c.Docs); err != nil {
		return Summary{}, err
	}

	// Autores.
	authorPaths, err := b.renderAuthors(c, posts)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, authorPaths...)

	// Series.
	seriesPaths, err := b.renderSeries(c.Series, posts)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, seriesPaths...)

	// Archivo.
	archivePaths, err := b.renderArchive(posts)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, archivePaths...)

	// Documentación.
	docsPaths, err := b.renderDocs(c.Docs)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, docsPaths...)

	// Páginas sueltas (colección "pages", p. ej. "about").
	pagePaths, err := b.renderPages(c.Pages)
	if err != nil {
		return Summary{}, err
	}
	paths = append(paths, pagePaths...)

	// Imágenes Open Graph.
	if b.opts.OG != nil {
		if err := b.opts.OG.Generate(b.opts.Site, posts); err != nil {
			return Summary{}, err
		}
	}

	// favicon + manifest generados.
	if err := b.writeSiteMeta(); err != nil {
		return Summary{}, err
	}

	// Feeds.
	if err := b.writeFeeds(sortByDate(posts), paths); err != nil {
		return Summary{}, err
	}

	summary := NewSummary(c)
	summary.Published = len(posts)
	summary.Scheduled = b.scheduled(c.Posts)
	summary.Warnings = warnings
	return summary, nil
}

func (b *Build) renderPosts(posts []domain.BlogPost, series []domain.Series) ([]string, error) {
	ordered := sortByDate(posts) // más recientes primero
	bySlug := make(map[string]domain.BlogPost, len(posts))
	for _, p := range posts {
		bySlug[p.Slug] = p
	}
	seriesBySlug := make(map[string]domain.Series, len(series))
	for _, s := range series {
		seriesBySlug[s.Slug] = s
	}

	var paths []string
	for i, p := range ordered {
		body, headings, err := b.md.Render(p.Body)
		if err != nil {
			return nil, err
		}
		data := b.newPage(p.Title, p.Description)
		data.Body = body
		data.Date = formatDate(p.Published)
		data.Tags = p.Tags
		data.Categories = p.Categories
		data.ReadingTime = readingTime(p.Body)
		data.TOC = tocItems(headings)
		data.NeedsMath = needsMath(p.Body)
		data.OGImage = b.ogImage(p.Slug)
		data.JSONLD = b.jsonLD(p)
		data.Related = relatedPosts(p, ordered, 3)

		// prev = más antiguo, next = más reciente.
		if i+1 < len(ordered) {
			data.Prev = postNav(ordered[i+1])
		}
		if i > 0 {
			data.Next = postNav(ordered[i-1])
		}
		data.SeriesNav = seriesNav(p, seriesBySlug, bySlug)
		if b.opts.Links != nil {
			data.Backlinks = b.opts.Links.Backlinks(p.Slug)
		}

		url := "/blog/" + p.Slug + "/"
		if err := b.writeTemplate("post", strings.TrimPrefix(url, "/")+"index.html", data); err != nil {
			return nil, err
		}
		for _, alias := range p.Aliases {
			if err := b.writeRedirect(alias, url); err != nil {
				return nil, err
			}
		}
		paths = append(paths, url)
	}
	return paths, nil
}

// featuredPost devuelve el primer post fijado, si existe.
func featuredPost(posts []domain.BlogPost) *PostCard {
	for _, p := range sortPinned(posts) {
		if p.Pinned {
			c := cardOf(p)
			return &c
		}
	}
	return nil
}

// homeTerms devuelve los términos con conteo de una taxonomía, ordenados por uso.
func (b *Build) homeTerms(posts []domain.BlogPost, field string) []Term {
	var tc *domain.TaxonomyConfig
	for i := range b.opts.Taxonomies {
		if b.opts.Taxonomies[i].Field == field {
			tc = &b.opts.Taxonomies[i]
			break
		}
	}
	if tc == nil {
		return nil
	}
	rt, _ := resolveTaxonomy(*tc, posts)
	base := strings.TrimSuffix(tc.URL, "/") + "/"
	var terms []Term
	for _, t := range rt.terms {
		count := countBy(posts, func(p domain.BlogPost) bool { return rt.postHas(p, t.Name) })
		if count == 0 {
			continue
		}
		terms = append(terms, Term{Name: t.Name, URL: base + t.Name + "/", Count: count})
	}
	sort.SliceStable(terms, func(i, j int) bool {
		if terms[i].Count != terms[j].Count {
			return terms[i].Count > terms[j].Count
		}
		return terms[i].Name < terms[j].Name
	})
	return terms
}

func postNav(p domain.BlogPost) *PostNav {
	return &PostNav{Title: p.Title, URL: "/blog/" + p.Slug + "/"}
}

// relatedPosts devuelve hasta n posts con más tags/categorías en común.
func relatedPosts(p domain.BlogPost, all []domain.BlogPost, n int) []PostCard {
	type scored struct {
		post  domain.BlogPost
		score int
	}
	var list []scored
	for _, o := range all {
		if o.Slug == p.Slug {
			continue
		}
		score := 0
		for _, t := range p.Tags {
			for _, ot := range o.Tags {
				if t == ot {
					score++
				}
			}
		}
		for _, c := range p.Categories {
			for _, oc := range o.Categories {
				if c == oc {
					score++
				}
			}
		}
		if score > 0 {
			list = append(list, scored{post: o, score: score})
		}
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		return list[i].post.Published.After(list[j].post.Published)
	})
	if len(list) > n {
		list = list[:n]
	}
	out := make([]PostCard, 0, len(list))
	for _, s := range list {
		out = append(out, cardOf(s.post))
	}
	return out
}

// seriesNav calcula la navegación de un post dentro de su serie.
func seriesNav(p domain.BlogPost, series map[string]domain.Series, bySlug map[string]domain.BlogPost) *SeriesNav {
	if p.Series == "" {
		return nil
	}
	s, ok := series[p.Series]
	if !ok {
		return nil
	}
	idx := -1
	for i, slug := range s.Chapters {
		if slug == p.Slug {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	sn := &SeriesNav{Title: s.Title, URL: "/series/" + s.Slug + "/", Index: idx + 1, Total: len(s.Chapters)}
	if idx > 0 {
		if q, ok := bySlug[s.Chapters[idx-1]]; ok {
			sn.Prev = postNav(q)
		}
	}
	if idx < len(s.Chapters)-1 {
		if q, ok := bySlug[s.Chapters[idx+1]]; ok {
			sn.Next = postNav(q)
		}
	}
	return sn
}

// renderList genera un listado paginado con la plantilla indicada.
func (b *Build) renderList(tmpl, base, title, desc string, posts []domain.BlogPost, decorate func(*PageData, string)) ([]string, error) {
	pages := paginate(posts, pageSize)
	var paths []string
	for i, chunk := range pages {
		page := i + 1
		data := b.newPage(title, desc)
		data.Posts = cardsOf(chunk)
		data.Page = newPagination(base, page, len(pages))
		data.Total = len(posts)
		url := pageURL(base, page)
		if decorate != nil {
			decorate(&data, url)
		}
		rel := strings.TrimPrefix(url, "/") + "index.html"
		if err := b.writeTemplate(tmpl, rel, data); err != nil {
			return nil, err
		}
		paths = append(paths, url)
	}
	return paths, nil
}

// renderTaxonomies genera los índices y las páginas de cada termino para todas
// las taxonomías declaradas en la config.
func (b *Build) renderTaxonomies(posts []domain.BlogPost) ([]string, []string, error) {
	var paths, warnings []string
	for _, tc := range b.opts.Taxonomies {
		rt, w := resolveTaxonomy(tc, posts)
		warnings = append(warnings, w...)

		base := strings.TrimSuffix(tc.URL, "/") + "/"

		label := tc.Label
		if v := b.opts.I18n["taxonomy."+tc.Name]; v != "" {
			label = v
		}

		index := b.newPage(label, "")
		for _, t := range rt.terms {
			count := countBy(posts, func(p domain.BlogPost) bool { return rt.postHas(p, t.Name) })
			// Solo se lista si su página se va a generar (evita enlaces rotos).
			if count == 0 && !tc.RenderEmpty {
				continue
			}
			index.Terms = append(index.Terms, Term{
				Name:  t.Name,
				URL:   base + t.Name + "/",
				Count: count,
			})
		}
		if err := b.writeTemplate("terms", strings.TrimPrefix(base, "/")+"index.html", index); err != nil {
			return nil, nil, err
		}
		paths = append(paths, base)

		for _, t := range rt.terms {
			count := countBy(posts, func(p domain.BlogPost) bool { return rt.postHas(p, t.Name) })
			if count == 0 && !tc.RenderEmpty {
				continue
			}
			filtered := filterPosts(posts, func(p domain.BlogPost) bool { return rt.postHas(p, t.Name) })
			title := tc.Prefix + t.Name
			desc := t.Description
			if desc == "" {
				desc = "Posts in " + tc.Name + ": " + t.Name
			}
			termBase := base + t.Name + "/"
			termPaths, err := b.renderList("taxonomy", termBase, title, desc, filtered, func(d *PageData, u string) {
				d.JSONLD = b.collectionPageJSONLD(title, u)
				d.FeedURL = termBase + "rss.xml"
			})
			if err != nil {
				return nil, nil, err
			}
			paths = append(paths, termPaths...)

			// Feed por término.
			feedSite := b.opts.Site
			feedSite.Title = b.opts.Site.Title + " · " + title
			rss, err := b.feeds.RSS(feedSite, filtered)
			if err != nil {
				return nil, nil, err
			}
			if err := b.output.Write(strings.TrimPrefix(termBase, "/")+"rss.xml", rss); err != nil {
				return nil, nil, err
			}
		}
	}
	return paths, warnings, nil
}

// SearchDoc es un documento del índice de búsqueda.
type SearchDoc struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Text        string   `json:"text"`
}

func (b *Build) renderSearch(posts []domain.BlogPost) error {
	lang := b.opts.Site.Language
	if lang == "" {
		lang = "en"
	}
	data := b.newPage(i18n.T(lang, "ui.search"), "Search.")
	if err := b.writeTemplate("search", "search/index.html", data); err != nil {
		return err
	}

	docs := make([]SearchDoc, 0, len(posts))
	for _, p := range posts {
		docs = append(docs, SearchDoc{
			Title:       p.Title,
			URL:         "/blog/" + p.Slug + "/",
			Description: p.Description,
			Tags:        p.Tags,
			Text:        plainText(p.Body),
		})
	}
	encoded, err := json.Marshal(docs)
	if err != nil {
		return err
	}
	return b.output.Write("search-index.json", encoded)
}

// ContentDate es una entrada del índice de fechas (para la isla calendar).
type ContentDate struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"` // post | doc
}

// writeContentDates genera content-dates.json con las fechas de posts y docs.
func (b *Build) writeContentDates(posts []domain.BlogPost, docs []domain.Doc) error {
	dates := make([]ContentDate, 0, len(posts)+len(docs))
	for _, p := range posts {
		dates = append(dates, ContentDate{
			Date:  p.Published.Format("2006-01-02"),
			Title: p.Title,
			URL:   "/blog/" + p.Slug + "/",
			Type:  "post",
		})
	}
	for _, d := range docs {
		if d.PubDate == nil {
			continue
		}
		dates = append(dates, ContentDate{
			Date:  d.PubDate.Format("2006-01-02"),
			Title: d.Title,
			URL:   "/docs/" + d.Slug + "/",
			Type:  "doc",
		})
	}
	encoded, err := json.Marshal(dates)
	if err != nil {
		return err
	}
	return b.output.Write("content-dates.json", encoded)
}

func (b *Build) renderAuthors(c domain.Content, posts []domain.BlogPost) ([]string, error) {
	info := map[string]domain.Author{}
	for _, a := range c.Authors {
		info[a.Slug] = a
	}

	set := map[string]struct{}{}
	var slugs []string
	for _, p := range posts {
		for _, slug := range p.Authors {
			if _, ok := set[slug]; !ok {
				set[slug] = struct{}{}
				slugs = append(slugs, slug)
			}
		}
	}
	sortStrings(slugs)

	index := b.newPage("Authors", "Browse posts by author.")
	for _, slug := range slugs {
		count := countBy(posts, func(p domain.BlogPost) bool { return hasAuthor(p, slug) })
		card := AuthorCard{Name: slug, Slug: slug, URL: "/authors/" + slug + "/", Count: count}
		if a, ok := info[slug]; ok {
			card.Name = a.Name
			card.Avatar = a.Avatar
			card.Bio = a.Bio
		}
		index.Authors = append(index.Authors, card)
	}
	if err := b.writeTemplate("authors", "authors/index.html", index); err != nil {
		return nil, err
	}
	paths := []string{"/authors/"}

	for _, slug := range slugs {
		filtered := filterPosts(posts, func(p domain.BlogPost) bool { return hasAuthor(p, slug) })
		base := "/authors/" + slug + "/"
		a := info[slug]
		name := slug
		if a.Name != "" {
			name = a.Name
		}
		profile := &AuthorProfile{Name: name, Slug: slug, Avatar: a.Avatar, Bio: a.Bio, Website: a.Website, Github: a.Github, Twitter: a.Twitter, Count: len(filtered)}

		pages := paginate(filtered, pageSize)
		for i, chunk := range pages {
			page := i + 1
			data := b.newPage(name, "Posts by "+name)
			data.Author = profile
			data.Posts = cardsOf(chunk)
			data.Page = newPagination(base, page, len(pages))
			data.Total = len(filtered)
			url := pageURL(base, page)
			data.JSONLD = b.personJSONLD(a, url)
			rel := strings.TrimPrefix(url, "/") + "index.html"
			if err := b.writeTemplate("author", rel, data); err != nil {
				return nil, err
			}
			paths = append(paths, url)
		}
	}
	return paths, nil
}

func (b *Build) renderSeries(series []domain.Series, posts []domain.BlogPost) ([]string, error) {
	bySlug := map[string]domain.BlogPost{}
	for _, p := range posts {
		bySlug[p.Slug] = p
	}

	index := b.newPage("Series", "Browse posts organized as book-like chapters.")
	var details []SeriesDetail
	for _, s := range series {
		var chapters []ChapterCard
		for i, slug := range s.Chapters {
			p, ok := bySlug[slug]
			if !ok {
				continue
			}
			chapters = append(chapters, ChapterCard{
				Number: i + 1, Title: p.Title, URL: "/blog/" + p.Slug + "/",
				Description: p.Description, Date: formatDate(p.Published),
				ReadingTime: readingTime(p.Body), Minutes: readingMinutes(p.Body),
			})
		}
		minutes := totalMinutes(chapters)
		index.SeriesList = append(index.SeriesList, SeriesCard{
			Title: s.Title, Slug: s.Slug, URL: "/series/" + s.Slug + "/",
			Cover: s.Cover, Description: s.Description, Chapters: len(chapters), Minutes: minutes,
		})
		details = append(details, SeriesDetail{
			Title: s.Title, Slug: s.Slug, Cover: s.Cover, Description: s.Description,
			Authors: s.Authors, Chapters: chapters, Minutes: minutes,
		})
	}
	if err := b.writeTemplate("series", "series/index.html", index); err != nil {
		return nil, err
	}
	paths := []string{"/series/"}

	for _, d := range details {
		data := b.newPage(d.Title+" — Series", d.Description)
		detail := d
		data.Series = &detail
		if err := b.writeTemplate("series-detail", "series/"+d.Slug+"/index.html", data); err != nil {
			return nil, err
		}
		paths = append(paths, "/series/"+d.Slug+"/")
	}
	return paths, nil
}

// renderDocs genera el índice de docs y cada página con su navegación lateral.
func (b *Build) renderDocs(docs []domain.Doc) ([]string, error) {
	sections := map[string][]domain.Doc{}
	var order []string
	for _, d := range docs {
		sec := d.Section
		if sec == "" {
			sec = "general"
		}
		if _, ok := sections[sec]; !ok {
			order = append(order, sec)
		}
		sections[sec] = append(sections[sec], d)
	}
	sort.Strings(order)
	for _, sec := range order {
		list := sections[sec]
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Order != list[j].Order {
				return list[i].Order < list[j].Order
			}
			return list[i].Title < list[j].Title
		})
		sections[sec] = list
	}

	buildNav := func(activeSlug string) []DocSection {
		nav := make([]DocSection, 0, len(order))
		for _, sec := range order {
			ds := DocSection{Section: sec}
			for _, d := range sections[sec] {
				ds.Docs = append(ds.Docs, DocLink{
					Title:  d.Title,
					URL:    "/docs/" + d.Slug + "/",
					Active: d.Slug == activeSlug,
				})
			}
			nav = append(nav, ds)
		}
		return nav
	}

	// Índice.
	index := b.newPage(i18n.T(langOf(b.opts.Site), "nav.docs"), "Documentation.")
	index.DocsNav = buildNav("")
	index.Breadcrumb = []Link{{Label: i18n.T(langOf(b.opts.Site), "nav.home"), URL: "/"}}
	if err := b.writeTemplate("docs-index", "docs/index.html", index); err != nil {
		return nil, err
	}
	paths := []string{"/docs/"}

	// Páginas.
	for _, sec := range order {
		for _, d := range sections[sec] {
			body, headings, err := b.md.Render(d.Body)
			if err != nil {
				return nil, err
			}
			data := b.newPage(d.Title, d.Description)
			data.Body = body
			data.TOC = tocItems(headings)
			data.DocsNav = buildNav(d.Slug)
			if b.opts.Links != nil {
				data.Backlinks = b.opts.Links.Backlinks(d.Slug)
			}
			data.Breadcrumb = []Link{
				{Label: i18n.T(langOf(b.opts.Site), "nav.home"), URL: "/"},
				{Label: i18n.T(langOf(b.opts.Site), "nav.docs"), URL: "/docs/"},
				{Label: d.Title, URL: "/docs/" + d.Slug + "/"},
			}
			url := "/docs/" + d.Slug + "/"
			if err := b.writeTemplate("docs", strings.TrimPrefix(url, "/")+"index.html", data); err != nil {
				return nil, err
			}
			for _, alias := range d.Aliases {
				if err := b.writeRedirect(alias, url); err != nil {
					return nil, err
				}
			}
			paths = append(paths, url)
		}
	}
	return paths, nil
}

// renderPages renderiza la colección "pages" (páginas sueltas) en /<slug>/.
func (b *Build) renderPages(pages []domain.Page) ([]string, error) {
	var paths []string
	for _, p := range pages {
		body, headings, err := b.md.Render(p.Body)
		if err != nil {
			return nil, err
		}
		data := b.newPage(p.Title, p.Description)
		data.Body = body
		data.Stack = p.Stack
		data.TOC = tocItems(headings)
		data.Breadcrumb = []Link{
			{Label: i18n.T(langOf(b.opts.Site), "nav.home"), URL: "/"},
			{Label: p.Title, URL: "/" + p.Slug + "/"},
		}
		if b.opts.Links != nil {
			data.Backlinks = b.opts.Links.Backlinks(p.Slug)
		}
		url := "/" + p.Slug + "/"
		if err := b.writeTemplate("page", strings.TrimPrefix(url, "/")+"index.html", data); err != nil {
			return nil, err
		}
		paths = append(paths, url)
	}
	return paths, nil
}

func langOf(site domain.SiteConfig) string {
	if site.Language == "" {
		return "en"
	}
	return site.Language
}

func (b *Build) renderArchive(posts []domain.BlogPost) ([]string, error) {
	sorted := sortByDate(posts)
	data := b.newPage("Archive", "All posts by year.")
	var years []ArchiveYear
	var current *ArchiveYear
	for _, p := range sorted {
		y := p.Published.Year()
		if current == nil || current.Year != y {
			years = append(years, ArchiveYear{Year: y})
			current = &years[len(years)-1]
		}
		cat := ""
		if len(p.Categories) > 0 {
			cat = p.Categories[0]
		}
		current.Posts = append(current.Posts, ArchivePost{
			Title: p.Title, URL: "/blog/" + p.Slug + "/", Date: p.Published.Format("01-02"), Category: cat,
		})
	}
	data.Years = years
	if err := b.writeTemplate("archive", "archive/index.html", data); err != nil {
		return nil, err
	}
	return []string{"/archive/"}, nil
}

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{3,8}$`)

// writeSiteMeta genera favicon.svg y site.webmanifest a partir de la config.
func (b *Build) writeSiteMeta() error {
	title := b.opts.Site.Title
	if title == "" {
		title = "godian"
	}
	accent := b.opts.Site.Color
	if !hexColorRe.MatchString(accent) {
		accent = "#2563eb"
	}

	initial := "G"
	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			initial = strings.ToUpper(string(r))
			break
		}
	}

	favicon := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<rect width="64" height="64" rx="14" fill="` + accent + `"/>` +
		`<text x="32" y="43" font-family="system-ui,-apple-system,Segoe UI,Roboto,sans-serif" ` +
		`font-size="36" font-weight="700" fill="#ffffff" text-anchor="middle">` + initial + `</text></svg>`
	if err := b.output.Write("favicon.svg", []byte(favicon)); err != nil {
		return err
	}

	manifest, err := json.Marshal(map[string]any{
		"name":             title,
		"short_name":       title,
		"start_url":        "/",
		"display":          "standalone",
		"background_color": "#ffffff",
		"theme_color":      accent,
		"icons": []map[string]any{
			{"src": "/favicon.svg", "sizes": "any", "type": "image/svg+xml"},
		},
	})
	if err != nil {
		return err
	}
	return b.output.Write("site.webmanifest", manifest)
}

// collectionPageJSONLD describe una página de listado (tag/categoría).
func (b *Build) collectionPageJSONLD(name, pageURL string) template.JS {
	base := strings.TrimRight(b.opts.Site.URL, "/")
	doc := map[string]any{
		"@context": "https://schema.org",
		"@type":    "CollectionPage",
		"name":     name,
		"url":      base + pageURL,
		"isPartOf": map[string]any{"@type": "WebSite", "url": base + "/"},
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return template.JS(data)
}

// personJSONLD describe el perfil de un autor.
func (b *Build) personJSONLD(a domain.Author, pageURL string) template.JS {
	base := strings.TrimRight(b.opts.Site.URL, "/")
	doc := map[string]any{
		"@context": "https://schema.org",
		"@type":    "Person",
		"name":     a.Name,
		"url":      base + pageURL,
	}
	if a.Avatar != "" {
		doc["image"] = a.Avatar
	}
	var same []string
	for _, u := range []string{a.Github, a.Twitter, a.Website} {
		if u != "" {
			same = append(same, u)
		}
	}
	if len(same) > 0 {
		doc["sameAs"] = same
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return template.JS(data)
}

func (b *Build) ogImage(slug string) string {
	base := strings.TrimRight(b.opts.Site.URL, "/")
	return base + "/og/" + slug + ".png"
}

func (b *Build) jsonLD(p domain.BlogPost) template.JS {
	author := ""
	if len(p.Authors) > 0 {
		author = p.Authors[0]
	}
	doc := map[string]any{
		"@context":      "https://schema.org",
		"@type":         "BlogPosting",
		"headline":      p.Title,
		"description":   p.Description,
		"datePublished": p.Published.Format(time.RFC3339),
		"author":        map[string]any{"@type": "Person", "name": author},
		"keywords":      strings.Join(p.Tags, ", "),
		"url":           strings.TrimRight(b.opts.Site.URL, "/") + "/blog/" + p.Slug + "/",
	}
	if p.Updated != nil {
		doc["dateModified"] = p.Updated.Format(time.RFC3339)
	}
	if b.opts.Site.URL != "" {
		doc["image"] = b.ogImage(p.Slug)
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return template.JS(data)
}

func (b *Build) writeTemplate(name, relPath string, data PageData) error {
	data.PageURL = urlFromRel(relPath)
	data.NavActive = b.navActive(data.PageURL)
	if b.opts.Site.URL != "" {
		data.Canonical = strings.TrimRight(b.opts.Site.URL, "/") + data.PageURL
	}
	var buf bytes.Buffer
	if err := b.render.Render(&buf, name, data); err != nil {
		return err
	}
	out := buf.Bytes()
	if b.opts.Minify {
		out = minify.HTML(out)
	}
	return b.output.Write(relPath, out)
}

func (b *Build) writeFeeds(posts []domain.BlogPost, paths []string) error {
	rss, err := b.feeds.RSS(b.opts.Site, posts)
	if err != nil {
		return err
	}
	if err := b.output.Write("rss.xml", rss); err != nil {
		return err
	}

	atom, err := b.feeds.Atom(b.opts.Site, posts)
	if err != nil {
		return err
	}
	if err := b.output.Write("atom.xml", atom); err != nil {
		return err
	}

	jsonFeed, err := b.feeds.JSONFeed(b.opts.Site, posts)
	if err != nil {
		return err
	}
	if err := b.output.Write("feed.json", jsonFeed); err != nil {
		return err
	}

	sitemaps, err := b.feeds.Sitemaps(b.opts.Site, paths)
	if err != nil {
		return err
	}
	for name, data := range sitemaps {
		if err := b.output.Write(name, data); err != nil {
			return err
		}
	}

	return b.writeRobots()
}

// writeRedirect genera una página de redirección (meta refresh) para una URL antigua.
func (b *Build) writeRedirect(alias, target string) error {
	alias = strings.Trim(strings.TrimSpace(alias), "/")
	if alias == "" {
		return nil
	}
	canonical := ""
	if b.opts.Site.URL != "" {
		canonical = strings.TrimRight(b.opts.Site.URL, "/") + target
	}
	html := `<!doctype html><html><head><meta charset="utf-8">` +
		`<meta name="robots" content="noindex">` +
		`<title>Redirecting…</title>` +
		`<link rel="canonical" href="` + canonical + `">` +
		`<meta http-equiv="refresh" content="0; url=` + target + `">` +
		`</head><body><p>Redirecting to <a href="` + target + `">` + target + `</a>…</p></body></html>`
	return b.output.Write(alias+"/index.html", []byte(html))
}

// writeRobots genera robots.txt (permite todo y apunta al sitemap).
func (b *Build) writeRobots() error {
	var sb strings.Builder
	sb.WriteString("User-agent: *\nAllow: /\n")
	if base := strings.TrimRight(b.opts.Site.URL, "/"); base != "" {
		sb.WriteString("\nSitemap: " + base + "/sitemap.xml\n")
	}
	return b.output.Write("robots.txt", []byte(sb.String()))
}

func (b *Build) newPage(title, desc string) PageData {
	og := ""
	if b.opts.Site.URL != "" {
		og = strings.TrimRight(b.opts.Site.URL, "/") + "/og/default.png"
	}
	return PageData{
		Site:        b.opts.Site,
		Title:       title,
		Description: desc,
		Nav:         b.nav,
		CSS:         b.css,
		JS:          b.js,
		Giscus:      b.opts.Giscus,
		Social:      b.opts.Social,
		Support:     b.opts.Support,
		Analytics:   b.opts.Analytics,
		Data:        b.opts.Data,
		OGImage:     og,
		SiteJSONLD:  b.siteJSONLD(),
		Year:        time.Now().Year(),
	}
}

// siteJSONLD construye el JSON-LD WebSite (con SearchAction) para todas las páginas.
func (b *Build) siteJSONLD() template.JS {
	if b.opts.Site.URL == "" {
		return ""
	}
	base := strings.TrimRight(b.opts.Site.URL, "/")
	doc := map[string]any{
		"@context": "https://schema.org",
		"@type":    "WebSite",
		"name":     b.opts.Site.Title,
		"url":      base,
		"potentialAction": map[string]any{
			"@type":       "SearchAction",
			"target":      base + "/search/?q={search_term_string}",
			"query-input": "required name=search_term_string",
		},
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return template.JS(data)
}

// urlFromRel deriva la URL pública de un archivo de salida.
func urlFromRel(relPath string) string {
	rel := strings.TrimPrefix(filepath.ToSlash(relPath), "./")
	if rel == "index.html" {
		return "/"
	}
	if strings.HasSuffix(rel, "/index.html") {
		return "/" + strings.TrimSuffix(rel, "index.html")
	}
	return "/" + rel
}

// navActive devuelve la entrada de nav que corresponde a la URL actual.
func (b *Build) navActive(pageURL string) string {
	best := ""
	for _, l := range b.nav {
		if l.URL == "/" {
			if pageURL == "/" {
				return "/"
			}
			continue
		}
		if strings.HasPrefix(pageURL, l.URL) && len(l.URL) > len(best) {
			best = l.URL
		}
	}
	return best
}

// prepareAssets aplica fingerprint (si procede) y copia los assets del tema.
func (b *Build) prepareAssets() error {
	if b.opts.Fingerprint {
		css, err := b.fingerprint(b.opts.CSS)
		if err != nil {
			return err
		}
		js, err := b.fingerprint(b.opts.JS)
		if err != nil {
			return err
		}
		b.css, b.js = css, js
	}
	if err := b.copyAssets(); err != nil {
		return err
	}
	if err := b.writeFonts(); err != nil {
		return err
	}
	return b.writeIslandManifest()
}

// writeFonts escribe las fuentes autohospedadas y su CSS (opcional).
func (b *Build) writeFonts() error {
	if len(b.opts.FontFiles) == 0 {
		return nil
	}
	for p, data := range b.opts.FontFiles {
		if err := b.output.Write(p, data); err != nil {
			return err
		}
	}
	if len(b.opts.FontsCSS) == 0 {
		return nil
	}
	css := b.minifyAsset("fonts.css", b.opts.FontsCSS)
	name := "assets/css/fonts.css"
	if b.opts.Fingerprint {
		sum := sha256.Sum256(css)
		name = "assets/css/fonts." + hex.EncodeToString(sum[:])[:8] + ".css"
	}
	if err := b.output.Write(name, css); err != nil {
		return err
	}
	b.css = append(b.css, "/"+name)
	return nil
}

// writeIslandManifest genera assets/islands/manifest.json (nombre → js/css).
func (b *Build) writeIslandManifest() error {
	type entry struct {
		JS  string `json:"js,omitempty"`
		CSS string `json:"css,omitempty"`
	}
	m := map[string]entry{}
	_ = fs.WalkDir(b.opts.AssetsFS, "islands", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		base := strings.TrimSuffix(path.Base(p), path.Ext(p))
		url := "/assets/" + p
		e := m[base]
		switch path.Ext(p) {
		case ".js":
			e.JS = url
		case ".css":
			e.CSS = url
		}
		m[base] = e
		return nil
	})
	if len(m) == 0 {
		return nil
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return b.output.Write("assets/islands/manifest.json", data)
}

// fingerprint escribe cada asset con un hash en el nombre y devuelve las URLs nuevas.
func (b *Build) fingerprint(urls []string) ([]string, error) {
	out := make([]string, len(urls))
	for i, u := range urls {
		rel := strings.TrimPrefix(u, "/assets/")
		data, err := fs.ReadFile(b.opts.AssetsFS, rel)
		if err != nil {
			out[i] = u
			continue
		}
		data = b.minifyAsset(rel, data)
		sum := sha256.Sum256(data)
		hashed := strings.TrimSuffix(u, path.Ext(u)) + "." + hex.EncodeToString(sum[:])[:8] + path.Ext(u)
		if err := b.output.Write(strings.TrimPrefix(hashed, "/"), data); err != nil {
			return nil, err
		}
		out[i] = hashed
	}
	return out, nil
}

// copyAssets copia los assets del tema al directorio de salida.
func (b *Build) copyAssets() error {
	return fs.WalkDir(b.opts.AssetsFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(b.opts.AssetsFS, p)
		if err != nil {
			return err
		}
		return b.output.Write(path.Join("assets", p), b.minifyAsset(p, data))
	})
}

func (b *Build) minifyAsset(name string, data []byte) []byte {
	if !b.opts.Minify {
		return data
	}
	switch strings.ToLower(path.Ext(name)) {
	case ".css":
		return minify.CSS(data)
	case ".js":
		return minify.JS(data)
	}
	return data
}

// published filtra los posts visibles: excluye borradores y posts futuros
// (publicación programada), salvo que se incluyan borradores.
func (b *Build) published(posts []domain.BlogPost) []domain.BlogPost {
	if b.opts.IncludeDrafts {
		return posts
	}
	out := make([]domain.BlogPost, 0, len(posts))
	for _, p := range posts {
		if p.Draft {
			continue
		}
		if !p.Published.IsZero() && p.Published.After(b.now) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// scheduled cuenta los posts programados (fecha futura, no borradores).
func (b *Build) scheduled(posts []domain.BlogPost) int {
	n := 0
	for _, p := range posts {
		if !p.Draft && !p.Published.IsZero() && p.Published.After(b.now) {
			n++
		}
	}
	return n
}

// navFromLinks convierte enlaces de config (con hijos) en enlaces de vista.
func navFromLinks(links []domain.NavLink) []Link {
	out := make([]Link, len(links))
	for i, l := range links {
		out[i] = Link{Label: l.Label, URL: l.URL, Children: navFromLinks(l.Children)}
	}
	return out
}

func navFor(dict map[string]string, lang string) []Link {
	if lang == "" {
		lang = "en"
	}
	label := func(key string) string {
		if v, ok := dict[key]; ok && v != "" {
			return v
		}
		return i18n.T(lang, key)
	}
	return []Link{
		{Label: label("nav.home"), URL: "/"},
		{Label: label("nav.blog"), URL: "/blog/"},
		{Label: label("nav.docs"), URL: "/docs/"},
		{Label: label("nav.tags"), URL: "/tags/"},
		{Label: label("nav.categories"), URL: "/categories/"},
		{Label: label("nav.authors"), URL: "/authors/"},
		{Label: label("nav.series"), URL: "/series/"},
		{Label: label("nav.archive"), URL: "/archive/"},
	}
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// Summary resume el contenido cargado y generado.
type Summary struct {
	Posts     int
	Published int
	Docs      int
	Authors   int
	Series    int
	Pages     int
	Drafts    int
	Scheduled int
	Warnings  []string
	Errors    []string
}

// NewSummary calcula el resumen a partir del contenido.
func NewSummary(c domain.Content) Summary {
	s := Summary{
		Posts:   len(c.Posts),
		Docs:    len(c.Docs),
		Authors: len(c.Authors),
		Series:  len(c.Series),
		Pages:   len(c.Pages),
	}
	for _, p := range c.Posts {
		if p.Draft {
			s.Drafts++
		}
	}
	return s
}
