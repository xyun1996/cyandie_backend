package platforms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *PlatformService
}

func NewHandler(svc *PlatformService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/api/v1/platforms", h.listPlatforms)
	router.Get("/api/v1/platforms/{name}/auth-url", h.getAuthURL)
	router.Post("/api/v1/platforms/{name}/callback", h.callback)
	router.Post("/api/v1/platforms/{name}/bind", h.bind)
	router.Delete("/api/v1/platforms/{name}/bind", h.unbind)
	router.Get("/api/v1/platforms/bindings", h.listBindings)
}

func (h *Handler) listPlatforms(w http.ResponseWriter, r *http.Request) {
	platforms := h.svc.registry.ListPlatforms()
	type platformInfo struct {
		Name         string   `json:"name"`
		Capabilities []string `json:"capabilities"`
	}
	result := make([]platformInfo, len(platforms))
	for i, name := range platforms {
		result[i] = platformInfo{Name: name, Capabilities: []string{"oauth"}}
	}
	writePlatformJSON(w, http.StatusOK, result)
}

func (h *Handler) getAuthURL(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	state := r.URL.Query().Get("state")
	if state == "" {
		state = fmt.Sprintf("st_%d", time.Now().UnixNano())
	}
	url, err := h.svc.GetAuthURL(name, state)
	if err != nil {
		writePlatformAppError(w, err)
		return
	}
	writePlatformJSON(w, http.StatusOK, map[string]string{"url": url, "state": state})
}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}
	result, err := h.svc.HandleCallback(r.Context(), name, req.Code)
	if err != nil {
		writePlatformAppError(w, err)
		return
	}
	writePlatformJSON(w, http.StatusOK, result)
}

func (h *Handler) bind(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}
	name := chi.URLParam(r, "name")
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}
	if err := h.svc.BindPlatform(r.Context(), userID, name, req.Code); err != nil {
		writePlatformAppError(w, err)
		return
	}
	writePlatformJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) unbind(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}
	name := chi.URLParam(r, "name")
	if err := h.svc.UnbindPlatform(r.Context(), userID, name); err != nil {
		writePlatformAppError(w, err)
		return
	}
	writePlatformJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listBindings(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}
	bindings, err := h.svc.ListBindings(r.Context(), userID)
	if err != nil {
		writePlatformAppError(w, err)
		return
	}
	type bindingInfo struct {
		Platform       string `json:"platform"`
		PlatformUserID string `json:"platformUserId"`
	}
	result := make([]bindingInfo, len(bindings))
	for i, b := range bindings {
		result[i] = bindingInfo{Platform: b.Platform, PlatformUserID: b.PlatformUserID}
	}
	writePlatformJSON(w, http.StatusOK, result)
}

func writePlatformJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writePlatformAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*coreerrors.AppError); ok {
		appErr.WriteHTTP(w)
		return
	}
	coreerrors.New(coreerrors.ErrInternal, "internal error").WriteHTTP(w)
}
