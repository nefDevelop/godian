// Package server implementa el servidor de desarrollo con recarga automática.
package server

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const reloadSnippet = `<script>new EventSource('/__reload').onmessage=function(){location.reload()}</script>`

// Server sirve el sitio generado y lo reconstruye al detectar cambios.
type Server struct {
	addr      string
	dir       string
	watch     []string
	rebuild   func() error
	mu        sync.Mutex
	clients   map[chan struct{}]struct{}
	signature string
}

// New crea un Server. watch son las rutas a vigilar; rebuild se invoca al cambiar.
func New(addr, dir string, watch []string, rebuild func() error) *Server {
	return &Server{
		addr:    addr,
		dir:     dir,
		watch:   watch,
		rebuild: rebuild,
		clients: map[chan struct{}]struct{}{},
	}
}

// Run arranca el servidor (bloquea hasta que falle o se cancele el contexto).
func (s *Server) Run(ctx context.Context) error {
	s.signature = signature(s.watch)
	go s.watchLoop(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/__reload", s.handleReload)
	mux.HandleFunc("/", s.handleStatic)

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	host := s.addr
	if strings.HasPrefix(host, ":") {
		host = "localhost" + host
	}
	fmt.Printf("Dev server en http://%s\n", host)
	return srv.ListenAndServe()
}

func (s *Server) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sig := signature(s.watch)
			if sig == s.signature {
				continue
			}
			s.signature = sig
			if err := s.rebuild(); err != nil {
				fmt.Fprintln(os.Stderr, "rebuild error:", err)
				continue
			}
			fmt.Println("reconstruido; recargando navegadores")
			s.notify()
		}
	}
}

func (s *Server) notify() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming no soportado", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			fmt.Fprint(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	upath := filepath.Clean(r.URL.Path)
	full := filepath.Join(s.dir, filepath.FromSlash(upath))
	if info, err := os.Stat(full); err == nil && info.IsDir() {
		full = filepath.Join(full, "index.html")
	}
	data, err := os.ReadFile(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(full, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		index := strings.LastIndex(string(data), "</body>")
		if index >= 0 {
			data = []byte(string(data[:index]) + reloadSnippet + string(data[index:]))
		} else {
			data = append(data, reloadSnippet...)
		}
	} else if ct := contentType(full); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

// contentType resuelve el MIME por extensión (necesario para CSS/JS, que
// http.DetectContentType no reconoce y serviría como text/plain).
func contentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	switch ext {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	}
	return ""
}

// signature devuelve una huella del estado (última modificación) de las rutas.
func signature(paths []string) string {
	var latest int64
	var count int
	for _, root := range paths {
		_ = filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			count++
			if mt := info.ModTime().UnixNano(); mt > latest {
				latest = mt
			}
			return nil
		})
	}
	return fmt.Sprintf("%d-%d", latest, count)
}
