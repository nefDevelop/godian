package domain

import "time"

// Heading es un encabezado extraído de un documento Markdown (para el TOC).
type Heading struct {
	Level int
	ID    string
	Text  string
}

// Difficulty representa el nivel de dificultad de un post.
type Difficulty string

const (
	DifficultyBeginner     Difficulty = "beginner"
	DifficultyIntermediate Difficulty = "intermediate"
	DifficultyAdvanced     Difficulty = "advanced"
)

// PostLinks agrupa los enlaces externos declarados en el frontmatter de un post.
type PostLinks struct {
	Website string
	Github  string
}

// Download describe un recurso descargable asociado a un post.
type Download struct {
	ID    string
	Label string
	URL   string
}

// BlogPost es un artículo del blog.
type BlogPost struct {
	Title       string
	Published   time.Time
	Updated     *time.Time
	Description string
	Image       string
	Tags        []string
	Categories  []string
	Authors     []string
	Difficulty  Difficulty
	Draft       bool
	Pinned      bool
	Series      string
	Order       *int
	Links       *PostLinks
	Downloads   []Download
	Aliases     []string

	// Campos derivados durante la carga.
	Slug     string
	Body     string
	FilePath string
}

// Doc es una página de documentación.
type Doc struct {
	Title       string
	Description string
	PubDate     *time.Time
	UpdatedDate *time.Time
	Order       int
	Section     string
	Tags        []string
	HeroImage   string
	Aliases     []string

	// Campos derivados durante la carga.
	Slug     string
	Body     string
	FilePath string
}

// Author es el perfil de un autor, definido en un archivo YAML/JSON.
type Author struct {
	Slug    string
	Name    string
	Avatar  string
	Bio     string
	Website string
	Github  string
	Twitter string
}

// Series es una serie de posts, definida en un archivo YAML/JSON.
type Series struct {
	Slug        string
	Title       string
	Description string
	Cover       string
	Authors     []string
	Chapters    []string
}

// Page es una página especial (por ejemplo "about").
type Page struct {
	Slug        string
	Title       string
	Description string
	Stack       []string
	Body        string
	FilePath    string
}

// Content agrupa todo el contenido cargado del sitio.
type Content struct {
	Posts   []BlogPost
	Docs    []Doc
	Authors []Author
	Series  []Series
	Pages   []Page
}
