package domain

import (
	"fmt"
	"net/url"
	"strings"
)

// Config es la configuración raíz del sitio, cargada desde site.yaml.
type Config struct {
	Site        SiteConfig         `yaml:"site"`
	Build       BuildConfig        `yaml:"build"`
	Render      RenderConfig       `yaml:"render"`
	Collections []CollectionConfig `yaml:"collections"`
	Taxonomies  []TaxonomyConfig   `yaml:"taxonomies"`
	Giscus      GiscusConfig       `yaml:"giscus"`
	Nav         []NavLink          `yaml:"nav"`
	Social      []SocialLink       `yaml:"social"`
	Support     []SocialLink       `yaml:"support"`
	Analytics   AnalyticsConfig    `yaml:"analytics"`
	Fonts       []FontConfig       `yaml:"fonts"`
	Admin       AdminConfig        `yaml:"admin"`
}

// AdminConfig configura el panel de edición.
// local: admin nativo en godian serve; remote: admin estático en dist/admin.
type AdminConfig struct {
	Mode     string `yaml:"mode"`     // local | remote | off
	Provider string `yaml:"provider"` // gitea | github
	APIBase  string `yaml:"api_base"` // URL directa del API (modo direct)
	BaseURL  string `yaml:"base_url"` // URL del proveedor (para el proxy)
	Owner    string `yaml:"owner"`
	Repo     string `yaml:"repo"`
	Branch   string `yaml:"branch"`
	APIMode  string `yaml:"api_mode"` // proxy | direct
}

// NormalizeAdmin completa y valida la configuración del admin.
func NormalizeAdmin(a *AdminConfig) error {
	a.Mode = strings.ToLower(strings.TrimSpace(a.Mode))
	if a.Mode == "" {
		a.Mode = "local"
	}
	switch a.Mode {
	case "local", "off":
		return nil
	case "remote":
	default:
		return fmt.Errorf("admin.mode inválido: %q (local, remote u off)", a.Mode)
	}

	a.Provider = strings.ToLower(strings.TrimSpace(a.Provider))
	if a.Provider == "" {
		a.Provider = "gitea"
	}
	if a.Provider != "gitea" && a.Provider != "github" {
		return fmt.Errorf("admin.provider inválido: %q", a.Provider)
	}
	a.Owner = strings.TrimSpace(a.Owner)
	a.Repo = strings.TrimSpace(a.Repo)
	a.Branch = strings.TrimSpace(a.Branch)
	if a.Branch == "" {
		a.Branch = "main"
	}
	if a.Owner == "" || a.Repo == "" {
		return fmt.Errorf("admin remoto: faltan owner y repo")
	}
	if a.Provider == "gitea" && strings.TrimSpace(a.BaseURL) == "" {
		return fmt.Errorf("admin remoto con Gitea: falta base_url")
	}
	a.BaseURL = strings.TrimRight(strings.TrimSpace(a.BaseURL), "/")

	a.APIMode = strings.ToLower(strings.TrimSpace(a.APIMode))
	if a.APIMode == "" {
		if a.Provider == "gitea" {
			a.APIMode = "proxy"
		} else {
			a.APIMode = "direct"
		}
	}
	if a.APIMode != "proxy" && a.APIMode != "direct" {
		return fmt.Errorf("admin.api_mode inválido: %q (proxy o direct)", a.APIMode)
	}
	if a.APIMode == "direct" && strings.TrimSpace(a.APIBase) == "" {
		a.APIBase = defaultAdminAPI(a.Provider, a.BaseURL)
	}
	a.APIBase = strings.TrimRight(strings.TrimSpace(a.APIBase), "/")
	return nil
}

func defaultAdminAPI(provider, baseURL string) string {
	switch provider {
	case "github":
		return "https://api.github.com"
	case "gitea":
		return strings.TrimRight(baseURL, "/") + "/api/v1"
	default:
		return ""
	}
}

