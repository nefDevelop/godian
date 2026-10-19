package content

import (
	"path"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
)

// Schema describe las colecciones y taxonomías para consumidores externos
// (por ejemplo, el CMS). Es la representación pública del contrato de contenido.
type Schema struct {
	Version     int                `json:"version"`
	ContentDir  string             `json:"content_dir"`
	AssetsDir   string             `json:"assets_dir"`
	UploadsDir  string             `json:"uploads_dir"`
	Collections []SchemaCollection `json:"collections"`
	Taxonomies  []SchemaTaxonomy   `json:"taxonomies"`
}

// SchemaCollection describe una colección y sus campos editables.
type SchemaCollection struct {
	Name      string        `json:"name"`
	Label     string        `json:"label"`
	Folder    string        `json:"folder"`
	Extension string        `json:"extension"`
	Format    string        `json:"format"`
	Fields    []SchemaField `json:"fields"`
}

// SchemaField describe un campo del frontmatter o de un YAML de datos.
type SchemaField struct {
	Name     string        `json:"name"`
	Label    string        `json:"label"`
	Type     string        `json:"type"`
	Required bool          `json:"required,omitempty"`
	Default  any           `json:"default,omitempty"`
	Options  []string      `json:"options,omitempty"`
	Source   string        `json:"source,omitempty"`
	Item     string        `json:"item,omitempty"`
	Fields   []SchemaField `json:"fields,omitempty"`
}

// SchemaTaxonomy describe una taxonomía configurable.
type SchemaTaxonomy struct {
	Name         string       `json:"name"`
	Label        string       `json:"label"`
	Field        string       `json:"field"`
	URL          string       `json:"url"`
	Hierarchical bool         `json:"hierarchical,omitempty"`
	Strict       bool         `json:"strict,omitempty"`
	Terms        []SchemaTerm `json:"terms"`
}

// SchemaTerm es un término declarado de una taxonomía.
type SchemaTerm struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
}

// BuildSchema construye el contrato de contenido a partir de la configuración.
func BuildSchema(cfg domain.Config) Schema {
	contentDir := cfg.Build.ContentDir
	if contentDir == "" {
		contentDir = "content"
	}
	assetsDir := cfg.Build.AssetsDir
	if assetsDir == "" {
		assetsDir = "assets"
	}
	collections := cfg.Collections
	if len(collections) == 0 {
		collections = domain.DefaultCollections()
	}
	taxonomies := cfg.Taxonomies
	if len(taxonomies) == 0 {
		taxonomies = domain.DefaultTaxonomies()
	}

	s := Schema{
		Version:     1,
		ContentDir:  contentDir,
		AssetsDir:   assetsDir,
		UploadsDir:  path.Join(assetsDir, "uploads"),
		Collections: make([]SchemaCollection, 0, len(collections)),
		Taxonomies:  make([]SchemaTaxonomy, 0, len(taxonomies)),
	}
	for _, c := range collections {
		s.Collections = append(s.Collections, collectionSchema(contentDir, c))
	}
	for _, t := range taxonomies {
		s.Taxonomies = append(s.Taxonomies, taxonomySchema(t))
	}
	return s
}

func collectionSchema(contentDir string, c domain.CollectionConfig) SchemaCollection {
	folder := path.Join(contentDir, c.Folder)
	ext, format := "md", "frontmatter"
	switch c.Name {
	case "authors", "series":
		ext, format = "yaml", "yaml"
	}
	return SchemaCollection{
		Name:      c.Name,
		Label:     collectionLabel(c.Name),
		Folder:    folder,
		Extension: ext,
		Format:    format,
		Fields:    collectionFields(c.Name),
	}
}

func collectionLabel(name string) string {
	switch name {
	case "blog":
		return "Blog"
	case "docs":
		return "Documentación"
	case "pages":
		return "Páginas"
	case "authors":
		return "Autores"
	case "series":
		return "Series"
	default:
		return strings.ToUpper(name[:1]) + name[1:]
	}
}

