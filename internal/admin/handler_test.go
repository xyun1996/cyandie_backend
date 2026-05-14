package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cyandie/backend/internal/auth"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Test helpers: mock Redis client + test AuthService for AuthGuard
// ---------------------------------------------------------------------------

type testRedisClient struct {
	store map[string]string
}

func newTestRedisClient() *testRedisClient {
	return &testRedisClient{store: make(map[string]string)}
}

func (m *testRedisClient) Set(_ context.Context, key string, value any, _ time.Duration) *redis.StatusCmd {
	switch v := value.(type) {
	case string:
		m.store[key] = v
	case []byte:
		m.store[key] = string(v)
	default:
		m.store[key] = fmt.Sprintf("%v", v)
	}
	return redis.NewStatusCmd(context.Background())
}

func (m *testRedisClient) Get(_ context.Context, key string) *redis.StringCmd {
	val, ok := m.store[key]
	cmd := redis.NewStringCmd(context.Background())
	if ok {
		cmd.SetVal(val)
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}

func (m *testRedisClient) Del(_ context.Context, keys ...string) *redis.IntCmd {
	for _, k := range keys {
		delete(m.store, k)
	}
	return redis.NewIntCmd(context.Background())
}

// newTestAuthService builds a real auth.AuthService with in-memory fakes
// so AuthGuard can sign and validate JWTs.
func newTestAuthService() *auth.AuthService {
	return auth.NewAuthService(auth.AuthServiceDeps{
		Queries:     &stubQueriesForAuth{},
		KeyManager:  auth.NewKeyManager([]auth.JWTKey{{KID: "k1", Secret: []byte("test-secret-key-1234567890123456")}}),
		Sessions:    auth.NewSessionStore(newTestRedisClient()),
		OTPNotifier: auth.LogNotifier{},
	})
}

// stubQueriesForAuth satisfies db.Querier with zero-value stubs.
// Only the methods needed by auth.AuthService are functional; the rest are no-ops.
type stubQueriesForAuth struct{}

func (s *stubQueriesForAuth) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *stubQueriesForAuth) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) GetUserByUsername(_ context.Context, _ string) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (s *stubQueriesForAuth) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (s *stubQueriesForAuth) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (s *stubQueriesForAuth) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (s *stubQueriesForAuth) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (s *stubQueriesForAuth) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (s *stubQueriesForAuth) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (s *stubQueriesForAuth) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (s *stubQueriesForAuth) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (s *stubQueriesForAuth) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (s *stubQueriesForAuth) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (s *stubQueriesForAuth) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (s *stubQueriesForAuth) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (s *stubQueriesForAuth) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (s *stubQueriesForAuth) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (s *stubQueriesForAuth) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) GetFriendshipByUsers(_ context.Context, _ db.GetFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (s *stubQueriesForAuth) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (s *stubQueriesForAuth) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (s *stubQueriesForAuth) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (s *stubQueriesForAuth) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return uuid.UUID{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (s *stubQueriesForAuth) ListRoomsByUser(_ context.Context, _ uuid.UUID) ([]db.ChatRoom, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (s *stubQueriesForAuth) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (s *stubQueriesForAuth) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (s *stubQueriesForAuth) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (s *stubQueriesForAuth) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}

// ---------------------------------------------------------------------------
// Router helper: wires routes exactly as module.go does
// ---------------------------------------------------------------------------

// newTestRouter builds a chi.Router with the same route layout as
// Module.RegisterRoutes so that AuthGuard is applied correctly.
func newTestRouter(handler *AdminHandler, authSvc *auth.AuthService) chi.Router {
	mux := chi.NewRouter()
	mux.Post("/login", handler.Login)
	mux.Group(func(r chi.Router) {
		r.Use(auth.AuthGuard(authSvc))
		r.Get("/users", handler.ListUsers)
		r.Put("/users/{id}/status", handler.UpdateUserStatus)
		r.Get("/audit-logs", handler.ListAuditLogs)
	})
	return mux
}

// generateTestToken creates a valid JWT using the test AuthService.
func generateTestToken(authSvc *auth.AuthService) string {
	token, err := authSvc.GenerateToken(uuid.New().String(), "admin")
	if err != nil {
		panic("failed to generate test token: " + err.Error())
	}
	return token
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAdminHandler_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	q := &mockAdminQueries{
		admin: db.AdminUser{
			ID:           uuid.New(),
			Username:     "admin1",
			PasswordHash: string(hash),
			Role:         "admin",
			Status:       "active",
		},
	}
	svc := NewAdminService(q, mockAuthService{})
	handler := NewAdminHandler(svc)
	router := newTestRouter(handler, newTestAuthService())

	body, _ := json.Marshal(AdminLoginRequest{Username: "admin1", Password: "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			ID          string `json:"id"`
			Username    string `json:"username"`
			Role        string `json:"role"`
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.Data.Username != "admin1" {
		t.Errorf("expected username admin1, got %s", resp.Data.Username)
	}
	if resp.Data.AccessToken != "mock-token" {
		t.Errorf("expected mock-token, got %s", resp.Data.AccessToken)
	}
}

func TestAdminHandler_Login_InvalidCredentials(t *testing.T) {
	q := &mockAdminQueries{adminErr: sql.ErrNoRows}
	svc := NewAdminService(q, mockAuthService{})
	handler := NewAdminHandler(svc)
	router := newTestRouter(handler, newTestAuthService())

	body, _ := json.Marshal(AdminLoginRequest{Username: "nobody", Password: "pass"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_Login_MissingFields(t *testing.T) {
	svc := NewAdminService(&mockAdminQueries{}, mockAuthService{})
	handler := NewAdminHandler(svc)
	router := newTestRouter(handler, newTestAuthService())

	// Send empty JSON object — both username and password are missing.
	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_ListUsers_RequiresAuth(t *testing.T) {
	svc := NewAdminService(&mockAdminQueries{}, mockAuthService{})
	handler := NewAdminHandler(svc)
	authSvc := newTestAuthService()
	router := newTestRouter(handler, authSvc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_ListUsers_WithAuth(t *testing.T) {
	q := &mockAdminQueries{
		users: []db.User{
			{ID: uuid.New(), Username: "user1", Status: "active"},
			{ID: uuid.New(), Username: "user2", Status: "active"},
		},
	}
	svc := NewAdminService(q, mockAuthService{})
	handler := NewAdminHandler(svc)
	authSvc := newTestAuthService()
	router := newTestRouter(handler, authSvc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(authSvc))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_UpdateUserStatus_RequiresAuth(t *testing.T) {
	svc := NewAdminService(&mockAdminQueries{}, mockAuthService{})
	handler := NewAdminHandler(svc)
	authSvc := newTestAuthService()
	router := newTestRouter(handler, authSvc)

	body, _ := json.Marshal(map[string]string{"status": "banned"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+uuid.New().String()+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_ListAuditLogs_RequiresAuth(t *testing.T) {
	svc := NewAdminService(&mockAdminQueries{}, mockAuthService{})
	handler := NewAdminHandler(svc)
	authSvc := newTestAuthService()
	router := newTestRouter(handler, authSvc)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_ListAuditLogs_WithAuth(t *testing.T) {
	q := &mockAdminQueries{
		logs: []db.AuditLog{
			{ID: uuid.New(), Action: "update_user_status", TargetType: "user"},
		},
	}
	svc := NewAdminService(q, mockAuthService{})
	handler := NewAdminHandler(svc)
	authSvc := newTestAuthService()
	router := newTestRouter(handler, authSvc)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestToken(authSvc))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminHandler_UpdateUserStatus_AuditLogOperatorID(t *testing.T) {
	operatorUUID := uuid.New()
	targetUUID := uuid.New()

	q := &mockAdminQueries{
		status: db.User{ID: targetUUID, Username: "user1", Status: "banned"},
	}
	svc := NewAdminService(q, mockAuthService{})
	handler := NewAdminHandler(svc)
	authSvc := newTestAuthService()
	router := newTestRouter(handler, authSvc)

	// Generate a token with a known userID so we can verify it shows up as operatorID.
	token, err := authSvc.GenerateToken(operatorUUID.String(), "admin")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"status": "banned"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+targetUUID.String()+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	// Verify that the audit log received the correct operatorID from the JWT context.
	if !q.auditLogParams.OperatorID.Valid {
		t.Fatal("expected OperatorID to be valid in audit log, but it was zero/invalid")
	}
	if q.auditLogParams.OperatorID.UUID != operatorUUID {
		t.Errorf("expected operatorID %s, got %s", operatorUUID, q.auditLogParams.OperatorID.UUID)
	}
	if q.auditLogParams.Action != "update_user_status" {
		t.Errorf("expected action update_user_status, got %s", q.auditLogParams.Action)
	}
	if q.auditLogParams.TargetType != "user" {
		t.Errorf("expected target type user, got %s", q.auditLogParams.TargetType)
	}
	if q.auditLogParams.TargetID != targetUUID.String() {
		t.Errorf("expected targetID %s, got %s", targetUUID.String(), q.auditLogParams.TargetID)
	}
}
