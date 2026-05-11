package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisAdapter struct {
	client *redis.Client
}

func NewRedisAdapter(client *redis.Client) *RedisAdapter {
	return &RedisAdapter{client: client}
}

func (a *RedisAdapter) Set(ctx context.Context, key string, value any, ttl time.Duration) *redis.StatusCmd {
	return a.client.Set(ctx, key, value, ttl)
}

func (a *RedisAdapter) Get(ctx context.Context, key string) *redis.StringCmd {
	return a.client.Get(ctx, key)
}

func (a *RedisAdapter) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return a.client.Del(ctx, keys...)
}
