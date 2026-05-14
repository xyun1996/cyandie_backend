package friends

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cyandie/backend/internal/auth"
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
	r.Post("/block", h.Block)
	r.Delete("/block/{userID}", h.Unblock)
	r.Get("/blocked", h.ListBlocked)
	r.Delete("/{userID}", h.RemoveFriend)
	r.Get("/recent", h.ListRecentContacts)
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

	fromID := auth.UserIDFromContext(r.Context())
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
	userID := auth.UserIDFromContext(r.Context())
	friendships, err := h.service.ListFriends(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, friendships)
}

func (h *FriendsHandler) ListPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	requests, err := h.service.ListPendingRequests(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

func (h *FriendsHandler) GetOnlineFriends(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	online, err := h.service.GetOnlineFriends(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, online)
}

func (h *FriendsHandler) Block(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}
	if err := h.service.Block(r.Context(), userID, req.UserID, req.Reason); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendsHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	blockedUserID := chi.URLParam(r, "userID")
	if err := h.service.Unblock(r.Context(), userID, blockedUserID); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendsHandler) ListBlocked(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	blocked, err := h.service.ListBlockedUsers(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, blocked)
}

func (h *FriendsHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	friendUserID := chi.URLParam(r, "userID")
	if err := h.service.RemoveFriend(r.Context(), userID, friendUserID); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendsHandler) ListRecentContacts(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	contacts, err := h.service.ListRecentContacts(r.Context(), userID, limit)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, contacts)
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
