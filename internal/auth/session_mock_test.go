package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type mockRedisClient struct {
	store map[string]string
}

func (m *mockRedisClient) Set(_ context.Context, key string, value any, _ time.Duration) *redis.StatusCmd {
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

func (m *mockRedisClient) Get(_ context.Context, key string) *redis.StringCmd {
	val, ok := m.store[key]
	cmd := redis.NewStringCmd(context.Background())
	if ok {
		cmd.SetVal(val)
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}

func (m *mockRedisClient) Del(_ context.Context, keys ...string) *redis.IntCmd {
	for _, k := range keys {
		delete(m.store, k)
	}
	return redis.NewIntCmd(context.Background())
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{store: make(map[string]string)}
}
