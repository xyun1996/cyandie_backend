package platforms

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_ListPlatforms(t *testing.T) {
	svc := newTestPlatformService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	r.Route("/api/v1/platforms", h.RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/platforms/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAuthURL(t *testing.T) {
	svc := newTestPlatformService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	r.Route("/api/v1/platforms", h.RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/platforms/wechat/auth-url?state=test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAuthURL_UnsupportedPlatform(t *testing.T) {
	svc := newTestPlatformService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	r.Route("/api/v1/platforms", h.RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/platforms/google/auth-url", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
