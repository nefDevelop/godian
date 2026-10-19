// Package admin implementa el panel de edición nativo de godian. No usa npm:
// las plantillas, el CSS y el JS van embebidos y el contenido se escribe
// directamente en content/, reutilizando el schema y el tema activo.
package admin

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/i18n"
	"github.com/nefdevelop/godian/internal/infrastructure/config"
	"github.com/nefdevelop/godian/internal/infrastructure/content"
	"github.com/nefdevelop/godian/internal/infrastructure/icons"
	"github.com/nefdevelop/godian/internal/infrastructure/markdown"
	"github.com/nefdevelop/godian/internal/infrastructure/theme"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed assets/*
var assetFS embed.FS

// PageData es el contrato que consumen las plantillas del admin.
type PageData struct {
	Title         string
	Site          domain.SiteConfig
	Lang          string
	CSRF          string
	ThemeName     string
	ThemeCSS      []string
	ThemeAdminCSS string
	FontsCSS      string
	ContentDir    string
	UploadsDir    string
	Collections   []CollectionView
	Collection    CollectionView
	Path          string
	Filename      string
	IsNew         bool
	IsMarkdown    bool
	Fields        []FieldView
	Body          string
	Saved         bool
	Error         string
	PublicURL     string
}

// Handler sirve el admin y escribe en el contenido local.
type Handler struct {
	configPath string
	csrf       string
	dashboard  *template.Template
	edit       *template.Template
}

// New carga las plantillas y genera el token CSRF de la sesión.
func New(configPath string) (*Handler, error) {
	funcs := template.FuncMap{
		"icon": func(string, ...string) template.HTML { return "" },
		"join": strings.Join,
		"json": prettyJSON,
		"date": formatDateValue,
	}

	dashboard, err := template.New("layout").Funcs(funcs).ParseFS(
		templateFS, "templates/layout.html", "templates/dashboard.html")
	if err != nil {
		return nil, fmt.Errorf("plantillas del dashboard: %w", err)
	}
	edit, err := template.New("layout").Funcs(funcs).ParseFS(
		templateFS, "templates/layout.html", "templates/edit.html", "templates/field.html")
	if err != nil {
		return nil, fmt.Errorf("plantillas del editor: %w", err)
	}

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, fmt.Errorf("generando token CSRF: %w", err)
	}
	return &Handler{
		configPath: configPath,
		csrf:       hex.EncodeToString(token),
		dashboard:  dashboard,
		edit:       edit,
	}, nil
}

// ServeHTTP enruta las peticiones del admin.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Los uploads son assets públicos del sitio y deben funcionar sin loopback.
	if strings.HasPrefix(r.URL.Path, "/uploads/") {
		h.handleUploads(w, r)
		return
	}
	if !isLoopback(r) {
		http.Error(w, "admin: solo accesible desde localhost", http.StatusForbidden)
		return
	}
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	switch {
	case r.URL.Path == "/admin":
		http.Redirect(w, r, "/admin/", http.StatusMovedPermanently)
	case r.URL.Path == "/admin/" || r.URL.Path == "/admin/index.html":
		h.handleDashboard(w, r)
	case r.URL.Path == "/admin/edit":
		h.handleEdit(w, r, false)
	case r.URL.Path == "/admin/new":
		h.handleEdit(w, r, true)
	case r.URL.Path == "/admin/save":
		h.handleSave(w, r)
	case r.URL.Path == "/admin/preview":
		h.handlePreview(w, r)
	case r.URL.Path == "/admin/upload":
		h.handleUpload(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/theme/"):
		h.handleTheme(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/assets/"):
		h.handleAssets(w, r)
	default:
		http.NotFound(w, r)
	}
}

// load recarga configuración y tema en cada petición: así los cambios en
// site.yaml o en el tema se reflejan sin reiniciar el servidor de desarrollo.
func (h *Handler) load() (domain.Config, *theme.Theme, content.Schema, error) {
	cfg, err := config.Load(h.configPath)
	if err != nil {
		return domain.Config{}, nil, content.Schema{}, err
	}
	th, err := theme.Load(cfg.Build.ThemeDir)
	if err != nil {
		return domain.Config{}, nil, content.Schema{}, fmt.Errorf("cargando tema: %w", err)
	}
	return cfg, th, content.BuildSchema(cfg), nil
}

// basePage prepara los datos comunes de cualquier pantalla del admin.
func (h *Handler) basePage(cfg domain.Config, th *theme.Theme, schema content.Schema) PageData {
	lang := cfg.Site.Language
	if lang == "" {
		lang = "en"
	}
	assets := loadThemeAssets(cfg, th)
	return PageData{
		Site:          cfg.Site,
		Lang:          lang,
		CSRF:          h.csrf,
		ThemeName:     assets.Name,
		ThemeCSS:      assets.CSS,
		ThemeAdminCSS: assets.AdminCSS,
		FontsCSS:      assets.FontsCSS,
		ContentDir:    schema.ContentDir,
		UploadsDir:    schema.UploadsDir,
	}
}

// render ejecuta la plantilla con los iconos del tema activo.
func (h *Handler) render(w http.ResponseWriter, tmpl *template.Template, data PageData, th *theme.Theme) {
	funcs := template.FuncMap{
		"join": strings.Join,
		"json": prettyJSON,
		"date": formatDateValue,
	}

	iconReg, err := icons.New()
	if err == nil {
		_ = iconReg.LoadDir(th.FS(), "icons")
		funcs["icon"] = func(name string, label ...string) template.HTML {
			text := ""
			if len(label) > 0 {
				text = label[0]
			}
			out, err := iconReg.HTML(name, text)
			if err != nil {
				return ""
			}
			return out
		}
	} else {
		funcs["icon"] = func(string, ...string) template.HTML { return "" }
	}

	clone, err := tmpl.Clone()
	if err != nil {
		http.Error(w, "admin: plantilla inválida", http.StatusInternalServerError)
		return
	}
	clone = clone.Funcs(funcs)

	var buf bytes.Buffer
	if err := clone.ExecuteTemplate(&buf, "layout", data); err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(buf.Bytes())
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	cfg, th, schema, err := h.load()
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	collections, err := listCollections(cfg, schema)
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := h.basePage(cfg, th, schema)
	data.Title = "Contenido"
	data.Collections = collections
	h.render(w, h.dashboard, data, th)
}

func (h *Handler) handleEdit(w http.ResponseWriter, r *http.Request, forceNew bool) {
	cfg, th, schema, err := h.load()
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	name := r.URL.Query().Get("collection")
	col, ok := findCollection(schema, name)
	if !ok {
		http.Error(w, "admin: colección desconocida", http.StatusBadRequest)
		return
	}

	data := h.basePage(cfg, th, schema)
	data.Collection = collectionView(col)
	data.IsMarkdown = col.Format == "frontmatter"
	data.Saved = r.URL.Query().Get("saved") == "1"

	pathParam := strings.TrimSpace(r.URL.Query().Get("path"))
	data.IsNew = forceNew || pathParam == ""
	if data.IsNew {
		fm := defaultValues(col)
		data.Filename = strings.TrimSpace(r.URL.Query().Get("filename"))
		data.Title = "Nuevo " + col.Label
		data.Body = ""
		data.Fields = buildFields(col, fm, cfg, schema)
	} else {
		full, rel, err := resolveExisting(col, pathParam)
		if err != nil {
			http.Error(w, "admin: "+err.Error(), http.StatusBadRequest)
			return
		}
		fm, body, err := readDocument(full, col.Format)
		if err != nil {
			http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
			return
		}
		base := filepath.Base(rel)
		data.Path = rel
		data.Filename = base
		data.Body = body
		data.Title = titleFrom(fm, body, base)
		data.PublicURL = publicURL(col.Name, slugFromPath(col.Folder, full))
		data.Fields = buildFields(col, fm, cfg, schema)
	}
	h.render(w, h.edit, data, th)
}

func (h *Handler) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "formulario inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !h.checkCSRF(r) {
		http.Error(w, "token CSRF inválido", http.StatusForbidden)
		return
	}

	cfg, th, schema, err := h.load()
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	col, ok := findCollection(schema, r.FormValue("collection"))
	if !ok {
		http.Error(w, "admin: colección desconocida", http.StatusBadRequest)
		return
	}

	fm, parseErr := parseForm(col, r, cfg)
	body := r.FormValue("body")
	pathParam := strings.TrimSpace(r.FormValue("path"))
	filename := strings.TrimSpace(r.FormValue("filename"))
	isNew := pathParam == ""

	if parseErr != nil {
		h.renderEditError(w, cfg, th, schema, col, fm, body, pathParam, filename, isNew, parseErr.Error())
		return
	}

	target, rel, err := resolveTarget(col, pathParam, filename, documentTitle(fm), isNew)
	if err != nil {
		h.renderEditError(w, cfg, th, schema, col, fm, body, pathParam, filename, isNew, err.Error())
		return
	}
	if err := writeDocument(target, col.Format, fm, body); err != nil {
		h.renderEditError(w, cfg, th, schema, col, fm, body, pathParam, filename, isNew, err.Error())
		return
	}

	edit := "/admin/edit?collection=" + url.QueryEscape(col.Name) +
		"&path=" + url.QueryEscape(filepath.ToSlash(rel)) + "&saved=1"
	http.Redirect(w, r, edit, http.StatusSeeOther)
}

func (h *Handler) renderEditError(w http.ResponseWriter, cfg domain.Config, th *theme.Theme, schema content.Schema, col content.SchemaCollection, fm map[string]any, body, pathParam, filename string, isNew bool, message string) {
	data := h.basePage(cfg, th, schema)
	data.Collection = collectionView(col)
	data.IsMarkdown = col.Format == "frontmatter"
	data.IsNew = isNew
	data.Path = pathParam
	data.Filename = filename
	data.Body = body
	data.Error = message
	data.Fields = buildFields(col, fm, cfg, schema)
	if isNew {
		data.Title = "Nuevo " + col.Label
	} else {
		data.Title = titleFrom(fm, body, filename)
		if pathParam != "" {
			if full, _, err := resolveExisting(col, pathParam); err == nil {
				data.PublicURL = publicURL(col.Name, slugFromPath(col.Folder, full))
			}
		}
	}
	h.render(w, h.edit, data, th)
}

func (h *Handler) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "formulario inválido", http.StatusBadRequest)
		return
	}
	if !h.checkCSRF(r) {
		http.Error(w, "token CSRF inválido", http.StatusForbidden)
		return
	}

	cfg, err := config.Load(h.configPath)
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	lang := cfg.Site.Language
	if lang == "" {
		lang = "en"
	}
	dict, _ := i18n.Load(filepath.FromSlash(cfg.Build.I18nDir), lang, nil)
	tr := func(key string) string {
		if v, ok := dict[key]; ok && v != "" {
			return v
		}
		return i18n.T(lang, key)
	}
	md := markdown.New(
		markdown.WithTranslator(tr),
		markdown.WithRenderConfig(cfg.Render),
	)
	out, _, err := md.Render(r.FormValue("content"))
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(out))
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "subida inválida: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !h.checkCSRF(r) {
		http.Error(w, "token CSRF inválido", http.StatusForbidden)
		return
	}
	cfg, err := config.Load(h.configPath)
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "falta el archivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	public, err := saveUpload(cfg, file, header)
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": public})
}

func (h *Handler) handleTheme(w http.ResponseWriter, r *http.Request) {
	_, th, _, err := h.load()
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	assets, err := th.AssetsFS()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, "/admin/theme/")
	if rel == "" || !fs.ValidPath(rel) {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(assets, rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(strings.ToLower(rel), ".css") {
		data = rewriteCSSAssets(data)
	}
	if ct := contentType(rel); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

func (h *Handler) handleAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := fs.Sub(assetFS, "assets")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.StripPrefix("/admin/assets/", http.FileServer(http.FS(assets))).ServeHTTP(w, r)
}

func (h *Handler) handleUploads(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(h.configPath)
	if err != nil {
		http.Error(w, "admin: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, "/uploads/")
	rel = path.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, "../") || !fs.ValidPath(rel) {
		http.NotFound(w, r)
		return
	}

	// El build copia uploads a dist/ y genera variantes webp/avif. Se sirve
	// primero dist/ para no romper el sitio; si aún no se ha reconstruido, se
	// cae al archivo original en assets/uploads (p. ej. una subida reciente).
	candidates := []string{
		filepath.Join(cfg.Build.OutputDir, "uploads", filepath.FromSlash(rel)),
		filepath.Join(cfg.Build.AssetsDir, "uploads", filepath.FromSlash(rel)),
	}
	for _, file := range candidates {
		if info, statErr := os.Stat(file); statErr == nil && !info.IsDir() {
			w.Header().Set("Cache-Control", "no-store")
			if ct := contentType(file); ct != "" {
				w.Header().Set("Content-Type", ct)
			}
			http.ServeFile(w, r, file)
			return
		}
	}
	http.NotFound(w, r)
}

func (h *Handler) checkCSRF(r *http.Request) bool {
	token := r.Header.Get("X-Admin-CSRF")
	if token == "" {
		token = r.FormValue("csrf")
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(h.csrf)) == 1
}

func findCollection(schema content.Schema, name string) (content.SchemaCollection, bool) {
	for _, c := range schema.Collections {
		if c.Name == name {
			return c, true
		}
	}
	return content.SchemaCollection{}, false
}

func collectionView(c content.SchemaCollection) CollectionView {
	return CollectionView{
		Name:      c.Name,
		Label:     c.Label,
		Folder:    c.Folder,
		Extension: c.Extension,
		Format:    c.Format,
	}
}

func isLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

// rewriteCSSAssets adapta las URLs absolutas del tema (/assets/...) al prefijo
// bajo el que el admin sirve sus assets.
func rewriteCSSAssets(data []byte) []byte {
	replacements := [][2]string{
		{`url(/assets/`, `url(/admin/theme/`},
		{`url('/assets/`, `url('/admin/theme/`},
		{`url("/assets/`, `url("/admin/theme/`},
	}
	for _, r := range replacements {
		data = bytes.ReplaceAll(data, []byte(r[0]), []byte(r[1]))
	}
	return data
}

func contentType(name string) string {
	if ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); ct != "" {
		return ct
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".json":
		return "application/json"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".avif":
		return "image/avif"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	}
	return ""
}
