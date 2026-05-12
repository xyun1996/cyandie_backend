package chat

import (
	"encoding/json"
	"net/http"
	"strconv"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *ChatService
}

func NewHandler(svc *ChatService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/api/v1/chat/rooms", h.listRooms)
	router.Post("/api/v1/chat/rooms", h.createRoom)
	router.Get("/api/v1/chat/rooms/{id}/messages", h.getMessages)
}

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	// TODO: implement room listing
	writeChatJSON(w, http.StatusOK, []any{})
}

func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}
	// TODO: implement room creation
	writeChatJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handler) getMessages(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	_, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	// TODO: implement message history query
	_ = roomID
	writeChatJSON(w, http.StatusOK, []any{})
}

func writeChatJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}
