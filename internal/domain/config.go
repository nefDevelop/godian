package domain

// Config es la configuración raíz del sitio, cargada desde site.yaml.
type Config struct {
	Site        SiteConfig         `yaml:"site"`
	Build       BuildConfig        `yaml:"build"`
	Collections []CollectionConfig `yaml:"collections"`
	Taxonomies  []TaxonomyConfig   `yaml:"taxonomies"`
	Giscus      GiscusConfig       `yaml:"giscus"`
	Nav         []NavLink          `yaml:"nav"`
	Social      []SocialLink       `yaml:"social"`
	Support     []SocialLink       `yaml:"support"`
	Analytics   AnalyticsConfig    `yaml:"analytics"`
}

// NavLink es un enlace de navegación configurable (con hijos opcionales).
type NavLink struct {
	Label    string    `yaml:"label"`
	URL      string    `yaml:"url"`
	Children []NavLink `yaml:"children"`
}

// SocialLink es un enlace social o de apoyo configurable.
type SocialLink struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
	Icon  string `yaml:"icon"`
}

// AnalyticsConfig configura un proveedor de analítica (opt-in).
type AnalyticsConfig struct {
	Provider string `yaml:"provider"` // plausible | ga4
	ID       string `yaml:"id"`
}

// GiscusConfig configura los comentarios vía GitHub Discussions.
type GiscusConfig struct {
	Repo       string `yaml:"repo"`
	RepoID     string `yaml:"repo_id"`
	Category   string `yaml:"category"`
	CategoryID string `yaml:"category_id"`
}

// SiteConfig contiene los metadatos globales del sitio.
type SiteConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
	Language    string `yaml:"language"`
	Color       string `yaml:"color"` // color de acento (hex) para el tema por defecto
}

// BuildConfig define las rutas y opciones del build.
type BuildConfig struct {
	ContentDir  string `yaml:"content_dir"`
	ThemeDir    string `yaml:"theme_dir"`
	OutputDir   string `yaml:"output_dir"`
	AssetsDir   string `yaml:"assets_dir"`
	DataDir     string `yaml:"data_dir"`
	I18nDir     string `yaml:"i18n_dir"`
	Minify      bool   `yaml:"minify"`
	Fingerprint bool   `yaml:"fingerprint"`
	Optimize    bool   `yaml:"optimize_images"`
	Clean       bool   `yaml:"clean"`

	ImagePresets map[string]ImagePreset `yaml:"image_presets"`
}

// ImagePreset define una variante de imagen con recorte/escala.
type ImagePreset struct {
	Width   int    `yaml:"width"`
	Height  int    `yaml:"height"`
	Fit     string `yaml:"fit"`   // fit | fill
	Focal   string `yaml:"focal"` // center | top | bottom | left | right
	Quality int    `yaml:"quality"`
	Blur    int    `yaml:"blur"`
}

// CollectionConfig describe una colección de contenido dentro de ContentDir.
type CollectionConfig struct {
	Name   string `yaml:"name"`
	Folder string `yaml:"folder"`
}

// TaxonomyConfig describe una taxonomía (tags, categorías, etc.).
type TaxonomyConfig struct {
	Name         string       `yaml:"name"`         // identificador interno
	Label        string       `yaml:"label"`        // título visible
	URL          string       `yaml:"url"`          // base de rutas, p. ej. /tags
	Field        string       `yaml:"field"`        // campo del frontmatter del post
	Prefix       string       `yaml:"prefix"`       // prefijo del título de término (p. ej. "#")
	Hierarchical bool         `yaml:"hierarchical"` // expande "a/b" en "a" y "a/b"
	RenderEmpty  bool         `yaml:"render_empty"` // genera páginas sin posts
	Strict       bool         `yaml:"strict"`       // avisa de términos no declarados
	Terms        []TermConfig `yaml:"terms"`
}

// TermConfig es la declaración de un término de taxonomía.
type TermConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Aliases     []string `yaml:"aliases"`
	Weight      int      `yaml:"weight"`
}

// DefaultConfig devuelve una configuración con valores por defecto.
func DefaultConfig() Config {
	return Config{
		Site: SiteConfig{
			Language: "en",
		},
		Build: BuildConfig{
			ContentDir:  "content",
			ThemeDir:    "themes/default",
			OutputDir:   "dist",
			AssetsDir:   "assets",
			DataDir:     "data",
			I18nDir:     "i18n",
			Minify:      true,
			Fingerprint: true,
			Optimize:    true,
			Clean:       true,
		},
	}
}

// DefaultCollections devuelve las colecciones estándar del tema.
func DefaultCollections() []CollectionConfig {
	return []CollectionConfig{
		{Name: "blog", Folder: "blog"},
		{Name: "docs", Folder: "docs"},
		{Name: "authors", Folder: "authors"},
		{Name: "series", Folder: "series"},
		{Name: "pages", Folder: "pages"},
	}
}

// DefaultTaxonomies devuelve las taxonomías estándar (tags y categorías).
func DefaultTaxonomies() []TaxonomyConfig {
	return []TaxonomyConfig{
		{Name: "tags", Label: "Tags", URL: "/tags", Field: "tags", Prefix: "#", Hierarchical: true},
		{Name: "categories", Label: "Categories", URL: "/categories", Field: "categories"},
	}
}
