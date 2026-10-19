package content

import (
	"reflect"
	"strings"
	"testing"

	"github.com/nefdevelop/godian/internal/domain"
)

func TestBuildSchema(t *testing.T) {
	cfg := domain.DefaultConfig()
	cfg.Taxonomies = domain.DefaultTaxonomies()
	cfg.Taxonomies[0].Terms = []domain.TermConfig{{Name: "go", Description: "Go", Aliases: []string{"golang"}}}
	s := BuildSchema(cfg)
	if s.Version != 1 || s.ContentDir != "content" || s.UploadsDir != "assets/uploads" {
		t.Errorf("schema base = %+v", s)
	}

	blog := schemaCollection(t, s, "blog")
	if blog.Folder != "content/blog" || blog.Extension != "md" || blog.Format != "frontmatter" {
		t.Errorf("blog = %+v", blog)
	}
	for _, name := range []string{"title", "published", "description", "image", "tags", "categories", "author", "difficulty", "draft", "pinned", "series", "order", "links", "downloads", "aliases"} {
		if !hasSchemaField(blog.Fields, name) {
			t.Errorf("blog sin campo %q", name)
		}
	}

	docs := schemaCollection(t, s, "docs")
	for _, name := range []string{"title", "description", "pubDate", "updatedDate", "order", "section", "tags", "heroImage", "aliases"} {
		if !hasSchemaField(docs.Fields, name) {
			t.Errorf("docs sin campo %q", name)
		}
	}

	authors := schemaCollection(t, s, "authors")
	if authors.Folder != "content/authors" || authors.Extension != "yaml" || authors.Format != "yaml" {
		t.Errorf("authors = %+v", authors)
	}

	var tags *SchemaTaxonomy
	for i := range s.Taxonomies {
		if s.Taxonomies[i].Name == "tags" {
			tags = &s.Taxonomies[i]
		}
	}
	if tags == nil || len(tags.Terms) != 1 || tags.Terms[0].Name != "go" || tags.Terms[0].Aliases[0] != "golang" {
		t.Errorf("taxonomía tags = %+v", tags)
	}
}

func TestSchemaCoversPostFrontmatter(t *testing.T) {
	checkCoverage(t, BuildSchema(domain.DefaultConfig()), "blog", reflect.TypeOf(postFrontmatter{}))
}

func TestSchemaCoversDocFrontmatter(t *testing.T) {
	checkCoverage(t, BuildSchema(domain.DefaultConfig()), "docs", reflect.TypeOf(docFrontmatter{}))
}

func checkCoverage(t *testing.T, s Schema, collection string, typ reflect.Type) {
	t.Helper()
	c := schemaCollection(t, s, collection)
	fields := map[string]bool{}
	for _, f := range c.Fields {
		fields[f.Name] = true
	}
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.Split(typ.Field(i).Tag.Get("yaml"), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		if !fields[tag] {
			t.Errorf("%s: schema sin campo %q", collection, tag)
		}
		delete(fields, tag)
	}
	for name := range fields {
		t.Errorf("%s: schema con campo extra %q", collection, name)
	}
}

func schemaCollection(t *testing.T, s Schema, name string) SchemaCollection {
	t.Helper()
	for _, c := range s.Collections {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("colección %q no encontrada", name)
	return SchemaCollection{}
}

func hasSchemaField(fields []SchemaField, name string) bool {
	for _, f := range fields {
		if f.Name == name {
			return true
		}
	}
	return false
}
