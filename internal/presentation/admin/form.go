package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/infrastructure/content"
)

// FieldView es un campo del schema ya preparado para la plantilla.
type FieldView struct {
	Name         string
	Label        string
	Type         string
	Required     bool
	Value        string
	Bool         bool
	Options      []string
	Fields       []FieldView
	JSON         string
	IsListObject bool
	InputType    string
}

func buildFields(col content.SchemaCollection, fm map[string]any, cfg domain.Config, schema content.Schema) []FieldView {
	fields := make([]FieldView, 0, len(col.Fields))
	for _, f := range col.Fields {
		fields = append(fields, buildField(f, f.Name, fm[f.Name], cfg, schema))
	}
	return fields
}

func buildField(f content.SchemaField, name string, raw any, cfg domain.Config, schema content.Schema) FieldView {
	view := FieldView{
		Name:     name,
		Label:    f.Label,
		Type:     f.Type,
		Required: f.Required,
	}

	switch f.Type {
	case "object":
		view.InputType = "object"
		nested := mapValue(raw)
		for _, child := range f.Fields {
			view.Fields = append(view.Fields,
				buildField(child, name+"."+child.Name, nested[child.Name], cfg, schema))
		}
	case "list":
		if len(f.Fields) > 0 {
			view.InputType = "json"
			view.IsListObject = true
			view.JSON = prettyJSON(raw)
		} else {
			view.InputType = "list"
			view.Value = joinList(raw)
			view.Options = sourceOptions(f, cfg, schema)
		}
	case "select":
		view.InputType = "select"
		view.Value = stringValue(raw)
		view.Options = mergeOptions(f.Options, sourceOptions(f, cfg, schema))
	case "boolean":
		view.InputType = "checkbox"
		view.Bool = boolValue(raw)
	case "datetime":
		view.InputType = "date"
		view.Value = formatDateValue(raw)
	case "number":
		view.InputType = "number"
		view.Value = stringValue(raw)
	case "image":
		view.InputType = "image"
		view.Value = stringValue(raw)
	case "text":
		view.InputType = "textarea"
		view.Value = stringValue(raw)
	default:
		view.InputType = "text"
		view.Value = stringValue(raw)
	}
	return view
}

func parseForm(col content.SchemaCollection, r *http.Request, cfg domain.Config) (map[string]any, error) {
	fm := map[string]any{}
	for _, f := range col.Fields {
		name := f.Name
		switch f.Type {
		case "object":
			nested := map[string]any{}
			for _, child := range f.Fields {
				full := name + "." + child.Name
				value, set, err := parseScalar(child, full, r, cfg)
				if err != nil {
					return fm, err
				}
				if set {
					nested[child.Name] = value
				}
			}
			if len(nested) > 0 {
				fm[name] = nested
			}
		case "list":
			if len(f.Fields) > 0 {
				raw := strings.TrimSpace(r.FormValue(name + "_json"))
				if raw == "" || raw == "[]" {
					continue
				}
				var items []map[string]any
				if err := json.Unmarshal([]byte(raw), &items); err != nil {
					return fm, fmt.Errorf("%s: JSON inválido: %w", f.Label, err)
				}
				if len(items) > 0 {
					fm[name] = items
				}
				continue
			}
			parts := splitList(r.FormValue(name))
			if len(parts) > 0 {
				fm[name] = parts
			}
		default:
			value, set, err := parseScalar(f, name, r, cfg)
			if err != nil {
				return fm, err
			}
			if set {
				fm[name] = value
			}
		}
	}

	for _, f := range col.Fields {
		if f.Required && isEmpty(fm[f.Name]) {
			return fm, fmt.Errorf("el campo %q es obligatorio", f.Label)
		}
	}
	return fm, nil
}

