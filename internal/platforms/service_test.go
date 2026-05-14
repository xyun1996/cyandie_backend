package platforms

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type platformMockQueries struct {
	bindings map[string]db.PlatformBinding
	users    map[string]db.User
}

func newPlatformMockQueries() *platformMockQueries {
	return &platformMockQueries{
		bindings: make(map[string]db.PlatformBinding),
		users:    make(map[string]db.User),
	}
}

func (m *platformMockQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *platformMockQueries) CreatePlatformBinding(_ context.Context, params db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	b := db.PlatformBinding{
		ID:             uuid.New(),
		UserID:         params.UserID,
		Platform:       params.Platform,
		PlatformUserID: params.PlatformUserID,
		Metadata:       pqtype.NullRawMessage{RawMessage: []byte(`{}`), Valid: true},
	}
	m.bindings[params.Platform+":"+params.PlatformUserID] = b
	return b, nil
}
func (m *platformMockQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *platformMockQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) {
	return 1, nil
}
func (m *platformMockQueries) CreateUser(_ context.Context, params db.CreateUserParams) (db.User, error) {
	u := db.User{ID: uuid.New(), Username: params.Username, Status: "active", Metadata: pqtype.NullRawMessage{RawMessage: []byte(`{}`), Valid: true}}
	m.users[params.Username] = u
	return u, nil
}
func (m *platformMockQueries) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *platformMockQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *platformMockQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *platformMockQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *platformMockQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *platformMockQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *platformMockQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *platformMockQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *platformMockQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *platformMockQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *platformMockQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (m *platformMockQueries) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *platformMockQueries) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (m *platformMockQueries) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (m *platformMockQueries) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (m *platformMockQueries) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (m *platformMockQueries) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *platformMockQueries) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *platformMockQueries) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) {
	return nil, nil
}
func (m *platformMockQueries) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (m *platformMockQueries) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *platformMockQueries) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *platformMockQueries) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (m *platformMockQueries) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (m *platformMockQueries) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (m *platformMockQueries) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *platformMockQueries) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *platformMockQueries) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetFriendshipByUsers(_ context.Context, _ db.GetFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *platformMockQueries) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *platformMockQueries) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *platformMockQueries) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *platformMockQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *platformMockQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *platformMockQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return nil, nil
}
func (m *platformMockQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return uuid.UUID{}, sql.ErrNoRows
}
func (m *platformMockQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *platformMockQueries) ListRoomsByUser(_ context.Context, _ uuid.UUID) ([]db.ChatRoom, error) { return nil, nil }
func (m *platformMockQueries) ListAllBlockRelations(_ context.Context) ([]db.BlockRelation, error) {
	return nil, nil
}

func newTestPlatformService() *PlatformService {
	reg := NewPlatformRegistry()
	reg.RegisterOAuth(&mockProvider{name: "wechat"})
	return NewPlatformService(newPlatformMockQueries(), reg)
}

func TestPlatformService_GetAuthURL(t *testing.T) {
	svc := newTestPlatformService()
	url, err := svc.GetAuthURL("wechat", "state123")
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty auth URL")
	}
}

func TestPlatformService_GetAuthURL_UnsupportedPlatform(t *testing.T) {
	svc := newTestPlatformService()
	_, err := svc.GetAuthURL("google", "state123")
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
}

func TestPlatformService_UnbindPlatform_NotBound(t *testing.T) {
	svc := newTestPlatformService()
	err := svc.UnbindPlatform(context.Background(), uuid.New().String(), "wechat")
	if err == nil {
		t.Error("expected error for unbound platform")
	}
}
