package health

import (
	"encoding/json"
	"net/http"

	"github.com/cyandie/backend/internal/core"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	core.BaseModule
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Name() string { return "health" }

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/healthz", h.ping)
}

func (h *Handler) ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
