package admin

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/infrastructure/theme"
)

type themeAssets struct {
	Name     string
	CSS      []string
	AdminCSS string
	FontsCSS string
}

// loadThemeAssets resuelve el CSS del tema que consume el admin. Se cargan los
// estilos globales, los chunks (para que el preview se parezca al artículo) y
// el admin.css opcional del tema. El JS público del tema no se carga.
func loadThemeAssets(cfg domain.Config, th *theme.Theme) themeAssets {
	out := themeAssets{Name: th.Manifest.Name}
	if out.Name == "" {
		out.Name = "default"
	}
	assets, err := th.AssetsFS()
	if err != nil {
		return out
	}

	seen := map[string]bool{}
	add := func(rel string) {
		rel = cleanThemePath(rel)
		if rel == "" || rel == "css/admin.css" || seen[rel] {
			return
		}
		if _, err := fs.Stat(assets, rel); err != nil {
			return
		}
		seen[rel] = true
		out.CSS = append(out.CSS, "/admin/theme/"+rel)
	}

	for _, public := range th.Manifest.Assets.CSS {
		add(themeRel(public))
	}

	if entries, err := fs.ReadDir(assets, "css/chunks"); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".css") {
				continue
			}
			add("css/chunks/" + entry.Name())
		}
	}

	if _, err := fs.Stat(assets, "css/admin.css"); err == nil {
		out.AdminCSS = "/admin/theme/css/admin.css"
	}

	// El build fingerprinta el CSS de fuentes (fonts.<hash>.css). Se enlaza la
	// variante real si el sitio ya se ha construido.
	matches, _ := filepath.Glob(filepath.Join(cfg.Build.OutputDir, "assets", "css", "fonts*.css"))
	for _, match := range matches {
		if filepath.Base(match) == "fonts.css" {
			out.FontsCSS = "/assets/css/fonts.css"
			return out
		}
	}
	if len(matches) > 0 {
		out.FontsCSS = "/assets/css/" + filepath.Base(matches[0])
	}
	return out
}

func themeRel(public string) string {
	public = strings.TrimSpace(public)
	if public == "" || strings.Contains(public, "://") {
		return ""
	}
	switch {
	case strings.HasPrefix(public, "/assets/"):
		return strings.TrimPrefix(public, "/assets/")
	case strings.HasPrefix(public, "assets/"):
		return strings.TrimPrefix(public, "assets/")
	case strings.HasPrefix(public, "/"):
		return strings.TrimPrefix(public, "/")
	default:
		return public
	}
}

func cleanThemePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	p = path.Clean(p)
	if p == "." || p == ".." || strings.HasPrefix(p, "../") || !fs.ValidPath(p) {
		return ""
	}
	return p
}
