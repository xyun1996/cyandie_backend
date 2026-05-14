package leaderboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sqlc-dev/pqtype"
)

type mockLBRedis struct {
	scores map[string]float64
}

func (m *mockLBRedis) ZAdd(_ context.Context, _ string, members ...redis.Z) *redis.IntCmd {
	for _, z := range members {
		m.scores[fmt.Sprintf("%v", z.Member)] = z.Score
	}
	return redis.NewIntCmd(context.Background())
}
func (m *mockLBRedis) ZRevRank(_ context.Context, _ string, member string) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	if _, ok := m.scores[member]; ok {
		cmd.SetVal(0)
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}
func (m *mockLBRedis) ZRevRangeWithScores(_ context.Context, _ string, _, _ int64) *redis.ZSliceCmd {
	cmd := redis.NewZSliceCmd(context.Background())
	var zs []redis.Z
	for member, score := range m.scores {
		zs = append(zs, redis.Z{Score: score, Member: member})
	}
	cmd.SetVal(zs)
	return cmd
}
func (m *mockLBRedis) ZScore(_ context.Context, _ string, member string) *redis.FloatCmd {
	cmd := redis.NewFloatCmd(context.Background())
	if score, ok := m.scores[member]; ok {
		cmd.SetVal(score)
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}
func (m *mockLBRedis) ZCard(_ context.Context, _ string) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	cmd.SetVal(int64(len(m.scores)))
	return cmd
}

type mockLBQueries struct {
	config db.LeaderboardConfig
}

func (m *mockLBQueries) GetLeaderboardConfig(_ context.Context, code string) (db.LeaderboardConfig, error) {
	if m.config.Code == code {
		return m.config, nil
	}
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *mockLBQueries) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) {
	return nil, nil
}
func (m *mockLBQueries) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *mockLBQueries) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
// Stub remaining Querier methods
func (m *mockLBQueries) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error) { return db.User{}, sql.ErrNoRows }
func (m *mockLBQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error) { return db.User{}, sql.ErrNoRows }
func (m *mockLBQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) { return nil, nil }
func (m *mockLBQueries) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) { return db.User{}, nil }
func (m *mockLBQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) { return db.User{}, nil }
func (m *mockLBQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) { return db.User{}, nil }
func (m *mockLBQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) { return db.Credential{}, sql.ErrNoRows }
func (m *mockLBQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) { return db.Credential{}, nil }
func (m *mockLBQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) { return nil, nil }
func (m *mockLBQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) { return db.UserSession{}, nil }
func (m *mockLBQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) { return db.UserSession{}, sql.ErrNoRows }
func (m *mockLBQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) { return db.UserSession{}, nil }
func (m *mockLBQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) { return nil, nil }
func (m *mockLBQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) { return db.PlatformBinding{}, sql.ErrNoRows }
func (m *mockLBQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) { return nil, nil }
func (m *mockLBQueries) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) { return db.PlatformBinding{}, nil }
func (m *mockLBQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) { return db.PlatformBinding{}, nil }
func (m *mockLBQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) { return 0, nil }
func (m *mockLBQueries) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockLBQueries) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (m *mockLBQueries) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (m *mockLBQueries) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (m *mockLBQueries) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (m *mockLBQueries) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (m *mockLBQueries) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockLBQueries) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *mockLBQueries) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (m *mockLBQueries) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (m *mockLBQueries) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (m *mockLBQueries) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockLBQueries) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockLBQueries) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, sql.ErrNoRows
}
func (m *mockLBQueries) GetFriendshipByUsers(_ context.Context, _ db.GetFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockLBQueries) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockLBQueries) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockLBQueries) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockLBQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockLBQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockLBQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return nil, nil
}
func (m *mockLBQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return uuid.UUID{}, sql.ErrNoRows
}
func (m *mockLBQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockLBQueries) ListRoomsByUser(_ context.Context, _ uuid.UUID) ([]db.ChatRoom, error) { return nil, nil }
func (m *mockLBQueries) ListAllBlockRelations(_ context.Context) ([]db.BlockRelation, error) {
	return nil, nil
}

func newTestLBService() *LeaderboardService {
	q := &mockLBQueries{config: db.LeaderboardConfig{
		ID: uuid.New(), Code: "test-board", Name: "Test Board",
		UpdateStrategy: "highest", MaxEntries: sql.NullInt32{Int32: 100, Valid: true}, Metadata: pqtype.NullRawMessage{RawMessage: json.RawMessage(`{}`), Valid: true},
	}}
	r := &mockLBRedis{scores: make(map[string]float64)}
	return NewLeaderboardService(q, r)
}

func TestLeaderboardService_SubmitScore(t *testing.T) {
	svc := newTestLBService()
	err := svc.SubmitScore(context.Background(), "test-board", uuid.New().String(), 100.0)
	if err != nil {
		t.Fatalf("SubmitScore failed: %v", err)
	}
}

func TestLeaderboardService_SubmitScore_BoardNotFound(t *testing.T) {
	svc := newTestLBService()
	err := svc.SubmitScore(context.Background(), "nonexistent", uuid.New().String(), 100.0)
	if err == nil {
		t.Error("expected error for nonexistent board")
	}
}

func TestLeaderboardService_SubmitScore_InvalidScore(t *testing.T) {
	svc := newTestLBService()
	tests := []struct {
		name  string
		score float64
	}{
		{"NaN", math.NaN()},
		{"positive Inf", math.Inf(1)},
		{"negative Inf", math.Inf(-1)},
		{"negative", -1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.SubmitScore(context.Background(), "test-board", uuid.New().String(), tt.score)
			if err == nil {
				t.Error("expected error for invalid score")
			}
		})
	}
}

func TestLeaderboardService_SubmitScore_InvalidUserID(t *testing.T) {
	svc := newTestLBService()
	err := svc.SubmitScore(context.Background(), "test-board", "not-a-uuid", 100.0)
	if err == nil {
		t.Error("expected error for invalid user ID")
	}
}

func TestLeaderboardService_GetRanking(t *testing.T) {
	svc := newTestLBService()
	uid1 := uuid.New().String()
	uid2 := uuid.New().String()
	svc.SubmitScore(context.Background(), "test-board", uid1, 100.0)
	svc.SubmitScore(context.Background(), "test-board", uid2, 200.0)

	entries, err := svc.GetRanking(context.Background(), "test-board", 10, 0)
	if err != nil {
		t.Fatalf("GetRanking failed: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected ranking entries")
	}
}
