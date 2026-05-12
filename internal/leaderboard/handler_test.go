package leaderboard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_GetRanking(t *testing.T) {
	svc := newTestLBService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard/test-board?limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SubmitScore(t *testing.T) {
	svc := newTestLBService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/leaderboard/test-board/submit", nil)
	req.Header.Set("X-User-ID", "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should get 400 since no body, but at least not 500
	if rec.Code == http.StatusInternalServerError {
		t.Errorf("expected not 500, got %d", rec.Code)
	}
}
