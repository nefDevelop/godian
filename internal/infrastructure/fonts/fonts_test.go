package fonts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchDownloadsAndRewrites(t *testing.T) {
	var srv *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/css2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		// El CSS apunta al propio servidor de pruebas.
		_, _ = io.WriteString(w, "/* latin */\n"+
			"@font-face { font-family: 'Inter'; font-weight: 400; src: url("+srv.URL+"/inter.woff2) format('woff2'); }\n"+
			"/* cyrillic */\n"+
			"@font-face { font-family: 'Inter'; font-weight: 400; src: url("+srv.URL+"/inter-cyr.woff2) format('woff2'); }\n")
	})
	mux.HandleFunc("/inter.woff2", func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "WOFF2DATA") })
	mux.HandleFunc("/inter-cyr.woff2", func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "CYR") })
	srv = httptest.NewServer(mux)
	defer srv.Close()

	res, err := Fetch(srv.Client(), []Spec{{Family: "Inter", Weights: []int{400}}}, srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.Contains(string(res.CSS), "/assets/fonts/inter.woff2") {
		t.Errorf("CSS sin URL local: %s", res.CSS)
	}
	if strings.Contains(string(res.CSS), "inter-cyr") {
		t.Errorf("subconjunto cyrillic no filtrado: %s", res.CSS)
	}
	if _, ok := res.Files["assets/fonts/inter.woff2"]; !ok {
		t.Errorf("falta el woff2 descargado: %v", res.Files)
	}
}

func TestBuildQuery(t *testing.T) {
	q := buildQuery([]Spec{{Family: "Inter", Weights: []int{400, 700}}})
	if !strings.Contains(q, "family=Inter:wght@400;700") || !strings.Contains(q, "display=swap") {
		t.Errorf("query = %s", q)
	}
	q2 := buildQuery([]Spec{{Family: "Roboto", Weights: []int{400}, Styles: []string{"normal", "italic"}, Display: "optional"}})
	if !strings.Contains(q2, "ital,wght@0,400;1,400") || !strings.Contains(q2, "display=optional") {
		t.Errorf("query italic = %s", q2)
	}
}

func TestFetchEmpty(t *testing.T) {
	res, err := Fetch(nil, nil, "")
	if err != nil || len(res.Files) != 0 {
		t.Errorf("Fetch vacío = %+v, %v", res, err)
	}
}
