package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/infrastructure/content"
	"github.com/nefdevelop/godian/internal/infrastructure/frontmatter"
)

// CollectionView es una colección con sus archivos para el dashboard.
type CollectionView struct {
	Name      string
	Label     string
	Folder    string
	Extension string
	Format    string
	Files     []FileView
	NewURL    string
}

// FileView es un archivo editable en el dashboard.
type FileView struct {
	Path    string
	Title   string
	Date    string
	EditURL string
	URL     string
}

// yamlDate emite fechas como timestamp sin comillas, igual que el contenido
// escrito a mano y que espera godian al leer el frontmatter.
type yamlDate string

// MarshalYAML convierte la fecha en un escalar con tag de timestamp.
func (d yamlDate) MarshalYAML() (any, error) {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!timestamp", Value: string(d)}, nil
}

func listCollections(cfg domain.Config, schema content.Schema) ([]CollectionView, error) {
	out := make([]CollectionView, 0, len(schema.Collections))
	for _, sc := range schema.Collections {
		view := collectionView(sc)
		view.NewURL = "/admin/new?collection=" + url.QueryEscape(sc.Name)
		root := filepath.FromSlash(sc.Folder)

		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				// Una colección puede no existir todavía; no es un error del admin.
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if strings.TrimPrefix(strings.ToLower(filepath.Ext(p)), ".") != sc.Extension {
				return nil
			}
			fm, body, readErr := readDocument(p, sc.Format)
			if readErr != nil {
				fm, body = map[string]any{}, ""
			}
			rel, relErr := filepath.Rel(".", p)
			if relErr != nil {
				rel = p
			}
			rel = filepath.ToSlash(rel)
			base := filepath.Base(p)
			view.Files = append(view.Files, FileView{
				Path:    rel,
				Title:   titleFrom(fm, body, base),
				Date:    dateFrom(fm),
				EditURL: "/admin/edit?collection=" + url.QueryEscape(sc.Name) + "&path=" + url.QueryEscape(rel),
				URL:     publicURL(sc.Name, slugFromPath(sc.Folder, p)),
			})
			return nil
		})

		sort.Slice(view.Files, func(i, j int) bool {
			return strings.ToLower(view.Files[i].Title) < strings.ToLower(view.Files[j].Title)
		})
		out = append(out, view)
	}
	return out, nil
}

func readDocument(full, format string) (map[string]any, string, error) {
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, "", err
	}
	fm := map[string]any{}
	switch format {
	case "frontmatter":
		head, body, err := frontmatter.Split(data)
		if err != nil {
			return nil, "", err
		}
		if len(head) > 0 {
			if err := yaml.Unmarshal(head, &fm); err != nil {
				return nil, "", fmt.Errorf("frontmatter inválido: %w", err)
			}
		}
		return fm, string(body), nil
	case "yaml":
		if len(strings.TrimSpace(string(data))) > 0 {
			if err := yaml.Unmarshal(data, &fm); err != nil {
				return nil, "", fmt.Errorf("YAML inválido: %w", err)
			}
		}
		return fm, "", nil
	case "json":
		if len(strings.TrimSpace(string(data))) > 0 {
			if err := json.Unmarshal(data, &fm); err != nil {
				return nil, "", fmt.Errorf("JSON inválido: %w", err)
			}
		}
		return fm, "", nil
	default:
		return nil, "", fmt.Errorf("formato desconocido: %s", format)
	}
}

