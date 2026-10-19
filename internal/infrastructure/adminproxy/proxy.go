// Package adminproxy reenvía las peticiones del admin estático al proveedor Git.
// Evita CORS en desarrollo y en despliegues con Functions same-origin.
package adminproxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nefdevelop/godian/internal/domain"
)

// Proxy es un pass-through sin secretos: reenvía el Authorization del navegador.
type Proxy struct {
	provider string
	base     string
	owner    string
	repo     string
	client   *http.Client
}

// New construye el proxy para admin.api_mode = proxy.
func New(cfg domain.AdminConfig) (*Proxy, error) {
	if err := domain.NormalizeAdmin(&cfg); err != nil {
		return nil, err
	}
	if cfg.APIMode != "proxy" {
		return nil, fmt.Errorf("adminproxy: api_mode debe ser proxy")
	}
	var base string
	switch cfg.Provider {
	case "gitea":
		base = strings.TrimRight(cfg.BaseURL, "/") + "/api/v1"
	case "github":
		base = "https://api.github.com"
	default:
		return nil, fmt.Errorf("adminproxy: proveedor no soportado: %s", cfg.Provider)
	}
	return &Proxy{
		provider: cfg.Provider,
		base:     base,
		owner:    cfg.Owner,
		repo:     cfg.Repo,
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ServeHTTP espera /admin/api/repos/{owner}/{repo}/...
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodOptions:
	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if r.Method == http.MethodOptions {
		w.Header().Set("Allow", "GET, PUT, DELETE, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/admin/api")
	if !strings.HasPrefix(path, "/repos/") {
		http.NotFound(w, r)
		return
	}
	rest := strings.TrimPrefix(path, "/repos/")
	parts := strings.SplitN(rest, "/", 3)
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	owner, err := url.PathUnescape(parts[0])
	if err != nil || owner != p.owner {
		http.Error(w, "repositorio no permitido", http.StatusForbidden)
		return
	}
	repo, err := url.PathUnescape(parts[1])
	if err != nil || repo != p.repo {
		http.Error(w, "repositorio no permitido", http.StatusForbidden)
		return
	}

	target := p.base + "/repos/" + url.PathEscape(p.owner) + "/" + url.PathEscape(p.repo)
	if len(parts) == 3 && parts[2] != "" {
		tail := parts[2]
		if tail != "contents" && !strings.HasPrefix(tail, "contents/") {
			http.Error(w, "ruta no permitida", http.StatusForbidden)
			return
		}
		for _, segment := range strings.Split(tail, "/") {
			if segment == ".." || segment == "." {
				http.Error(w, "ruta no permitida", http.StatusForbidden)
				return
			}
		}
		target += "/" + tail
	}
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		http.Error(w, "petición inválida", http.StatusBadRequest)
		return
	}
	for _, key := range []string{"Authorization", "Content-Type", "Accept", "X-GitHub-Api-Version"} {
		if value := r.Header.Get(key); value != "" {
			req.Header.Set(key, value)
		}
	}
	req.Header.Set("User-Agent", "godian-admin-proxy")

	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, "no se pudo contactar con el proveedor Git", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Cache-Control", "no-store")
	for key, values := range resp.Header {
		switch strings.ToLower(key) {
		case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
			"te", "trailer", "transfer-encoding", "upgrade", "content-length":
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
