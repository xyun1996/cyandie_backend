package friends

import (
	"encoding/json"
	"net/http"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

type FriendsHandler struct {
	service *FriendsService
}

func NewFriendsHandler(service *FriendsService) *FriendsHandler {
	return &FriendsHandler{service: service}
}

func (h *FriendsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/request", h.SendRequest)
	r.Put("/{id}/accept", h.AcceptRequest)
	r.Delete("/{id}/reject", h.RejectRequest)
	r.Get("/", h.ListFriends)
	r.Get("/pending", h.ListPendingRequests)
	r.Get("/online", h.GetOnlineFriends)
	return r
}

func (h *FriendsHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ToUserID string `json:"to_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	fromID, _ := r.Context().Value("userID").(string)
	friendship, err := h.service.SendRequest(r.Context(), fromID, body.ToUserID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, friendship)
}

func (h *FriendsHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	friendship, err := h.service.AcceptRequest(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, friendship)
}

func (h *FriendsHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.RejectRequest(r.Context(), id); err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (h *FriendsHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(string)
	friendships, err := h.service.ListFriends(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, friendships)
}

func (h *FriendsHandler) ListPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(string)
	requests, err := h.service.ListPendingRequests(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

func (h *FriendsHandler) GetOnlineFriends(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(string)
	online, err := h.service.GetOnlineFriends(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, online)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writeAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*coreerrors.AppError); ok {
		appErr.WithRequestID(w.Header().Get("X-Request-ID")).WriteHTTP(w)
		return
	}
	coreerrors.New(coreerrors.ErrInternal, "internal error").WriteHTTP(w)
}