func collectionFields(name string) []SchemaField {
	switch name {
	case "blog":
		return []SchemaField{
			{Name: "title", Label: "Título", Type: "string", Required: true},
			{Name: "published", Label: "Fecha de publicación", Type: "datetime", Required: true},
			{Name: "updated", Label: "Última actualización", Type: "datetime"},
			{Name: "description", Label: "Descripción", Type: "text", Required: true},
			{Name: "image", Label: "Imagen de portada", Type: "image"},
			{Name: "tags", Label: "Etiquetas", Type: "list", Item: "string", Source: "taxonomy:tags"},
			{Name: "categories", Label: "Categorías", Type: "list", Item: "string", Source: "taxonomy:categories"},
			{Name: "author", Label: "Autor(es)", Type: "list", Item: "string", Source: "authors"},
			{Name: "difficulty", Label: "Dificultad", Type: "select", Options: []string{"beginner", "intermediate", "advanced"}},
			{Name: "draft", Label: "Borrador", Type: "boolean", Default: false},
			{Name: "pinned", Label: "Fijado", Type: "boolean", Default: false},
			{Name: "series", Label: "Serie", Type: "select", Source: "series"},
			{Name: "order", Label: "Orden en la serie", Type: "number"},
			{Name: "links", Label: "Enlaces", Type: "object", Fields: []SchemaField{
				{Name: "website", Label: "Sitio web", Type: "url"},
				{Name: "github", Label: "GitHub", Type: "url"},
			}},
			{Name: "downloads", Label: "Descargas", Type: "list", Fields: []SchemaField{
				{Name: "id", Label: "ID", Type: "string"},
				{Name: "label", Label: "Etiqueta", Type: "string"},
				{Name: "url", Label: "URL", Type: "url"},
			}},
			{Name: "aliases", Label: "Aliases", Type: "list", Item: "string"},
		}
	case "docs":
		return []SchemaField{
			{Name: "title", Label: "Título", Type: "string", Required: true},
			{Name: "description", Label: "Descripción", Type: "text", Required: true},
			{Name: "pubDate", Label: "Fecha de publicación", Type: "datetime"},
			{Name: "updatedDate", Label: "Última actualización", Type: "datetime"},
			{Name: "order", Label: "Orden", Type: "number", Default: 0},
			{Name: "section", Label: "Sección", Type: "string"},
			{Name: "tags", Label: "Etiquetas", Type: "list", Item: "string", Source: "taxonomy:tags"},
			{Name: "heroImage", Label: "Imagen principal", Type: "image"},
			{Name: "aliases", Label: "Aliases", Type: "list", Item: "string"},
		}
	case "pages":
		return []SchemaField{
			{Name: "title", Label: "Título", Type: "string"},
			{Name: "description", Label: "Descripción", Type: "text"},
			{Name: "stack", Label: "Stack", Type: "list", Item: "string"},
		}
	case "authors":
		return []SchemaField{
			{Name: "name", Label: "Nombre", Type: "string", Required: true},
			{Name: "avatar", Label: "Avatar", Type: "image"},
			{Name: "bio", Label: "Biografía", Type: "text"},
			{Name: "website", Label: "Sitio web", Type: "url"},
			{Name: "github", Label: "GitHub", Type: "url"},
			{Name: "twitter", Label: "Twitter/X", Type: "url"},
		}
	case "series":
		return []SchemaField{
			{Name: "title", Label: "Título", Type: "string", Required: true},
			{Name: "description", Label: "Descripción", Type: "text"},
			{Name: "cover", Label: "Portada", Type: "image"},
			{Name: "author", Label: "Autor(es)", Type: "list", Item: "string", Source: "authors"},
			{Name: "chapters", Label: "Capítulos", Type: "list", Item: "string"},
		}
	default:
		return nil
	}
}

func taxonomySchema(t domain.TaxonomyConfig) SchemaTaxonomy {
	label := t.Label
	if label == "" {
		label = t.Name
	}
	terms := make([]SchemaTerm, 0, len(t.Terms))
	for _, term := range t.Terms {
		terms = append(terms, SchemaTerm{
			Name:        term.Name,
			Description: term.Description,
			Aliases:     term.Aliases,
		})
	}
	return SchemaTaxonomy{
		Name:         t.Name,
		Label:        label,
		Field:        t.Field,
		URL:          t.URL,
		Hierarchical: t.Hierarchical,
		Strict:       t.Strict,
		Terms:        terms,
	}
}
