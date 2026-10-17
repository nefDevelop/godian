package application

import (
	"html/template"
	"strconv"

	"github.com/nefdevelop/godian/internal/domain"
)

// Link es un enlace de navegación (con hijos opcionales).
type Link struct {
	Label    string
	URL      string
	Children []Link
}

// PostCard es la vista resumida de un post en listados.
type PostCard struct {
	Title       string
	URL         string
	Description string
	Date        string
	Tags        []string
	ReadingTime string
}

// Term es un término de taxonomía (tag o categoría) con su conteo.
type Term struct {
	Name  string
	URL   string
	Count int
}

// Pagination describe la paginación de un listado.
type Pagination struct {
	Current int
	Last    int
	BaseURL string
	PrevURL string
	NextURL string
}

// HasPrev indica si existe página anterior.
func (p Pagination) HasPrev() bool { return p.Current > 1 }

// HasNext indica si existe página siguiente.
func (p Pagination) HasNext() bool { return p.Current < p.Last }

// Pages devuelve los números de página (1..Last).
func (p Pagination) Pages() []int {
	out := make([]int, 0, p.Last)
	for i := 1; i <= p.Last; i++ {
		out = append(out, i)
	}
	return out
}

// DocLink es un enlace a una página de documentación.
type DocLink struct {
	Title  string
	URL    string
	Active bool
}

// DocSection agrupa enlaces de documentación por sección.
type DocSection struct {
	Section string
	Docs    []DocLink
}

// SiteStats resume conteos para widgets.
type SiteStats struct {
	Posts      int
	Tags       int
	Categories int
}

// PostNav es una referencia a un post vecino.
type PostNav struct {
	Title string
	URL   string
}

// SeriesNav describe la posición de un post dentro de su serie.
type SeriesNav struct {
	Title string
	URL   string
	Index int
	Total int
	Prev  *PostNav
	Next  *PostNav
}

// TOCItem es una entrada de la tabla de contenidos.
type TOCItem struct {
	Level int
	ID    string
	Text  string
	URL   string
}

// ArchiveYear agrupa los posts de un año.
type ArchiveYear struct {
	Year  int
	Posts []ArchivePost
}

// ArchivePost es la vista de un post en el archivo.
type ArchivePost struct {
	Title    string
	URL      string
	Date     string
	Category string
}

// AuthorCard es la vista resumida de un autor.
type AuthorCard struct {
	Name   string
	Slug   string
	URL    string
	Avatar string
	Bio    string
	Count  int
}

// AuthorProfile es la vista completa de un autor.
type AuthorProfile struct {
	Name    string
	Slug    string
	Avatar  string
	Bio     string
	Website string
	Github  string
	Twitter string
	Count   int
}

// ChapterCard es la vista de un capítulo dentro de una serie.
type ChapterCard struct {
	Number      int
	Title       string
	URL         string
	Description string
	Date        string
	ReadingTime string
	Minutes     int
}

// SeriesCard es la vista resumida de una serie.
type SeriesCard struct {
	Title       string
	Slug        string
	URL         string
	Cover       string
	Description string
	Chapters    int
	Minutes     int
}

// SeriesDetail es la vista completa de una serie.
type SeriesDetail struct {
	Title       string
	Slug        string
	Cover       string
	Description string
	Authors     []string
	Chapters    []ChapterCard
	Minutes     int
}

// PageData es el modelo de datos que consumen las plantillas.
type PageData struct {
	Site        domain.SiteConfig
	PageURL     string
	Canonical   string
	NavActive   string
	Title       string
	Description string
	Body        template.HTML
	Date        string
	Tags        []string
	Categories  []string
	ReadingTime string
	TOC         []TOCItem
	Prev        *PostNav
	Next        *PostNav
	Related     []PostCard
	SeriesNav   *SeriesNav
	DocsNav     []DocSection
	Breadcrumb  []Link
	Featured    *PostCard
	HomeTags    []Term
	HomeCats    []Term
	Stats       *SiteStats
	NeedsMath   bool
	OGImage     string
	JSONLD      template.JS
	SiteJSONLD  template.JS
	Giscus      domain.GiscusConfig
	Social      []domain.SocialLink
	Support     []domain.SocialLink
	Analytics   domain.AnalyticsConfig
	Data        map[string]any
	Posts       []PostCard
	Nav         []Link
	CSS         []string
	JS          []string
	Year        int

	// F2: taxonomías, paginación y listados.
	Terms      []Term
	Page       Pagination
	Total      int
	Years      []ArchiveYear
	Authors    []AuthorCard
	Author     *AuthorProfile
	SeriesList []SeriesCard
	Series     *SeriesDetail
}

// pageURL construye la URL de una página de listado.
func pageURL(base string, page int) string {
	if page <= 1 {
		return base
	}
	return base + strconv.Itoa(page) + "/"
}

// newPagination construye la paginación de un listado.
func newPagination(base string, current, last int) Pagination {
	p := Pagination{Current: current, Last: last, BaseURL: base}
	if current > 1 {
		p.PrevURL = pageURL(base, current-1)
	}
	if current < last {
		p.NextURL = pageURL(base, current+1)
	}
	return p
}