// parseScalar convierte un campo simple del formulario. Devuelve set=false
// cuando el usuario lo dejó vacío (o desmarcó un checkbox opcional).
func parseScalar(f content.SchemaField, name string, r *http.Request, cfg domain.Config) (any, bool, error) {
	raw := strings.TrimSpace(r.FormValue(name))
	switch f.Type {
	case "boolean":
		_, checked := r.Form[name]
		return checked, true, nil
	case "number":
		if raw == "" {
			return nil, false, nil
		}
		n, err := parseNumber(raw)
		if err != nil {
			return nil, false, fmt.Errorf("%s: número inválido: %w", f.Label, err)
		}
		return n, true, nil
	case "datetime":
		if raw == "" {
			return nil, false, nil
		}
		if _, err := time.Parse("2006-01-02", raw); err != nil {
			return nil, false, fmt.Errorf("%s: fecha inválida (usa YYYY-MM-DD)", f.Label)
		}
		return yamlDate(raw), true, nil
	case "list":
		parts := splitList(raw)
		if len(parts) == 0 {
			return nil, false, nil
		}
		return parts, true, nil
	case "image":
		file, header, err := r.FormFile("upload_" + name)
		if err == nil {
			defer file.Close()
			public, saveErr := saveUpload(cfg, file, header)
			if saveErr != nil {
				return nil, false, saveErr
			}
			return public, true, nil
		}
		if !errors.Is(err, http.ErrMissingFile) {
			return nil, false, fmt.Errorf("%s: %w", f.Label, err)
		}
		if raw == "" {
			return nil, false, nil
		}
		return raw, true, nil
	default:
		if raw == "" {
			return nil, false, nil
		}
		return raw, true, nil
	}
}

func parseNumber(raw string) (any, error) {
	if n, err := strconv.Atoi(raw); err == nil {
		return n, nil
	}
	return strconv.ParseFloat(raw, 64)
}

func sourceOptions(f content.SchemaField, cfg domain.Config, schema content.Schema) []string {
	source := strings.TrimSpace(f.Source)
	if source == "" {
		return nil
	}
	if name, ok := strings.CutPrefix(source, "taxonomy:"); ok {
		for _, t := range schema.Taxonomies {
			if t.Name != name {
				continue
			}
			terms := make([]string, 0, len(t.Terms))
			for _, term := range t.Terms {
				if term.Name != "" {
					terms = append(terms, term.Name)
				}
			}
			return terms
		}
		return nil
	}
	switch source {
	case "authors":
		return yamlNames(cfg, "authors", "name")
	case "series":
		return yamlNames(cfg, "series", "title")
	default:
		return nil
	}
}

func mergeOptions(groups ...[]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, group := range groups {
		for _, item := range group {
			item = strings.TrimSpace(item)
			if item == "" || seen[item] {
				continue
			}
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

func mapValue(v any) map[string]any {
	switch value := v.(type) {
	case map[string]any:
		return value
	case map[any]any:
		out := make(map[string]any, len(value))
		for k, item := range value {
			out[fmt.Sprint(k)] = item
		}
		return out
	default:
		return map[string]any{}
	}
}

func boolValue(v any) bool {
	switch value := v.(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "1", "on", "yes", "sí", "si":
			return true
		}
	}
	return false
}

func joinList(v any) string {
	switch value := v.(type) {
	case nil:
		return ""
	case []string:
		return strings.Join(value, ", ")
	case []any:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			parts = append(parts, stringValue(item))
		}
		return strings.Join(parts, ", ")
	default:
		return stringValue(value)
	}
}

func splitList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	seen := map[string]bool{}
	var out []string
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" || seen[field] {
			continue
		}
		seen[field] = true
		out = append(out, field)
	}
	return out
}

func isEmpty(v any) bool {
	switch value := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(value) == ""
	case yamlDate:
		return strings.TrimSpace(string(value)) == ""
	case []string:
		return len(value) == 0
	case []any:
		return len(value) == 0
	case map[string]any:
		return len(value) == 0
	default:
		return false
	}
}

func prettyJSON(v any) string {
	if v == nil {
		return "[]"
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "[]"
	}
	if string(data) == "null" {
		return "[]"
	}
	return string(data)
}
