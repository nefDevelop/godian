package adminproxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nefdevelop/godian/internal/domain"
)

func TestProxyForwardsRepoPathAndAuth(t *testing.T) {
	var gotPath, gotAuth, gotQuery string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sha":"abc"}`))
	}))
	defer upstream.Close()

	proxy, err := New(domain.AdminConfig{
		Mode:     "remote",
		Provider: "gitea",
		BaseURL:  upstream.URL,
		Owner:    "pi",
		Repo:     "godian",
		Branch:   "main",
		APIMode:  "proxy",
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/admin/api/repos/pi/godian/contents/content/blog/x.md?ref=main", nil)
	req.Header.Set("Authorization", "token secreto")
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if gotPath != "/api/v1/repos/pi/godian/contents/content/blog/x.md" {
		t.Errorf("upstream path = %q", gotPath)
	}
	if gotQuery != "ref=main" {
		t.Errorf("upstream query = %q", gotQuery)
	}
	if gotAuth != "token secreto" {
		t.Errorf("upstream auth = %q", gotAuth)
	}
	if !strings.Contains(rec.Body.String(), `"sha":"abc"`) {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestProxyRejectsOtherRepo(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no debería llamar al upstream")
	}))
	defer upstream.Close()

	proxy, err := New(domain.AdminConfig{
		Mode:     "remote",
		Provider: "gitea",
		BaseURL:  upstream.URL,
		Owner:    "pi",
		Repo:     "godian",
		APIMode:  "proxy",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/admin/api/repos/otro/godian/contents/x", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestProxyRejectsTraversal(t *testing.T) {
	proxy, err := New(domain.AdminConfig{
		Mode:     "remote",
		Provider: "gitea",
		BaseURL:  "https://git.example.com",
		Owner:    "pi",
		Repo:     "godian",
		APIMode:  "proxy",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/admin/api/repos/pi/godian/contents/../../otro/repo", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestProxyRejectsMethod(t *testing.T) {
	proxy, err := New(domain.AdminConfig{
		Mode:     "remote",
		Provider: "gitea",
		BaseURL:  "https://git.example.com",
		Owner:    "pi",
		Repo:     "godian",
		APIMode:  "proxy",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/admin/api/repos/pi/godian/contents/x", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
