package chat

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cyandie/backend/internal/auth"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type Handler struct {
	service *ChatService
}

func NewHandler(service *ChatService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/rooms", h.listRooms)
	r.Post("/rooms", h.createRoom)
	r.Get("/rooms/{roomID}/messages", h.getMessages)
}

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	rooms, err := h.service.queries.ListRoomsByUser(r.Context(), uid)
	if err != nil {
		http.Error(w, `{"error":"failed to list rooms"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

type createRoomRequest struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Metadata string `json:"metadata,omitempty"`
}

func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	room, err := h.service.queries.CreateChatRoom(r.Context(), db.CreateChatRoomParams{
		Type:     req.Type,
		Name:     sql.NullString{String: req.Name, Valid: req.Name != ""},
		Metadata: pqtype.NullRawMessage{RawMessage: []byte(req.Metadata), Valid: req.Metadata != ""},
	})
	if err != nil {
		http.Error(w, `{"error":"failed to create room"}`, http.StatusInternalServerError)
		return
	}

	// Add creator as member
	uid, _ := uuid.Parse(userID)
	if _, err := h.service.queries.AddRoomMember(r.Context(), db.AddRoomMemberParams{
		RoomID: room.ID,
		UserID: uid,
		Role:   "owner",
	}); err != nil {
		http.Error(w, `{"error":"failed to add room member"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

func (h *Handler) getMessages(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	roomID := chi.URLParam(r, "roomID")
	roomUID, err := uuid.Parse(roomID)
	if err != nil {
		http.Error(w, `{"error":"invalid room id"}`, http.StatusBadRequest)
		return
	}

	// Verify user is a member of the room
	members, err := h.service.queries.GetRoomMembers(r.Context(), roomUID)
	if err != nil {
		http.Error(w, `{"error":"room not found"}`, http.StatusNotFound)
		return
	}

	isMember := false
	for _, m := range members {
		if m.UserID.String() == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	limit := int32(50)
	offset := int32(0)
	if l, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32); err == nil && l > 0 {
		limit = int32(l)
		if limit > 200 {
			limit = 200
		}
	}
	if o, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32); err == nil && o >= 0 {
		offset = int32(o)
	}

	messages, err := h.service.queries.GetChatMessages(r.Context(), db.GetChatMessagesParams{
		RoomID: roomUID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to get messages"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
