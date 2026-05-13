package users

import (
	"context"
	"database/sql"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type mockUserQueries struct {
	user db.User
}

func (m *mockUserQueries) GetUserByID(_ context.Context, id uuid.UUID) (db.User, error) {
	m.user.ID = id
	m.user.Username = "testuser"
	m.user.Status = "active"
	m.user.Metadata = pqtype.NullRawMessage{Valid: true, RawMessage: []byte(`{}`)}
	return m.user, nil
}
func (m *mockUserQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error) {
	return m.user, nil
}
func (m *mockUserQueries) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockUserQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *mockUserQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockUserQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockUserQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *mockUserQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *mockUserQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *mockUserQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockUserQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *mockUserQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockUserQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (m *mockUserQueries) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockUserQueries) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (m *mockUserQueries) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (m *mockUserQueries) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (m *mockUserQueries) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (m *mockUserQueries) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (m *mockUserQueries) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockUserQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockUserQueries) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *mockUserQueries) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockUserQueries) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (m *mockUserQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockUserQueries) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *mockUserQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *mockUserQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *mockUserQueries) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) {
	return nil, nil
}
func (m *mockUserQueries) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *mockUserQueries) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (m *mockUserQueries) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (m *mockUserQueries) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (m *mockUserQueries) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockUserQueries) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockUserQueries) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, sql.ErrNoRows
}
func (m *mockUserQueries) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockUserQueries) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockUserQueries) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockUserQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockUserQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockUserQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return nil, nil
}
func (m *mockUserQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return uuid.UUID{}, sql.ErrNoRows
}
func (m *mockUserQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