// FontConfig declara una fuente a descargar y autohospedar durante el build.
type FontConfig struct {
	Family  string   `yaml:"family"`  // nombre en Google Fonts (p. ej. "Inter")
	Weights []int    `yaml:"weights"` // pesos (por defecto [400])
	Styles  []string `yaml:"styles"`  // normal | italic (por defecto [normal])
	Subsets []string `yaml:"subsets"` // latin, latin-ext, cyrillic… (por defecto [latin])
	Display string   `yaml:"display"` // swap (por defecto) | block | fallback | optional
}

// RenderConfig agrupa los render hooks configurables: cómo ciertos elementos
// Markdown se convierten en HTML (encabezados, enlaces externos e imágenes).
type RenderConfig struct {
	Headings      HeadingHooks      `yaml:"headings"`
	ExternalLinks ExternalLinkHooks `yaml:"external_links"`
	Images        ImageHooks        `yaml:"images"`
}

// HeadingHooks configura el render de los encabezados.
type HeadingHooks struct {
	Anchors bool   `yaml:"anchors"`      // añade un ancla enlazable a cada encabezado
	Anchor  string `yaml:"anchor_label"` // símbolo/texto del ancla
}

// ExternalLinkHooks configura el render de los enlaces externos.
type ExternalLinkHooks struct {
	Target string `yaml:"target"` // p. ej. "_blank"; vacío = sin atributo target
	Rel    string `yaml:"rel"`    // p. ej. "noopener noreferrer"; vacío = sin rel
}

// ImageHooks configura el render de las imágenes.
type ImageHooks struct {
	Loading  string `yaml:"loading"`  // lazy | eager | "" (sin atributo)
	Decoding string `yaml:"decoding"` // async | sync | "" (sin atributo)
	Figure   bool   `yaml:"figure"`   // envuelve imágenes con alt en <figure>/<figcaption>
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
	Color       string `yaml:"color"`     // color de acento (hex) para el tema por defecto
	Font        string `yaml:"font"`      // familia tipográfica del cuerpo (p. ej. "Inter")
	BasePath    string `yaml:"base_path"` // subruta de despliegue; si se omite, se deriva de URL
}

// EffectiveBasePath devuelve la subruta normalizada: explícita o derivada de URL.
func (s SiteConfig) EffectiveBasePath() string {
	if strings.TrimSpace(s.BasePath) != "" {
		return NormalizeBasePath(s.BasePath)
	}
	return BasePathFromURL(s.URL)
}

// BasePathFromURL extrae y normaliza el path de una URL.
func BasePathFromURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return NormalizeBasePath(u.Path)
}

// NormalizeBasePath deja la subruta con "/" inicial y sin "/" final.
func NormalizeBasePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimRight(p, "/")
}

// NormalizeSite resuelve BasePath y alinea Site.URL con esa subruta.
// Así canonical, feeds, sitemap y OG usan la URL pública correcta.
func NormalizeSite(s *SiteConfig) {
	s.BasePath = s.EffectiveBasePath()
	if s.BasePath == "" || strings.TrimSpace(s.URL) == "" {
		return
	}
	u, err := url.Parse(s.URL)
	if err != nil {
		return
	}
	u.Path = s.BasePath
	s.URL = strings.TrimRight(u.String(), "/")
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
		Render: DefaultRenderConfig(),
		Admin:  AdminConfig{Mode: "local"},
	}
}

// DefaultRenderConfig devuelve los render hooks por defecto, equivalentes al
// comportamiento histórico del renderer (anclas, target/rel en externos,
// imágenes lazy en <figure>).
func DefaultRenderConfig() RenderConfig {
	return RenderConfig{
		Headings:      HeadingHooks{Anchors: true, Anchor: "#"},
		ExternalLinks: ExternalLinkHooks{Target: "_blank", Rel: "noopener noreferrer"},
		Images:        ImageHooks{Loading: "lazy", Decoding: "async", Figure: true},
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
