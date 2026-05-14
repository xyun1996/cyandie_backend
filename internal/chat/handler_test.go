package chat

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyandie/backend/internal/auth"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Helper: build a chi.Router with the chat handler routes, optionally
// injecting a userID into the request context to simulate authentication.
// ---------------------------------------------------------------------------

func setupHandlerRouter(q *mockQuerier) *chi.Mux {
	srv := NewTCPServer("127.0.0.1:0", 0, 0)
	svc := NewChatService(q, srv, nil, nil, nil)
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func authCtx(userID string) context.Context {
	return context.WithValue(context.Background(), auth.UserIDKey, userID)
}

// ---------------------------------------------------------------------------
// listRooms tests
// ---------------------------------------------------------------------------

func TestListRooms_Unauthorized(t *testing.T) {
	r := setupHandlerRouter(&mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	// No user ID in context → 401
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestListRooms_InvalidUserID(t *testing.T) {
	r := setupHandlerRouter(&mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req = req.WithContext(authCtx("not-a-uuid"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestListRooms_Success(t *testing.T) {
	uid := uuid.New()
	q := &mockQuerier{
		rooms: []db.ChatRoom{
			{ID: uuid.New(), Type: "group", Name: sql.NullString{String: "general", Valid: true}},
			{ID: uuid.New(), Type: "dm", Name: sql.NullString{String: "", Valid: false}},
		},
	}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var rooms []db.ChatRoom
	if err := json.NewDecoder(rr.Body).Decode(&rooms); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}
}

func TestListRooms_DBError(t *testing.T) {
	uid := uuid.New()
	q := &mockQuerier{roomsErr: sql.ErrConnDone}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// createRoom tests
// ---------------------------------------------------------------------------

func TestCreateRoom_Unauthorized(t *testing.T) {
	r := setupHandlerRouter(&mockQuerier{})

	body := `{"type":"group","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestCreateRoom_InvalidBody(t *testing.T) {
	uid := uuid.New()
	r := setupHandlerRouter(&mockQuerier{})

	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString("not-json"))
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCreateRoom_Success(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{
		room: db.ChatRoom{
			ID:   roomID,
			Type: "group",
			Name: sql.NullString{String: "general", Valid: true},
		},
		member: db.ChatRoomMember{
			ID:     uuid.New(),
			RoomID: roomID,
			UserID: uid,
			Role:   "owner",
		},
	}
	r := setupHandlerRouter(q)

	body := `{"type":"group","name":"general"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString(body))
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var room db.ChatRoom
	if err := json.NewDecoder(rr.Body).Decode(&room); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if room.ID != roomID {
		t.Errorf("expected room ID %s, got %s", roomID, room.ID)
	}
}

func TestCreateRoom_AddRoomMemberError(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{
		room: db.ChatRoom{
			ID:   roomID,
			Type: "group",
			Name: sql.NullString{String: "general", Valid: true},
		},
		membErr2: sql.ErrConnDone,
	}
	r := setupHandlerRouter(q)

	body := `{"type":"group","name":"general"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString(body))
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on AddRoomMember error, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateRoom_DBError(t *testing.T) {
	uid := uuid.New()
	q := &mockQuerier{roomErr: sql.ErrConnDone}
	r := setupHandlerRouter(q)

	body := `{"type":"group","name":"fail"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString(body))
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// getMessages tests
// ---------------------------------------------------------------------------

func TestGetMessages_Unauthorized(t *testing.T) {
	r := setupHandlerRouter(&mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+uuid.New().String()+"/messages", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestGetMessages_InvalidRoomID(t *testing.T) {
	uid := uuid.New()
	r := setupHandlerRouter(&mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/rooms/not-a-uuid/messages", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestGetMessages_RoomNotFound(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{membErr: sql.ErrNoRows}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/messages", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestGetMessages_NotMember(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	otherUser := uuid.New()
	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: otherUser, Role: "owner"},
		},
	}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/messages", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestGetMessages_Success(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: uid, Role: "member"},
		},
		messages: []db.ChatMessage{
			{ID: uuid.New(), RoomID: roomID, SenderID: uid, Content: "hello", Type: "text"},
			{ID: uuid.New(), RoomID: roomID, SenderID: uid, Content: "world", Type: "text"},
		},
	}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/messages", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var msgs []db.ChatMessage
	if err := json.NewDecoder(rr.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestGetMessages_MessagesDBError(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: uid, Role: "member"},
		},
		msgsErr: sql.ErrConnDone,
	}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/messages", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestGetMessages_PaginationDefaults(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: uid, Role: "member"},
		},
		messages: []db.ChatMessage{},
	}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/messages", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if q.lastGetMessagesLimit != 50 {
		t.Errorf("expected default limit 50, got %d", q.lastGetMessagesLimit)
	}
	if q.lastGetMessagesOffset != 0 {
		t.Errorf("expected default offset 0, got %d", q.lastGetMessagesOffset)
	}
}

func TestGetMessages_PaginationCustom(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: uid, Role: "member"},
		},
		messages: []db.ChatMessage{},
	}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/messages?limit=100&offset=50", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if q.lastGetMessagesLimit != 100 {
		t.Errorf("expected limit 100, got %d", q.lastGetMessagesLimit)
	}
	if q.lastGetMessagesOffset != 50 {
		t.Errorf("expected offset 50, got %d", q.lastGetMessagesOffset)
	}
}

func TestGetMessages_PaginationMaxLimit(t *testing.T) {
	uid := uuid.New()
	roomID := uuid.New()
	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: uid, Role: "member"},
		},
		messages: []db.ChatMessage{},
	}
	r := setupHandlerRouter(q)

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String()+"/messages?limit=999&offset=0", nil)
	req = req.WithContext(authCtx(uid.String()))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if q.lastGetMessagesLimit != 200 {
		t.Errorf("expected limit clamped to 200, got %d", q.lastGetMessagesLimit)
	}
}
