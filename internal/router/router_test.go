package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"go-web-bot/internal/config"
)

func TestFrontendFallbackServesSPAHistoryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dist := t.TempDir()
	index := []byte("<html><body>admin app</body></html>")
	if err := os.WriteFile(filepath.Join(dist, "index.html"), index, 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	r := gin.New()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	registerFrontend(r, config.Config{FrontendDist: dist, AdminRoutePrefix: "/api/admin"})

	for _, path := range []string{"/", "/login", "/users", "/bot"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, w.Code, http.StatusOK)
		}
		if w.Body.String() != string(index) {
			t.Fatalf("GET %s body = %q, want %q", path, w.Body.String(), string(index))
		}
	}
}

func TestFrontendFallbackServesDistAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dist := t.TempDir()
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	assetsDir := filepath.Join(dist, "assets")
	if err := os.Mkdir(assetsDir, 0o755); err != nil {
		t.Fatalf("make assets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	r := gin.New()
	registerFrontend(r, config.Config{FrontendDist: dist, AdminRoutePrefix: "/api/admin"})

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("asset status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "console.log('ok')" {
		t.Fatalf("asset body = %q", w.Body.String())
	}
}

func TestMissingFrontendDistKeepsServiceRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerFrontend(r, config.Config{FrontendDist: filepath.Join(t.TempDir(), "missing"), AdminRoutePrefix: "/api/admin"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("root status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "telegram bot admin service" {
		t.Fatalf("root body = %q", w.Body.String())
	}
}

func TestFrontendFallbackDoesNotMaskBackend404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dist := t.TempDir()
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	r := gin.New()
	registerFrontend(r, config.Config{FrontendDist: dist, AdminRoutePrefix: "/api/admin"})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("backend 404 status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if w.Body.String() == "index" {
		t.Fatal("backend 404 was masked by frontend index.html")
	}
}

func TestFrontendFallbackDoesNotMaskExactBackendPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dist := t.TempDir()
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	r := gin.New()
	registerFrontend(r, config.Config{FrontendDist: dist, AdminRoutePrefix: "/api/admin"})

	for _, path := range []string{"/api/admin", "/telegram", "/app-config.js"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("GET %s status = %d, want %d", path, w.Code, http.StatusNotFound)
		}
		if w.Body.String() == "index" {
			t.Fatalf("GET %s was masked by frontend index.html", path)
		}
	}
}

func TestNewRegistersRoutesWithoutPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = New(config.Config{AdminRoutePrefix: "/api/admin", FrontendDist: filepath.Join(t.TempDir(), "missing")}, nil)
}
