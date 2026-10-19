package application

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/nefdevelop/godian/internal/domain"
)

type memOutput struct {
	files map[string][]byte
}

func (m *memOutput) Write(rel string, data []byte) error {
	if m.files == nil {
		m.files = map[string][]byte{}
	}
	m.files[rel] = data
	return nil
}

func (m *memOutput) Clean() error { m.files = map[string][]byte{}; return nil }

func TestWriteAdminLocalNoop(t *testing.T) {
	out := &memOutput{}
	b := &Build{output: out, opts: Options{Admin: domain.AdminConfig{Mode: "local"}}}
	if err := b.writeAdmin(domain.Content{}); err != nil {
		t.Fatal(err)
	}
	if len(out.files) != 0 {
		t.Errorf("modo local escribió %d archivos", len(out.files))
	}
}

func TestWriteAdminRemote(t *testing.T) {
	out := &memOutput{}
	b := &Build{
		output: out,
		opts: Options{
			Site: domain.SiteConfig{
				Title:    "Sitio",
				Language: "es",
				BasePath: "/base",
			},
			Admin: domain.AdminConfig{
				Mode:     "remote",
				Provider: "gitea",
				BaseURL:  "https://git.example.com",
				Owner:    "pi",
				Repo:     "godian",
				Branch:   "main",
				APIMode:  "proxy",
			},
			AdminSchema: []byte("{\"version\":1}\n"),
			Taxonomies: []domain.TaxonomyConfig{{
				Name:  "tags",
				Terms: []domain.TermConfig{{Name: "go"}, {Name: "web"}},
			}},
			AssetsFS: fstest.MapFS{
				"css/admin.css":    &fstest.MapFile{Data: []byte("/* admin */")},
				"css/styles.css":   &fstest.MapFile{Data: []byte("/* base */")},
				"css/chunks/x.css": &fstest.MapFile{Data: []byte("/* chunk */")},
			},
		},
		css:       []string{"/assets/css/styles.css"},
		chunkURLs: map[string]string{"x": "/assets/css/chunks/x.css"},
	}
	content := domain.Content{
		Authors: []domain.Author{{Name: "Nef"}},
		Series:  []domain.Series{{Title: "Go"}},
	}
	if err := b.writeAdmin(content); err != nil {
		t.Fatal(err)
	}

	index := string(out.files["admin/index.html"])
	for _, want := range []string{
		"/base/admin/assets/app.js",
		"/base/admin/assets/admin.css",
		"/base/assets/css/admin.css",
		"/base/assets/css/styles.css",
		"/base/assets/css/chunks/x.css",
		"noindex,nofollow",
	} {
		if !strings.Contains(index, want) {
			t.Errorf("falta %q en admin/index.html", want)
		}
	}

	config := string(out.files["admin/config.json"])
	for _, want := range []string{
		`"api_mode": "proxy"`,
		`"api_base": "/admin/api"`,
		`"uploads_url": "/base/uploads"`,
		`"owner": "pi"`,
	} {
		if !strings.Contains(config, want) {
			t.Errorf("falta %q en config.json: %s", want, config)
		}
	}
	if string(out.files["admin/schema.json"]) != "{\"version\":1}\n" {
		t.Errorf("schema.json = %q", out.files["admin/schema.json"])
	}
	options := string(out.files["admin/options.json"])
	for _, want := range []string{`"go"`, `"web"`, `"Nef"`, `"Go"`} {
		if !strings.Contains(options, want) {
			t.Errorf("falta %q en options.json: %s", want, options)
		}
	}
	for _, file := range []string{"admin/assets/app.js", "admin/vendor/js-yaml.min.js"} {
		if len(out.files[file]) == 0 {
			t.Errorf("no se copió %s", file)
		}
	}
	if !strings.Contains(string(out.files["_headers"]), "X-Frame-Options: DENY") {
		t.Errorf("_headers = %q", out.files["_headers"])
	}
}
