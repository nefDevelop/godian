// Package theme carga el tema del sitio: manifiesto, plantillas y assets.
// El tema por defecto va embebido en el binario; un theme_dir en disco lo
// sobrescribe archivo a archivo (herencia con fallback al default).
package theme

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/nefdevelop/godian/themes"
)

// Assets describe los recursos que declara el tema.
type Assets struct {
	CSS []string `yaml:"css"`
	JS  []string `yaml:"js"`
}

// Manifest es el contenido de theme.yaml.
type Manifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Assets  Assets `yaml:"assets"`
}

// Theme es un tema cargado y listo para usar.
type Theme struct {
	fsys     fs.FS
	Manifest Manifest
}

// Load carga el tema embebido y, si themeDir es un directorio existente,
// superpone sus archivos sobre el tema por defecto.
func Load(themeDir string) (*Theme, error) {
	base := themes.Default()

	var tfsys fs.FS = base
	if themeDir != "" {
		if info, statErr := os.Stat(themeDir); statErr == nil && info.IsDir() {
			tfsys = overlay{primary: os.DirFS(themeDir), fallback: base}
		}
	}

	m, err := loadManifest(tfsys)
	if err != nil {
		return nil, err
	}
	return &Theme{fsys: tfsys, Manifest: m}, nil
}

// Scaffold copia el tema por defecto embebido a dir para poder personalizarlo
// (overlay: los archivos presentes en dir sobrescriben el tema embebido).
func Scaffold(dir string) error {
	base := themes.Default()
	return fs.WalkDir(base, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dir, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(base, p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// FS devuelve la raíz del tema (templates + assets + theme.yaml).
func (t *Theme) FS() fs.FS {
	return t.fsys
}

// TemplateFS devuelve el subárbol "templates" del tema.
func (t *Theme) TemplateFS() (fs.FS, error) {
	return fs.Sub(t.fsys, "templates")
}

// AssetsFS devuelve el subárbol "assets" del tema, si existe.
func (t *Theme) AssetsFS() (fs.FS, error) {
	return fs.Sub(t.fsys, "assets")
}

func loadManifest(fsys fs.FS) (Manifest, error) {
	var m Manifest
	data, err := fs.ReadFile(fsys, "theme.yaml")
	if err != nil {
		return m, fmt.Errorf("leyendo theme.yaml: %w", err)
	}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("parseando theme.yaml: %w", err)
	}
	return m, nil
}

// overlay es un fs.FS que resuelve primero en primary (disco) y cae al
// fallback (embebido) cuando el archivo no existe.
type overlay struct {
	primary  fs.FS
	fallback fs.FS
}

func (o overlay) Open(name string) (fs.File, error) {
	if f, err := o.primary.Open(name); err == nil {
		return f, nil
	}
	return o.fallback.Open(name)
}

func (o overlay) ReadDir(name string) ([]fs.DirEntry, error) {
	merged := map[string]fs.DirEntry{}
	// Un directorio puede existir solo en el tema en disco (p. ej. icons/) o solo
	// en el embebido: se tolera que falte en cualquiera de los dos.
	entries, err := fs.ReadDir(o.fallback, name)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	for _, e := range entries {
		merged[e.Name()] = e
	}
	primaryEntries, pErr := fs.ReadDir(o.primary, name)
	if pErr != nil && !errors.Is(pErr, fs.ErrNotExist) {
		return nil, pErr
	}
	for _, e := range primaryEntries {
		merged[e.Name()] = e
	}
	if len(merged) == 0 {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	names := make([]string, 0, len(merged))
	for n := range merged {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]fs.DirEntry, 0, len(names))
	for _, n := range names {
		out = append(out, merged[n])
	}
	return out, nil
}
