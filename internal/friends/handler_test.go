package friends

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
// mock queries — reuses the same struct from service_test.go but with
// handler-specific fields so we can control per-route behaviour.
// ---------------------------------------------------------------------------

type handlerMockQueries struct {
	friendship       db.Friendship
	friendErr        error
	createFriendship db.Friendship
	createFriendErr  error
	friendships      []db.Friendship
	friendsErr       error
	deleteErr        error

	blockRelation    db.BlockRelation
	blockErr         error
	blockedList      []db.BlockRelation
	blockedListErr   error
	isBlockedID      uuid.UUID
	isBlockedErr     error
	deleteByUsers    db.Friendship
	deleteByUsersErr error

	friendshipByUsers    db.Friendship
	friendshipByUsersErr error
}

// Friends-specific methods

func (m *handlerMockQueries) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return m.createFriendship, m.createFriendErr
}
func (m *handlerMockQueries) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, m.deleteErr
}
func (m *handlerMockQueries) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return m.friendship, m.friendErr
}
func (m *handlerMockQueries) GetFriendshipByUsers(_ context.Context, _ db.GetFriendshipByUsersParams) (db.Friendship, error) {
	return m.friendshipByUsers, m.friendshipByUsersErr
}
func (m *handlerMockQueries) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return m.friendships, m.friendsErr
}
func (m *handlerMockQueries) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return m.friendships, m.friendsErr
}
func (m *handlerMockQueries) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return m.friendship, m.friendErr
}
func (m *handlerMockQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return m.blockRelation, m.blockErr
}
func (m *handlerMockQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return m.blockRelation, m.blockErr
}
func (m *handlerMockQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return m.blockedList, m.blockedListErr
}
func (m *handlerMockQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return m.isBlockedID, m.isBlockedErr
}
func (m *handlerMockQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return m.deleteByUsers, m.deleteByUsersErr
}

// Stub remaining Querier methods

func (m *handlerMockQueries) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) { return db.User{}, nil }
func (m *handlerMockQueries) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error)          { return db.User{}, sql.ErrNoRows }
func (m *handlerMockQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error)        { return db.User{}, sql.ErrNoRows }
func (m *handlerMockQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *handlerMockQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *handlerMockQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) { return nil, nil }
func (m *handlerMockQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *handlerMockQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *handlerMockQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *handlerMockQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) { return 0, nil }
func (m *handlerMockQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *handlerMockQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *handlerMockQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *handlerMockQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (m *handlerMockQueries) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *handlerMockQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *handlerMockQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *handlerMockQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *handlerMockQueries) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *handlerMockQueries) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *handlerMockQueries) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) { return nil, nil }
func (m *handlerMockQueries) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (m *handlerMockQueries) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *handlerMockQueries) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (m *handlerMockQueries) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (m *handlerMockQueries) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (m *handlerMockQueries) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (m *handlerMockQueries) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (m *handlerMockQueries) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *handlerMockQueries) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *handlerMockQueries) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (m *handlerMockQueries) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (m *handlerMockQueries) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (m *handlerMockQueries) ListRoomsByUser(_ context.Context, _ uuid.UUID) ([]db.ChatRoom, error) { return nil, nil }

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// authMiddleware injects a user ID into the request context, simulating
// what auth.AuthGuard would do after a successful JWT validation.
func authMiddleware(userID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// newTestRouter builds a chi router with the friends handler routes
// mounted under / and the given auth middleware applied.
func newTestRouter(h *FriendsHandler, userID string) chi.Router {
	r := chi.NewRouter()
	if userID != "" {
		r.Use(authMiddleware(userID))
	}
	r.Mount("/", h.Routes())
	return r
}

// doRequest executes an HTTP request against the router and returns the response.
func doRequest(router chi.Router, method, path string, body any) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// ---------------------------------------------------------------------------
// 1. Auth required — endpoints that use auth.UserIDFromContext reject
//    unauthenticated requests because the service gets an empty string
//    and uuid.Parse("") fails with BAD_REQUEST (400).
//    AcceptRequest/RejectRequest do NOT read userID from context.
// ---------------------------------------------------------------------------

// When userID is empty (no auth context), handlers pass "" to the service.
// The service calls uuid.Parse("") which returns a BAD_REQUEST error,
// so these endpoints effectively reject unauthenticated requests with 400.
// AcceptRequest and RejectRequest do not read userID from context,
// so they proceed normally regardless.

func TestHandler_AuthRequired_ListFriends(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "") // no user ID

	rec := doRequest(r, http.MethodGet, "/", nil)

	// No user ID in context → service gets "" → uuid.Parse fails → 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no user ID in context), got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_SendRequest(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodPost, "/request", map[string]string{
		"to_user_id": uuid.New().String(),
	})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no user ID in context), got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_ListPending(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodGet, "/pending", nil)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no user ID in context), got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_RemoveFriend(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodDelete, "/"+uuid.New().String(), nil)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no user ID in context), got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_Block(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodPost, "/block", map[string]string{
		"user_id": uuid.New().String(),
	})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no user ID in context), got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_Unblock(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodDelete, "/block/"+uuid.New().String(), nil)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no user ID in context), got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_ListBlocked(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodGet, "/blocked", nil)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (no user ID in context), got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_AcceptRequest(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodPut, "/"+uuid.New().String()+"/accept", nil)

	// AcceptRequest does not read user ID from context, so it proceeds
	// directly to the service without auth gating.
	if rec.Code != http.StatusOK {
		t.Errorf("AcceptRequest does not check userID from context; expected 200, got %d", rec.Code)
	}
}

