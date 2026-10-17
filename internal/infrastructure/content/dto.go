package content

import (
	"time"

	"github.com/nefdevelop/godian/internal/domain"
)

// postFrontmatter refleja el esquema de la colección "blog".
type postFrontmatter struct {
	Title       string        `yaml:"title"`
	Published   time.Time     `yaml:"published"`
	Updated     *time.Time    `yaml:"updated"`
	Description string        `yaml:"description"`
	Image       string        `yaml:"image"`
	Tags        []string      `yaml:"tags"`
	Categories  []string      `yaml:"categories"`
	Author      []string      `yaml:"author"`
	Difficulty  string        `yaml:"difficulty"`
	Draft       bool          `yaml:"draft"`
	Pinned      bool          `yaml:"pinned"`
	Series      string        `yaml:"series"`
	Order       *int          `yaml:"order"`
	Links       *linksDTO     `yaml:"links"`
	Downloads   []downloadDTO `yaml:"downloads"`
	Aliases     []string      `yaml:"aliases"`
}

type linksDTO struct {
	Website string `yaml:"website"`
	Github  string `yaml:"github"`
}

type downloadDTO struct {
	ID    string `yaml:"id"`
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}

func (p postFrontmatter) toDomain(slug, body, path string) domain.BlogPost {
	post := domain.BlogPost{
		Title:       p.Title,
		Published:   p.Published,
		Updated:     p.Updated,
		Description: p.Description,
		Image:       p.Image,
		Tags:        orEmpty(p.Tags),
		Categories:  orEmpty(p.Categories),
		Authors:     orEmpty(p.Author),
		Difficulty:  domain.Difficulty(p.Difficulty),
		Draft:       p.Draft,
		Pinned:      p.Pinned,
		Series:      p.Series,
		Order:       p.Order,
		Aliases:     orEmpty(p.Aliases),
		Slug:        slug,
		Body:        body,
		FilePath:    path,
	}
	if p.Links != nil {
		post.Links = &domain.PostLinks{Website: p.Links.Website, Github: p.Links.Github}
	}
	for _, d := range p.Downloads {
		post.Downloads = append(post.Downloads, domain.Download{ID: d.ID, Label: d.Label, URL: d.URL})
	}
	return post
}

// docFrontmatter refleja el esquema de la colección "docs".
type docFrontmatter struct {
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
	PubDate     *time.Time `yaml:"pubDate"`
	UpdatedDate *time.Time `yaml:"updatedDate"`
	Order       int        `yaml:"order"`
	Section     string     `yaml:"section"`
	Tags        []string   `yaml:"tags"`
	HeroImage   string     `yaml:"heroImage"`
	Aliases     []string   `yaml:"aliases"`
}

func (d docFrontmatter) toDomain(slug, body, path string) domain.Doc {
	return domain.Doc{
		Title:       d.Title,
		Description: d.Description,
		PubDate:     d.PubDate,
		UpdatedDate: d.UpdatedDate,
		Order:       d.Order,
		Section:     d.Section,
		Tags:        orEmpty(d.Tags),
		HeroImage:   d.HeroImage,
		Aliases:     orEmpty(d.Aliases),
		Slug:        slug,
		Body:        body,
		FilePath:    path,
	}
}

// authorFile refleja el esquema de la colección "authors".
type authorFile struct {
	Name    string `yaml:"name"`
	Avatar  string `yaml:"avatar"`
	Bio     string `yaml:"bio"`
	Website string `yaml:"website"`
	Github  string `yaml:"github"`
	Twitter string `yaml:"twitter"`
}

func (a authorFile) toDomain(slug string) domain.Author {
	return domain.Author{
		Slug:    slug,
		Name:    a.Name,
		Avatar:  a.Avatar,
		Bio:     a.Bio,
		Website: a.Website,
		Github:  a.Github,
		Twitter: a.Twitter,
	}
}

// seriesFile refleja el esquema de la colección "series".
type seriesFile struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Cover       string   `yaml:"cover"`
	Author      []string `yaml:"author"`
	Chapters    []string `yaml:"chapters"`
}

func (s seriesFile) toDomain(slug string) domain.Series {
	return domain.Series{
		Slug:        slug,
		Title:       s.Title,
		Description: s.Description,
		Cover:       s.Cover,
		Authors:     orEmpty(s.Author),
		Chapters:    orEmpty(s.Chapters),
	}
}

// pageFrontmatter refleja el esquema de la colección "pages".
type pageFrontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Stack       []string `yaml:"stack"`
}

func (p pageFrontmatter) toDomain(slug, body, path string) domain.Page {
	return domain.Page{
		Slug:        slug,
		Title:       p.Title,
		Description: p.Description,
		Stack:       orEmpty(p.Stack),
		Body:        body,
		FilePath:    path,
	}
}

// orEmpty garantiza slices no nulos para un manejo uniforme.
func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