func writeDocument(full, format string, fm map[string]any, body string) error {
	data, err := renderDocument(format, fm, body)
	if err != nil {
		return err
	}
	dir := filepath.Dir(full)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".godian-admin-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := os.Rename(name, full); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}

func renderDocument(format string, fm map[string]any, body string) ([]byte, error) {
	switch format {
	case "frontmatter":
		head, err := yaml.Marshal(fm)
		if err != nil {
			return nil, err
		}
		var b strings.Builder
		b.WriteString("---\n")
		b.Write(head)
		b.WriteString("---\n\n")
		b.WriteString(body)
		return []byte(b.String()), nil
	case "yaml":
		return yaml.Marshal(fm)
	case "json":
		data, err := json.MarshalIndent(fm, "", "  ")
		if err != nil {
			return nil, err
		}
		return append(data, '\n'), nil
	default:
		return nil, fmt.Errorf("formato desconocido: %s", format)
	}
}

func defaultValues(col content.SchemaCollection) map[string]any {
	fm := map[string]any{}
	today := time.Now().Format("2006-01-02")
	for _, f := range col.Fields {
		if f.Default != nil {
			fm[f.Name] = f.Default
		}
		if f.Type == "datetime" && (f.Name == "published" || f.Name == "pubDate") {
			if _, ok := fm[f.Name]; !ok {
				fm[f.Name] = today
			}
		}
	}
	if col.Name == "blog" {
		fm["draft"] = true
	}
	return fm
}

func documentTitle(fm map[string]any) string {
	if title := stringValue(fm["title"]); title != "" {
		return title
	}
	return stringValue(fm["name"])
}

func titleFrom(fm map[string]any, body, fallback string) string {
	if title := stringValue(fm["title"]); title != "" {
		return title
	}
	if name := stringValue(fm["name"]); name != "" {
		return name
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return strings.TrimSuffix(fallback, filepath.Ext(fallback))
}

func dateFrom(fm map[string]any) string {
	for _, key := range []string{"published", "pubDate", "date", "updatedDate", "updated"} {
		if v, ok := fm[key]; ok {
			if s := formatDateValue(v); s != "" {
				return s
			}
		}
	}
	return ""
}

func formatDateValue(v any) string {
	switch value := v.(type) {
	case time.Time:
		return value.Format("2006-01-02")
	case *time.Time:
		if value != nil {
			return value.Format("2006-01-02")
		}
	case string:
		value = strings.TrimSpace(value)
		if len(value) >= 10 {
			return value[:10]
		}
		return value
	case yamlDate:
		return string(value)
	}
	return ""
}

func stringValue(v any) string {
	switch value := v.(type) {
	case nil:
		return ""
	case string:
		return value
	case yamlDate:
		return string(value)
	case time.Time:
		return value.Format("2006-01-02")
	case []string:
		return strings.Join(value, ", ")
	case []any:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			parts = append(parts, stringValue(item))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(value)
	}
}

func resolveExisting(col content.SchemaCollection, pathParam string) (full, rel string, err error) {
	pathParam = strings.TrimSpace(pathParam)
	if pathParam == "" {
		return "", "", errors.New("falta la ruta del archivo")
	}
	clean := filepath.Clean(filepath.FromSlash(pathParam))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." ||
		strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("ruta no permitida")
	}
	full, err = filepath.Abs(clean)
	if err != nil {
		return "", "", err
	}
	root, err := filepath.Abs(filepath.FromSlash(col.Folder))
	if err != nil {
		return "", "", err
	}
	if !within(root, full) {
		return "", "", errors.New("el archivo no pertenece a la colección")
	}
	if strings.TrimPrefix(strings.ToLower(filepath.Ext(full)), ".") != col.Extension {
		return "", "", fmt.Errorf("se esperaba un archivo .%s", col.Extension)
	}
	if info, statErr := os.Stat(full); statErr != nil || info.IsDir() {
		return "", "", errors.New("el archivo no existe")
	}
	rel, err = filepath.Rel(".", full)
	if err != nil {
		rel = full
	}
	return full, filepath.ToSlash(rel), nil
}

func resolveTarget(col content.SchemaCollection, pathParam, filename, title string, isNew bool) (full, rel string, err error) {
	if !isNew {
		return resolveExisting(col, pathParam)
	}

	name := strings.TrimSpace(filename)
	if name == "" {
		name = title
	}
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = slugify(name)
	if name == "" {
		name = "entrada"
	}
	name += "." + col.Extension

	root, err := filepath.Abs(filepath.FromSlash(col.Folder))
	if err != nil {
		return "", "", err
	}
	full = filepath.Join(root, name)
	if !within(root, full) {
		return "", "", errors.New("ruta no permitida")
	}
	if _, statErr := os.Stat(full); statErr == nil {
		return "", "", fmt.Errorf("el archivo %q ya existe", name)
	}
	rel, err = filepath.Rel(".", full)
	if err != nil {
		rel = full
	}
	return full, filepath.ToSlash(rel), nil
}

func within(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

var accentReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "ä", "a", "â", "a", "ã", "a",
	"é", "e", "è", "e", "ë", "e", "ê", "e",
	"í", "i", "ì", "i", "ï", "i", "î", "i",
	"ó", "o", "ò", "o", "ö", "o", "ô", "o", "õ", "o",
	"ú", "u", "ù", "u", "ü", "u", "û", "u",
	"ñ", "n", "ç", "c",
)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = accentReplacer.Replace(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func slugFromPath(folder, full string) string {
	root := filepath.FromSlash(folder)
	rel, err := filepath.Rel(root, full)
	if err != nil {
		rel = filepath.Base(full)
	}
	rel = filepath.ToSlash(rel)
	return strings.TrimSuffix(rel, filepath.Ext(rel))
}

func publicURL(collection, slug string) string {
	slug = strings.TrimSuffix(slug, "/index")
	switch collection {
	case "blog":
		return "/blog/" + escapeSlug(slug) + "/"
	case "docs":
		return "/docs/" + escapeSlug(slug) + "/"
	case "pages":
		if slug == "" || slug == "index" {
			return "/"
		}
		return "/" + escapeSlug(slug) + "/"
	case "series":
		return "/series/" + escapeSlug(slug) + "/"
	default:
		return ""
	}
}

func escapeSlug(slug string) string {
	parts := strings.Split(slug, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func saveUpload(cfg domain.Config, file multipart.File, header *multipart.FileHeader) (string, error) {
	const maxSize = 10 << 20
	data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxSize {
		return "", errors.New("la imagen supera 10 MB")
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		".gif": true, ".avif": true, ".svg": true,
	}
	if !allowed[ext] {
		return "", errors.New("formato de imagen no permitido")
	}
	ct := http.DetectContentType(data)
	if !strings.HasPrefix(ct, "image/") && !(ext == ".svg" && strings.Contains(ct, "xml")) {
		return "", errors.New("el archivo no parece una imagen")
	}

	base := slugify(strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename)))
	if base == "" {
		base = "imagen"
	}
	name := fmt.Sprintf("%s-%d%s", base, time.Now().UnixNano(), ext)
	dir := filepath.Join(cfg.Build.AssetsDir, "uploads")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

// yamlNames lee los YAML de una carpeta y devuelve un campo de texto (nombre o
// título), ordenado y sin duplicados. Se usa para sugerencias y selects.
func yamlNames(cfg domain.Config, folder, field string) []string {
	root := filepath.Join(cfg.Build.ContentDir, folder)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		fm, _, err := readDocument(filepath.Join(root, entry.Name()), "yaml")
		if err != nil {
			continue
		}
		value := stringValue(fm[field])
		if value == "" {
			value = stringValue(fm["name"])
		}
		if value == "" {
			value = stringValue(fm["title"])
		}
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
