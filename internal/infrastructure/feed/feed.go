// Package feed genera RSS y sitemap sin dependencias externas.
package feed

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"time"

	"github.com/nefdevelop/godian/internal/domain"
)

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

// Sitemap construye un sitemap a partir de las rutas indicadas.
func (g *Generator) Sitemap(site domain.SiteConfig, paths []string) ([]byte, error) {
	doc := sitemapDoc{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, p := range paths {
		doc.URLs = append(doc.URLs, sitemapURL{Loc: site.URL + p})
	}
	return marshal(doc)
}

func marshal(v any) ([]byte, error) {
	body, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}
