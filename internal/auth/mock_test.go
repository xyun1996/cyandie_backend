package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type mockQueries struct {
	users       map[string]db.User
	credentials map[string]db.Credential
	passwords   map[string]string // username -> plaintext password for testing
}

func newMockQueries() *mockQueries {
	return &mockQueries{
		users:       make(map[string]db.User),
		credentials: make(map[string]db.Credential),
		passwords:   make(map[string]string),
	}
}

func (m *mockQueries) GetUserByID(_ context.Context, id uuid.UUID) (db.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQueries) GetUserByUsername(_ context.Context, username string) (db.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQueries) CreateUser(_ context.Context, params db.CreateUserParams) (db.User, error) {
	if _, exists := m.users[params.Username]; exists {
		return db.User{}, fmt.Errorf("duplicate username")
	}
	user := db.User{
		ID:          uuid.New(),
		Username:    params.Username,
		Email:       params.Email,
		DisplayName: params.DisplayName,
		Status:      "active",
		Metadata:    pqtype.NullRawMessage{Valid: true, RawMessage: []byte(`{}`)},
	}
	m.users[params.Username] = user
	return user, nil
}

func (m *mockQueries) GetCredentialByTypeIdentifier(_ context.Context, params db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	key := params.Type + ":" + params.Identifier
	if c, ok := m.credentials[key]; ok {
		return c, nil
	}
	return db.Credential{}, sql.ErrNoRows
}

func (m *mockQueries) CreateCredential(_ context.Context, params db.CreateCredentialParams) (db.Credential, error) {
	key := params.Type + ":" + params.Identifier
	cred := db.Credential{
		ID:         uuid.New(),
		UserID:     params.UserID,
		Type:       params.Type,
		Identifier: params.Identifier,
		SecretHash: params.SecretHash,
		Verified:   params.Verified,
	}
	m.credentials[key] = cred
	return cred, nil
}

func (m *mockQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *mockQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *mockQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *mockQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (m *mockQueries) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockQueries) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (m *mockQueries) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (m *mockQueries) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (m *mockQueries) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (m *mockQueries) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (m *mockQueries) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockQueries) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *mockQueries) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockQueries) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (m *mockQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockQueries) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *mockQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *mockQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *mockQueries) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) {
	return nil, nil
}
func (m *mockQueries) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *mockQueries) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (m *mockQueries) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (m *mockQueries) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (m *mockQueries) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockQueries) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockQueries) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, sql.ErrNoRows
}
func (m *mockQueries) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockQueries) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockQueries) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return nil, nil
}
func (m *mockQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return uuid.UUID{}, sql.ErrNoRows
}
func (m *mockQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
