// Package content descubre y parsea las colecciones de contenido desde el
// sistema de archivos, respetando el esquema de frontmatter de astro-ri.
package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/infrastructure/frontmatter"
)

// Loader carga colecciones de contenido.
type Loader struct {
	contentDir  string
	collections map[string]string
}

// NewLoader construye un Loader a partir de la configuración del sitio.
func NewLoader(cfg domain.Config) *Loader {
	cols := make(map[string]string, len(cfg.Collections))
	for _, c := range cfg.Collections {
		cols[c.Name] = c.Folder
	}
	return &Loader{contentDir: cfg.Build.ContentDir, collections: cols}
}

// Load carga todas las colecciones conocidas.
func (l *Loader) Load() (domain.Content, error) {
	var c domain.Content
	var err error
	if c.Posts, err = l.LoadPosts(); err != nil {
		return c, err
	}
	if c.Docs, err = l.LoadDocs(); err != nil {
		return c, err
	}
	if c.Authors, err = l.LoadAuthors(); err != nil {
		return c, err
	}
	if c.Series, err = l.LoadSeries(); err != nil {
		return c, err
	}
	if c.Pages, err = l.LoadPages(); err != nil {
		return c, err
	}
	return c, nil
}

// LoadPosts carga la colección "blog".
func (l *Loader) LoadPosts() ([]domain.BlogPost, error) {
	files, err := listFiles(l.folder("blog", "blog"), ".md", ".mdx")
	if err != nil {
		return nil, err
	}
	posts := make([]domain.BlogPost, 0, len(files))
	seen := map[string]string{}
	for _, f := range files {
		slug := slugOf(f.rel)
		if err := checkDuplicate(seen, slug, f.rel); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("leyendo %s: %w", f.path, err)
		}
		fm, body, err := frontmatter.Split(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.path, err)
		}
		var dto postFrontmatter
		if err := frontmatter.Unmarshal(fm, &dto); err != nil {
			return nil, fmt.Errorf("%s: %w", f.path, err)
		}
		posts = append(posts, dto.toDomain(slug, string(body), f.path))
	}
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Published.After(posts[j].Published)
	})
	return posts, nil
}

// LoadDocs carga la colección "docs".
func (l *Loader) LoadDocs() ([]domain.Doc, error) {
	files, err := listFiles(l.folder("docs", "docs"), ".md", ".mdx")
	if err != nil {
		return nil, err
	}
	docs := make([]domain.Doc, 0, len(files))
	seen := map[string]string{}
	for _, f := range files {
		slug := slugOf(f.rel)
		if err := checkDuplicate(seen, slug, f.rel); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("leyendo %s: %w", f.path, err)
		}
		fm, body, err := frontmatter.Split(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.path, err)
		}
		var dto docFrontmatter
		if err := frontmatter.Unmarshal(fm, &dto); err != nil {
			return nil, fmt.Errorf("%s: %w", f.path, err)
		}
		docs = append(docs, dto.toDomain(slug, string(body), f.path))
	}
	return docs, nil
}

// LoadAuthors carga la colección "authors" (archivos YAML).
func (l *Loader) LoadAuthors() ([]domain.Author, error) {
	files, err := listFiles(l.folder("authors", "authors"), ".yaml", ".yml")
	if err != nil {
		return nil, err
	}
	authors := make([]domain.Author, 0, len(files))
	seen := map[string]string{}
	for _, f := range files {
		slug := slugOf(f.rel)
		if err := checkDuplicate(seen, slug, f.rel); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("leyendo %s: %w", f.path, err)
		}
		var dto authorFile
		if err := frontmatter.Unmarshal(data, &dto); err != nil {
			return nil, fmt.Errorf("%s: %w", f.path, err)
		}
		authors = append(authors, dto.toDomain(slug))
	}
	return authors, nil
}

// LoadSeries carga la colección "series" (archivos YAML).
func (l *Loader) LoadSeries() ([]domain.Series, error) {
	files, err := listFiles(l.folder("series", "series"), ".yaml", ".yml")
	if err != nil {
		return nil, err
	}
	series := make([]domain.Series, 0, len(files))
	seen := map[string]string{}
	for _, f := range files {
		slug := slugOf(f.rel)
		if err := checkDuplicate(seen, slug, f.rel); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("leyendo %s: %w", f.path, err)
		}
		var dto seriesFile
		if err := frontmatter.Unmarshal(data, &dto); err != nil {
			return nil, fmt.Errorf("%s: %w", f.path, err)
		}
		series = append(series, dto.toDomain(slug))
	}
	return series, nil
}

// LoadPages carga la colección "pages" (Markdown con frontmatter o YAML).
func (l *Loader) LoadPages() ([]domain.Page, error) {
	dir := l.folder("pages", "pages")
	files, err := listFiles(dir, ".md", ".mdx", ".yaml", ".yml")
	if err != nil {
		return nil, err
	}
	pages := make([]domain.Page, 0, len(files))
	seen := map[string]string{}
	for _, f := range files {
		slug := slugOf(f.rel)
		if err := checkDuplicate(seen, slug, f.rel); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(f.path)
		if err != nil {
			return nil, fmt.Errorf("leyendo %s: %w", f.path, err)
		}
		var dto pageFrontmatter
		var body []byte
		switch strings.ToLower(filepath.Ext(f.path)) {
		case ".yaml", ".yml":
			if err := frontmatter.Unmarshal(data, &dto); err != nil {
				return nil, fmt.Errorf("%s: %w", f.path, err)
			}
		default:
			fm, b, err := frontmatter.Split(data)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.path, err)
			}
			if err := frontmatter.Unmarshal(fm, &dto); err != nil {
				return nil, fmt.Errorf("%s: %w", f.path, err)
			}
			body = b
		}
		pages = append(pages, dto.toDomain(slug, string(body), f.path))
	}
	return pages, nil
}

// folder resuelve la carpeta de una colección, con fallback.
func (l *Loader) folder(name, fallback string) string {
	if f, ok := l.collections[name]; ok && f != "" {
		return filepath.Join(l.contentDir, f)
	}
	return filepath.Join(l.contentDir, fallback)
}

// listFiles devuelve (dir absoluto, ruta relativa) de los archivos bajo dir
// con alguna de las extensiones dadas, recursivamente e ignorando ocultos.
// Un directorio inexistente no es un error.
func listFiles(dir string, exts ...string) ([]fileRef, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	var out []fileRef
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		for _, want := range exts {
			if ext == want {
				rel, rerr := filepath.Rel(dir, path)
				if rerr != nil {
					return rerr
				}
				out = append(out, fileRef{path: path, rel: rel})
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out, nil
}

// fileRef es la ubicación de un archivo de contenido.
type fileRef struct {
	path string // ruta completa para leer
	rel  string // ruta relativa a la carpeta de la colección
}

// slugOf deriva el slug de la ruta relativa (sin extensión, con "/").
func slugOf(rel string) string {
	rel = filepath.ToSlash(rel)
	return strings.TrimSuffix(rel, filepath.Ext(rel))
}

// checkDuplicate detecta slugs repetidos dentro de una colección.
func checkDuplicate(seen map[string]string, slug, rel string) error {
	if prev, ok := seen[slug]; ok {
		return fmt.Errorf("slug duplicado %q: %s y %s", slug, prev, rel)
	}
	seen[slug] = rel
	return nil
}
