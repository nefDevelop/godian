// Package feed genera RSS y sitemap sin dependencias externas.
package feed

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/nefdevelop/godian/internal/domain"
)

// MaxSitemapURLs es el máximo de URLs por sitemap antes de partirlo en varios
// ficheros y emitir un índice (límite recomendado por el protocolo sitemaps).
const MaxSitemapURLs = 50000

// Generator construye feeds del sitio.
type Generator struct{}

// New crea un Generator.
func New() *Generator { return &Generator{} }

type rssDoc struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language,omitempty"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description,omitempty"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid"`
}

// RSS construye un feed RSS 2.0 con los posts publicados.
func (g *Generator) RSS(site domain.SiteConfig, posts []domain.BlogPost) ([]byte, error) {
	doc := rssDoc{
		Version: "2.0",
		Channel: rssChannel{
			Title:       site.Title,
			Link:        site.URL,
			Description: site.Description,
			Language:    site.Language,
		},
	}
	for _, p := range posts {
		link := site.URL + "/blog/" + p.Slug + "/"
		doc.Channel.Items = append(doc.Channel.Items, rssItem{
			Title:       p.Title,
			Link:        link,
			Description: p.Description,
			PubDate:     p.Published.Format(time.RFC1123Z),
			GUID:        link,
		})
	}
	return marshal(doc)
}

type atomDoc struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    []atomLink  `xml:"link"`
	Updated string      `xml:"updated"`
	ID      string      `xml:"id"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type atomEntry struct {
	Title     string   `xml:"title"`
	Link      atomLink `xml:"link"`
	ID        string   `xml:"id"`
	Updated   string   `xml:"updated"`
	Published string   `xml:"published,omitempty"`
	Summary   string   `xml:"summary,omitempty"`
}

// Atom construye un feed Atom 1.0 con los posts publicados.
func (g *Generator) Atom(site domain.SiteConfig, posts []domain.BlogPost) ([]byte, error) {
	base := strings.TrimRight(site.URL, "/")
	doc := atomDoc{
		Xmlns: "http://www.w3.org/2005/Atom",
		Title: site.Title,
		ID:    base + "/",
		Link: []atomLink{
			{Href: base + "/atom.xml", Rel: "self", Type: "application/atom+xml"},
			{Href: base + "/", Rel: "alternate", Type: "text/html"},
		},
	}
	latest := time.Now()
	for i, p := range posts {
		link := base + "/blog/" + p.Slug + "/"
		updated := p.Published
		if p.Updated != nil {
			updated = *p.Updated
		}
		if i == 0 {
			latest = updated
		}
		doc.Entries = append(doc.Entries, atomEntry{
			Title:     p.Title,
			Link:      atomLink{Href: link, Rel: "alternate", Type: "text/html"},
			ID:        link,
			Updated:   updated.Format(time.RFC3339),
			Published: p.Published.Format(time.RFC3339),
			Summary:   p.Description,
		})
	}
	doc.Updated = latest.Format(time.RFC3339)
	return marshal(doc)
}

type jsonFeed struct {
	Version     string         `json:"version"`
	Title       string         `json:"title"`
	HomePageURL string         `json:"home_page_url"`
	FeedURL     string         `json:"feed_url"`
	Description string         `json:"description,omitempty"`
	Items       []jsonFeedItem `json:"items"`
}

type jsonFeedItem struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary,omitempty"`
	DatePublished string   `json:"date_published,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// JSONFeed construye un JSON Feed 1.1 con los posts publicados.
func (g *Generator) JSONFeed(site domain.SiteConfig, posts []domain.BlogPost) ([]byte, error) {
	base := strings.TrimRight(site.URL, "/")
	doc := jsonFeed{
		Version:     "https://jsonfeed.org/version/1.1",
		Title:       site.Title,
		HomePageURL: base + "/",
		FeedURL:     base + "/feed.json",
		Description: site.Description,
		Items:       make([]jsonFeedItem, 0, len(posts)),
	}
	for _, p := range posts {
		link := base + "/blog/" + p.Slug + "/"
		doc.Items = append(doc.Items, jsonFeedItem{
			ID:            link,
			URL:           link,
			Title:         p.Title,
			Summary:       p.Description,
			DatePublished: p.Published.Format(time.RFC3339),
			Tags:          p.Tags,
		})
	}
	return json.MarshalIndent(doc, "", "  ")
}

type sitemapDoc struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemapIndexDoc struct {
	XMLName  xml.Name       `xml:"sitemapindex"`
	Xmlns    string         `xml:"xmlns,attr"`
	Sitemaps []sitemapEntry `xml:"sitemap"`
}

type sitemapEntry struct {
	Loc string `xml:"loc"`
}

// Sitemap construye un sitemap (urlset) a partir de las rutas indicadas.
func (g *Generator) Sitemap(site domain.SiteConfig, paths []string) ([]byte, error) {
	doc := sitemapDoc{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, p := range paths {
		doc.URLs = append(doc.URLs, sitemapURL{Loc: site.URL + p})
	}
	return marshal(doc)
}

// SitemapIndex construye un índice de sitemaps que apunta a las rutas hijas.
// Devuelve el índice; los hijos los genera Sitemaps.
func (g *Generator) SitemapIndex(site domain.SiteConfig, paths []string) ([]byte, error) {
	doc := sitemapIndexDoc{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, p := range paths {
		doc.Sitemaps = append(doc.Sitemaps, sitemapEntry{Loc: site.URL + p})
	}
	return marshal(doc)
}

// Sitemaps construye el sitemap del sitio. Si el número de rutas no supera
// MaxSitemapURLs, emite un único `sitemap.xml`. En caso contrario, lo parte en
// `sitemap-N.xml` y emite `sitemap.xml` como índice. La clave del mapa es la
// ruta relativa del fichero dentro de dist/.
func (g *Generator) Sitemaps(site domain.SiteConfig, paths []string) (map[string][]byte, error) {
	return g.sitemaps(site, paths, MaxSitemapURLs)
}

// sitemaps implementa el troceado con un límite parametrizable (testeable).
func (g *Generator) sitemaps(site domain.SiteConfig, paths []string, max int) (map[string][]byte, error) {
	files := make(map[string][]byte)
	if len(paths) <= max {
		data, err := g.Sitemap(site, paths)
		if err != nil {
			return nil, err
		}
		files["sitemap.xml"] = data
		return files, nil
	}

	var children []string
	for i := 0; i < len(paths); i += max {
		end := i + max
		if end > len(paths) {
			end = len(paths)
		}
		name := fmt.Sprintf("sitemap-%d.xml", i/max+1)
		data, err := g.Sitemap(site, paths[i:end])
		if err != nil {
			return nil, err
		}
		files[name] = data
		children = append(children, "/"+name)
	}

	index, err := g.SitemapIndex(site, children)
	if err != nil {
		return nil, err
	}
	files["sitemap.xml"] = index
	return files, nil
}

func marshal(v any) ([]byte, error) {
	body, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}
