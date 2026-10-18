// Package fonts descarga fuentes de Google Fonts y las autohospeda en el build:
// obtiene el CSS de @font-face, descarga los woff2 de los subconjuntos pedidos y
// devuelve un CSS con rutas locales (sin peticiones a terceros en tiempo de uso).
package fonts

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// Spec describe una fuente a descargar.
type Spec struct {
	Family  string
	Weights []int
	Styles  []string // normal | italic
	Subsets []string // latin, latin-ext, cyrillic…
	Display string   // swap | block | fallback | optional
}

// Result es el CSS generado y los ficheros a escribir (ruta relativa → datos).
type Result struct {
	CSS   []byte
	Files map[string][]byte // p. ej. "assets/fonts/<fichero>.woff2" → datos
}

// DefaultBaseURL es la API de Google Fonts.
const DefaultBaseURL = "https://fonts.googleapis.com"

// modernUA hace que la API devuelva woff2 (con UA antiguo devuelve ttf).
const modernUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36"

var (
	faceRe = regexp.MustCompile(`(?s)/\*\s*([^*]+?)\s*\*/\s*(@font-face\s*\{.*?\})`)
	urlRe  = regexp.MustCompile(`url\((https?://[^)]+?)\)`)
)

// Fetch descarga las fuentes y devuelve los @font-face con URLs locales.
// baseURL vacío usa Google Fonts; se puede apuntar a un servidor de pruebas.
func Fetch(client *http.Client, specs []Spec, baseURL string) (Result, error) {
	if len(specs) == 0 {
		return Result{}, nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	css, err := getText(client, baseURL+"/css2?"+buildQuery(specs))
	if err != nil {
		return Result{}, err
	}

	// Subconjuntos a conservar (por defecto, latin).
	subsets := map[string]bool{}
	for _, s := range specs {
		for _, sub := range s.Subsets {
			subsets[strings.TrimSpace(sub)] = true
		}
	}
	if len(subsets) == 0 {
		subsets["latin"] = true
	}

	res := Result{Files: map[string][]byte{}}
	downloaded := map[string]string{} // url remota → ruta local
	var out strings.Builder

	for _, m := range faceRe.FindAllStringSubmatch(css, -1) {
		subset := strings.TrimSpace(m[1])
		if !subsets[subset] {
			continue
		}
		block := m[2]
		for _, u := range urlRe.FindAllStringSubmatch(block, -1) {
			remote := u[1]
			local, ok := downloaded[remote]
			if !ok {
				data, derr := getBytes(client, remote)
				if derr != nil {
					return Result{}, derr
				}
				local = "assets/fonts/" + path.Base(remote)
				res.Files[local] = data
				downloaded[remote] = local
			}
			block = strings.ReplaceAll(block, remote, "/"+local)
		}
		out.WriteString(block)
		out.WriteString("\n")
	}
	res.CSS = []byte(out.String())
	return res, nil
}

// buildQuery construye la querystring de la API css2.
func buildQuery(specs []Spec) string {
	var parts []string
	display := "swap"
	for _, s := range specs {
		family := strings.TrimSpace(s.Family)
		if family == "" {
			continue
		}
		if s.Display != "" {
			display = s.Display
		}
		weights := s.Weights
		if len(weights) == 0 {
			weights = []int{400}
		}
		styles := s.Styles
		if len(styles) == 0 {
			styles = []string{"normal"}
		}
		hasItalic := false
		for _, st := range styles {
			if st == "italic" {
				hasItalic = true
			}
		}
		var pairs []string
		if hasItalic {
			for _, st := range styles {
				ital := "0"
				if st == "italic" {
					ital = "1"
				}
				for _, w := range weights {
					pairs = append(pairs, ital+","+strconv.Itoa(w))
				}
			}
			parts = append(parts, "family="+url.QueryEscape(family)+":ital,wght@"+strings.Join(pairs, ";"))
		} else {
			for _, w := range weights {
				pairs = append(pairs, strconv.Itoa(w))
			}
			parts = append(parts, "family="+url.QueryEscape(family)+":wght@"+strings.Join(pairs, ";"))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(append(parts, "display="+display), "&")
}

func getText(client *http.Client, u string) (string, error) {
	b, err := getBytes(client, u)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getBytes(client *http.Client, u string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", modernUA)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fonts: %s devolvio %d", u, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