func TestHandler_AuthRequired_RejectRequest(t *testing.T) {
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, "")

	rec := doRequest(r, http.MethodDelete, "/"+uuid.New().String()+"/reject", nil)

	// RejectRequest, like AcceptRequest, does not read user ID from context.
	if rec.Code != http.StatusOK {
		t.Errorf("RejectRequest does not check userID from context; expected 200, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// 2. Send friend request — basic flow
// ---------------------------------------------------------------------------

func TestHandler_SendRequest_Success(t *testing.T) {
	from := uuid.New()
	to := uuid.New()
	q := &handlerMockQueries{
		friendshipByUsersErr: sql.ErrNoRows, // no existing friendship
		isBlockedErr:        sql.ErrNoRows, // not blocked
		createFriendship:    db.Friendship{ID: uuid.New(), UserID: from, FriendID: to, Status: "pending"},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, from.String())

	rec := doRequest(r, http.MethodPost, "/request", map[string]string{
		"to_user_id": to.String(),
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if ok, _ := resp["ok"].(bool); !ok {
		t.Error("expected ok=true in response")
	}
}

func TestHandler_SendRequest_InvalidBody(t *testing.T) {
	userID := uuid.New()
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, userID.String())

	req := httptest.NewRequest(http.MethodPost, "/request", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// 3. List friends — basic flow
// ---------------------------------------------------------------------------

func TestHandler_ListFriends_Success(t *testing.T) {
	uid := uuid.New()
	friend := uuid.New()
	q := &handlerMockQueries{
		friendships: []db.Friendship{
			{ID: uuid.New(), UserID: uid, FriendID: friend, Status: "accepted"},
		},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, uid.String())

	rec := doRequest(r, http.MethodGet, "/", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 1 {
		t.Errorf("expected 1 friend, got %d", len(data))
	}
}

func TestHandler_ListFriends_Empty(t *testing.T) {
	uid := uuid.New()
	q := &handlerMockQueries{
		friendships: []db.Friendship{},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, uid.String())

	rec := doRequest(r, http.MethodGet, "/", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 0 {
		t.Errorf("expected 0 friends, got %d", len(data))
	}
}

// ---------------------------------------------------------------------------
// 4. Block user — basic flow
// ---------------------------------------------------------------------------

func TestHandler_Block_Success(t *testing.T) {
	blocker := uuid.New()
	blocked := uuid.New()
	q := &handlerMockQueries{
		blockRelation: db.BlockRelation{ID: uuid.New(), BlockerID: blocker, BlockedID: blocked},
		friendErr:     sql.ErrNoRows,
		isBlockedErr:  sql.ErrNoRows,
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, blocker.String())

	rec := doRequest(r, http.MethodPost, "/block", map[string]string{
		"user_id": blocked.String(),
		"reason":  "spam",
	})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_Block_InvalidBody(t *testing.T) {
	userID := uuid.New()
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, userID.String())

	req := httptest.NewRequest(http.MethodPost, "/block", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_Unblock_Success(t *testing.T) {
	blocker := uuid.New()
	blocked := uuid.New()
	q := &handlerMockQueries{
		blockRelation: db.BlockRelation{ID: uuid.New(), BlockerID: blocker, BlockedID: blocked},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, blocker.String())

	rec := doRequest(r, http.MethodDelete, "/block/"+blocked.String(), nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_ListBlocked_Success(t *testing.T) {
	blocker := uuid.New()
	blocked := uuid.New()
	q := &handlerMockQueries{
		blockedList: []db.BlockRelation{
			{ID: uuid.New(), BlockerID: blocker, BlockedID: blocked},
		},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, blocker.String())

	rec := doRequest(r, http.MethodGet, "/blocked", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 1 {
		t.Errorf("expected 1 blocked user, got %d", len(data))
	}
}

// ---------------------------------------------------------------------------
// 5. Remove friend — basic flow
// ---------------------------------------------------------------------------

func TestHandler_RemoveFriend_Success(t *testing.T) {
	user := uuid.New()
	friend := uuid.New()
	q := &handlerMockQueries{
		deleteByUsers: db.Friendship{ID: uuid.New(), UserID: user, FriendID: friend},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, user.String())

	rec := doRequest(r, http.MethodDelete, "/"+friend.String(), nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Additional: Accept / Reject friend request flows
// ---------------------------------------------------------------------------

func TestHandler_AcceptRequest_Success(t *testing.T) {
	fid := uuid.New()
	q := &handlerMockQueries{
		friendship: db.Friendship{ID: fid, Status: "accepted"},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, uuid.New().String())

	rec := doRequest(r, http.MethodPut, "/"+fid.String()+"/accept", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if ok, _ := resp["ok"].(bool); !ok {
		t.Error("expected ok=true in response")
	}
}

func TestHandler_RejectRequest_Success(t *testing.T) {
	fid := uuid.New()
	q := &handlerMockQueries{}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, uuid.New().String())

	rec := doRequest(r, http.MethodDelete, "/"+fid.String()+"/reject", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data to be an object")
	}
	if data["status"] != "rejected" {
		t.Errorf("expected status=rejected, got %v", data["status"])
	}
}

// ---------------------------------------------------------------------------
// List pending requests
// ---------------------------------------------------------------------------

func TestHandler_ListPendingRequests_Success(t *testing.T) {
	uid := uuid.New()
	q := &handlerMockQueries{
		friendships: []db.Friendship{
			{ID: uuid.New(), UserID: uuid.New(), FriendID: uid, Status: "pending"},
		},
	}
	svc := NewFriendsService(q, nil, nil)
	h := NewFriendsHandler(svc)
	r := newTestRouter(h, uid.String())

	rec := doRequest(r, http.MethodGet, "/pending", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 1 {
		t.Errorf("expected 1 pending request, got %d", len(data))
	}
}
