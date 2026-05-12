package leaderboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	err := svc.SubmitScore(context.Background(), "test-board", "user-1", 100.0)
	if err != nil {
		t.Fatalf("SubmitScore failed: %v", err)
	}
}

func TestLeaderboardService_SubmitScore_BoardNotFound(t *testing.T) {
	svc := newTestLBService()
	err := svc.SubmitScore(context.Background(), "nonexistent", "user-1", 100.0)
	if err == nil {
		t.Error("expected error for nonexistent board")
	}
}

func TestLeaderboardService_GetRanking(t *testing.T) {
	svc := newTestLBService()
	svc.SubmitScore(context.Background(), "test-board", "user-1", 100.0)
	svc.SubmitScore(context.Background(), "test-board", "user-2", 200.0)

	entries, err := svc.GetRanking(context.Background(), "test-board", 10, 0)
	if err != nil {
		t.Fatalf("GetRanking failed: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected ranking entries")
	}
}
