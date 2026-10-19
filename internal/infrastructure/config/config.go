// Package config carga la configuración del sitio desde site.yaml.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/nefdevelop/godian/internal/domain"
)

// Load lee el archivo indicado y aplica los valores por defecto que falten.
func Load(path string) (domain.Config, error) {
	cfg := domain.DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, fmt.Errorf("no se encontró la configuración %q (usa -config)", path)
		}
		return cfg, fmt.Errorf("leyendo %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("configuración inválida en %q: %w", path, err)
	}
	if err := applyDefaults(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// applyDefaults completa las colecciones y taxonomías estándar cuando no se declaran.
func applyDefaults(cfg *domain.Config) error {
	domain.NormalizeSite(&cfg.Site)
	if len(cfg.Collections) == 0 {
		cfg.Collections = domain.DefaultCollections()
	}
	if len(cfg.Taxonomies) == 0 {
		cfg.Taxonomies = domain.DefaultTaxonomies()
	}
	for i := range cfg.Taxonomies {
		if cfg.Taxonomies[i].URL == "" {
			cfg.Taxonomies[i].URL = "/" + cfg.Taxonomies[i].Name
		}
	}
	return domain.NormalizeAdmin(&cfg.Admin)
}
