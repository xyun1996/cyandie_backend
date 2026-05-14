package admin

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockAdminQueries struct {
	admin     db.AdminUser
	adminErr  error
	users     []db.User
	usersErr  error
	status    db.User
	statusErr error
	logs      []db.AuditLog
	logsErr   error

	// auditLogParams captures the last CreateAuditLog call's parameters.
	auditLogParams db.CreateAuditLogParams
}

type mockAuthService struct{}

func (mockAuthService) GenerateToken(_ string) (string, error) { return "mock-token", nil }

func (m *mockAdminQueries) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *mockAdminQueries) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return m.admin, m.adminErr
}
func (m *mockAdminQueries) CreateAuditLog(_ context.Context, params db.CreateAuditLogParams) (db.AuditLog, error) {
	m.auditLogParams = params
	return db.AuditLog{}, nil
}
func (m *mockAdminQueries) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return m.logs, m.logsErr
}
func (m *mockAdminQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return m.users, m.usersErr
}
func (m *mockAdminQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return m.status, m.statusErr
}

// Stub remaining Querier methods
func (m *mockAdminQueries) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) { return db.User{}, nil }
func (m *mockAdminQueries) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error)          { return db.User{}, sql.ErrNoRows }
func (m *mockAdminQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error)        { return db.User{}, sql.ErrNoRows }
func (m *mockAdminQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockAdminQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *mockAdminQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *mockAdminQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *mockAdminQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) { return 0, nil }
func (m *mockAdminQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockAdminQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *mockAdminQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockAdminQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (m *mockAdminQueries) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockAdminQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *mockAdminQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *mockAdminQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockAdminQueries) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *mockAdminQueries) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *mockAdminQueries) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) { return nil, nil }
func (m *mockAdminQueries) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (m *mockAdminQueries) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockAdminQueries) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (m *mockAdminQueries) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (m *mockAdminQueries) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (m *mockAdminQueries) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (m *mockAdminQueries) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (m *mockAdminQueries) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockAdminQueries) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockAdminQueries) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockAdminQueries) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, sql.ErrNoRows
}
func (m *mockAdminQueries) GetFriendshipByUsers(_ context.Context, _ db.GetFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockAdminQueries) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) { return nil, nil }
func (m *mockAdminQueries) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockAdminQueries) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockAdminQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockAdminQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockAdminQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return nil, nil
}
func (m *mockAdminQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return uuid.UUID{}, sql.ErrNoRows
}
func (m *mockAdminQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockAdminQueries) ListRoomsByUser(_ context.Context, _ uuid.UUID) ([]db.ChatRoom, error) { return nil, nil }

func TestAdminService_Login_Success(t *testing.T) {
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

	admin, err := svc.Login(context.Background(), AdminLoginRequest{Username: "admin1", Password: "pass123"})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if admin.Username != "admin1" {
		t.Errorf("expected username admin1, got %s", admin.Username)
	}
}

func TestAdminService_Login_InvalidPassword(t *testing.T) {
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

	_, err := svc.Login(context.Background(), AdminLoginRequest{Username: "admin1", Password: "wrong"})
	if err == nil {
		t.Error("expected error for invalid password")
	}
}

func TestAdminService_Login_NotFound(t *testing.T) {
	q := &mockAdminQueries{adminErr: sql.ErrNoRows}
	svc := NewAdminService(q, mockAuthService{})

	_, err := svc.Login(context.Background(), AdminLoginRequest{Username: "nobody", Password: "pass"})
	if err == nil {
		t.Error("expected error for nonexistent admin")
	}
}

func TestAdminService_Login_BannedAdmin(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	q := &mockAdminQueries{
		admin: db.AdminUser{
			ID:           uuid.New(),
			Username:     "bannedadmin",
			PasswordHash: string(hash),
			Role:         "admin",
			Status:       "banned",
		},
	}
	svc := NewAdminService(q, mockAuthService{})

	_, err := svc.Login(context.Background(), AdminLoginRequest{Username: "bannedadmin", Password: "pass123"})
	if err == nil {
		t.Error("expected error for banned admin")
	}
}

func TestAdminService_UpdateUserStatus(t *testing.T) {
	uid := uuid.New()
	q := &mockAdminQueries{
		status: db.User{ID: uid, Username: "user1", Status: "banned"},
	}
	svc := NewAdminService(q, mockAuthService{})

	user, err := svc.UpdateUserStatus(context.Background(), uid.String(), "banned")
	if err != nil {
		t.Fatalf("UpdateUserStatus failed: %v", err)
	}
	if user.Status != "banned" {
		t.Errorf("expected status banned, got %s", user.Status)
	}
}

func TestAdminService_UpdateUserStatus_InvalidStatus(t *testing.T) {
	q := &mockAdminQueries{}
	svc := NewAdminService(q, mockAuthService{})

	_, err := svc.UpdateUserStatus(context.Background(), uuid.New().String(), "invalid")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestAdminService_ListAuditLogs(t *testing.T) {
	q := &mockAdminQueries{
		logs: []db.AuditLog{
			{ID: uuid.New(), Action: "update_user_status", TargetType: "user"},
		},
	}
	svc := NewAdminService(q, mockAuthService{})

	logs, err := svc.ListAuditLogs(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}
