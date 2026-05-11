package cache

import (
	"context"
	"fmt"

	"github.com/cyandie/backend/internal/core/config"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	*redis.Client
}

func New(cfg config.CacheConfig) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Cache{client}, nil
}

func (c *Cache) Close() error {
	return c.Client.Close()
}
