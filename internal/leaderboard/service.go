package leaderboard

import (
	"context"
	"fmt"

	"github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type LeaderboardService struct {
	queries db.Querier
	redis   redisClient
}

type redisClient interface {
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRevRank(ctx context.Context, key string, member string) *redis.IntCmd
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd
	ZScore(ctx context.Context, key string, member string) *redis.FloatCmd
	ZCard(ctx context.Context, key string) *redis.IntCmd
}

func NewLeaderboardService(queries db.Querier, redis redisClient) *LeaderboardService {
	return &LeaderboardService{queries: queries, redis: redis}
}

func (s *LeaderboardService) SubmitScore(ctx context.Context, boardCode string, userID string, score float64) error {
	config, err := s.queries.GetLeaderboardConfig(ctx, boardCode)
	if err != nil {
		return errors.New(errors.ErrNotFound, "leaderboard not found")
	}

	uid, _ := uuid.Parse(userID)
	_, err = s.queries.CreateScore(ctx, db.CreateScoreParams{
		BoardID: config.ID,
		UserID:  uid,
		Score:   score,
	})
	if err != nil {
		return fmt.Errorf("create score: %w", err)
	}

	key := "leaderboard:" + config.Code
	s.redis.ZAdd(ctx, key, redis.Z{Score: score, Member: userID})

	return nil
}

func (s *LeaderboardService) GetRanking(ctx context.Context, boardCode string, limit, offset int) ([]RankEntry, error) {
	_, err := s.queries.GetLeaderboardConfig(ctx, boardCode)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "leaderboard not found")
	}

	key := "leaderboard:" + boardCode
	results, err := s.redis.ZRevRangeWithScores(ctx, key, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("redis zrevrange: %w", err)
	}

	entries := make([]RankEntry, len(results))
	for i, z := range results {
		rank := int64(offset) + int64(i) + 1
		entries[i] = RankEntry{
			Rank:    rank,
			UserID:  fmt.Sprintf("%v", z.Member),
			Score:   z.Score,
		}
	}
	return entries, nil
}

func (s *LeaderboardService) GetUserRank(ctx context.Context, boardCode string, userID string) (*RankEntry, error) {
	_, err := s.queries.GetLeaderboardConfig(ctx, boardCode)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "leaderboard not found")
	}

	key := "leaderboard:" + boardCode
	rank, err := s.redis.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "user not ranked")
	}

	score, err := s.redis.ZScore(ctx, key, userID).Result()
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "user not ranked")
	}

	return &RankEntry{
		Rank:   rank + 1,
		UserID: userID,
		Score:  score,
	}, nil
}

type RankEntry struct {
	Rank    int64   `json:"rank"`
	UserID  string  `json:"userId"`
	Score   float64 `json:"score"`
}
