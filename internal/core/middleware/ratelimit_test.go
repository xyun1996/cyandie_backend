package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type mockRedisLimiter struct {
	mu    sync.Mutex
	store map[string]int64
}

func newMockRedisLimiter() *mockRedisLimiter {
	return &mockRedisLimiter{store: make(map[string]int64)}
}

func (m *mockRedisLimiter) Incr(_ context.Context, key string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key]++
	cmd := redis.NewIntCmd(context.Background())
	cmd.SetVal(m.store[key])
	return cmd
}

func (m *mockRedisLimiter) Expire(_ context.Context, _ string, _ time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(context.Background())
	cmd.SetVal(true)
	return cmd
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	limiter := NewRateLimiter(newMockRedisLimiter(), RateLimitConfig{
		Limit:  5,
		Window: "1m",
	})

	handler := limiter.Middleware("test-group")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	limiter := NewRateLimiter(newMockRedisLimiter(), RateLimitConfig{
		Limit:  3,
		Window: "1m",
	})

	handler := limiter.Middleware("test-group")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

func TestRateLimit_SetsHeaders(t *testing.T) {
	limiter := NewRateLimiter(newMockRedisLimiter(), RateLimitConfig{
		Limit:  10,
		Window: "1m",
	})

	handler := limiter.Middleware("test-group")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("expected X-RateLimit-Limit header")
	}
}
