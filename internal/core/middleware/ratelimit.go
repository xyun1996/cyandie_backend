package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	Limit  int    `yaml:"limit"`
	Window string `yaml:"window"`
}

type RedisLimiterClient interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, ttl time.Duration) *redis.BoolCmd
}

type RateLimiter struct {
	client RedisLimiterClient
	config RateLimitConfig
}

func NewRateLimiter(client RedisLimiterClient, config RateLimitConfig) *RateLimiter {
	return &RateLimiter{client: client, config: config}
}

func (rl *RateLimiter) Middleware(group string) func(http.Handler) http.Handler {
	windowDuration := rl.parseWindow()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity := rl.resolveIdentity(r)
			key := fmt.Sprintf("ratelimit:%s:%s:%d", identity, group, time.Now().Truncate(windowDuration).Unix())

			count, err := rl.client.Incr(r.Context(), key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rl.client.Expire(r.Context(), key, windowDuration*2)
			}

			limit := rl.config.Limit
			remaining := limit - int(count)
			if remaining < 0 {
				remaining = 0
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if int(count) > limit {
				w.Header().Set("Retry-After", strconv.Itoa(int(windowDuration.Seconds())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(fmt.Sprintf(`{"ok":false,"error":{"code":"RATE_LIMITED","message":"rate limit exceeded, retry after %ds","requestId":"%s"}}`, int(windowDuration.Seconds()), r.Header.Get("X-Request-ID"))))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) resolveIdentity(r *http.Request) string {
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return "user:" + uid
	}
	return "ip:" + r.RemoteAddr
}

func (rl *RateLimiter) parseWindow() time.Duration {
	d, err := time.ParseDuration(rl.config.Window)
	if err != nil {
		return time.Minute
	}
	return d
}
