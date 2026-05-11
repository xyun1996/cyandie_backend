package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Session struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	RefreshTokenHash string `json:"refresh_token_hash"`
	CreatedAt        int64  `json:"created_at"`
}

type RedisClient interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type SessionStore struct {
	client RedisClient
	prefix string
}

func NewSessionStore(client RedisClient) *SessionStore {
	return &SessionStore{
		client: client,
		prefix: "session:",
	}
}

func (s *SessionStore) Create(userID, refreshTokenHash string, ttlSeconds int) (*Session, error) {
	session := &Session{
		ID:               generateSessionID(),
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		CreatedAt:        time.Now().Unix(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("marshal session: %w", err)
	}

	key := s.prefix + session.ID
	err = s.client.Set(context.Background(), key, data, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return nil, fmt.Errorf("redis set: %w", err)
	}

	return session, nil
}

func (s *SessionStore) Get(sessionID string) (*Session, error) {
	key := s.prefix + sessionID
	data, err := s.client.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &session, nil
}

func (s *SessionStore) Revoke(sessionID string) error {
	key := s.prefix + sessionID
	return s.client.Del(context.Background(), key).Err()
}

func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
