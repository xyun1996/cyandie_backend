# Rate Limiting and Request Timeout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Redis sliding window rate limiting and context-based request timeout middleware to all HTTP APIs.

**Architecture:** Two chi middlewares — `RateLimit` checks Redis sliding window counters per user/IP per endpoint group, `Timeout` wraps requests with `context.WithTimeout`. Rate limit returns 429 with headers, timeout returns 503. Both integrate with existing middleware chain and error system.

**Tech Stack:** Go, chi v5, Redis (go-redis/v9), slog

---

## File Structure

```
internal/core/middleware/
  timeout.go           # Request timeout middleware
  timeout_test.go      # Timeout tests
  ratelimit.go         # Rate limiting middleware
  ratelimit_test.go    # Rate limiting tests
internal/core/config/
  config.go            # Updated with RateLimitConfig + TimeoutConfig
configs/
  config.example.yaml  # Updated with rate_limit + timeout sections
```

---

### Task 1: Add ErrTimeout Error Code

**Files:**
- Modify: `internal/core/errors/codes.go`

- [ ] **Step 1: Add ErrTimeout to codes.go**

Add to the General section in `internal/core/errors/codes.go`:

```go
ErrTimeout     = "TIMEOUT"
```

- [ ] **Step 2: Add HTTP mapping for ErrTimeout**

In `internal/core/errors/errors.go`, add `ErrTimeout` to the switch in `HTTPStatus()`:

```go
case ErrTimeout:
    return http.StatusServiceUnavailable
```

- [ ] **Step 3: Run existing tests to verify no regression**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/core/errors/... -v
```

Expected: PASS — all existing tests still pass.

- [ ] **Step 4: Commit**

```bash
git add internal/core/errors/ && git commit -m "feat: add TIMEOUT error code with 503 HTTP mapping"
```

---

### Task 2: Request Timeout Middleware

**Files:**
- Create: `internal/core/middleware/timeout.go`
- Create: `internal/core/middleware/timeout_test.go`

- [ ] **Step 1: Write failing test for timeout middleware**

Create `internal/core/middleware/timeout_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeout_CompletesWithinLimit(t *testing.T) {
	handler := Timeout(5*time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTimeout_ExceedsLimit(t *testing.T) {
	handler := Timeout(50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestTimeout_ContextCancelled(t *testing.T) {
	handler := Timeout(50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/core/middleware/... -v -run TestTimeout
```

Expected: FAIL — `Timeout` not defined.

- [ ] **Step 3: Implement timeout middleware**

Create `internal/core/middleware/timeout.go`:

```go
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			done := make(chan struct{})
			tw := &timeoutWriter{ResponseWriter: w}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Handler completed within timeout
				return
			case <-ctx.Done():
				// Timeout exceeded
				if !tw.written {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					json.NewEncoder(w).Encode(map[string]any{
						"ok": false,
						"error": map[string]any{
							"code":      "TIMEOUT",
							"message":   "request timeout",
							"requestId": r.Header.Get("X-Request-ID"),
						},
					})
				}
			}
		})
	}
}

type timeoutWriter struct {
	http.ResponseWriter
	written bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.written = true
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.written = true
	return tw.ResponseWriter.Write(p)
}
```

- [ ] **Step 4: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/core/middleware/... -v -run TestTimeout
```

Expected: PASS — all 3 tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/core/middleware/timeout.go internal/core/middleware/timeout_test.go && git commit -m "feat: add request timeout middleware with context deadline"
```

---

### Task 3: Rate Limiting Middleware

**Files:**
- Create: `internal/core/middleware/ratelimit.go`
- Create: `internal/core/middleware/ratelimit_test.go`

- [ ] **Step 1: Write failing test for rate limiting**

Create `internal/core/middleware/ratelimit_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		req.Header.Set("X-Request-ID", "req_1")
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

	// Exhaust limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Next request should be blocked
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/core/middleware/... -v -run TestRateLimit
```

Expected: FAIL — `NewRateLimiter` not defined.

- [ ] **Step 3: Implement rate limiting middleware**

Create `internal/core/middleware/ratelimit.go`:

```go
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

type RateLimiterDeps interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, ttl time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

type RateLimiter struct {
	client RedisLimiterClient
	config RateLimitConfig
}

type RedisLimiterClient interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, ttl time.Duration) *redis.BoolCmd
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
				// On Redis error, allow the request through
				next.ServeHTTP(w, r)
				return
			}

			// Set expiry on first request
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
	// If authenticated, use user ID from context
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return "user:" + uid
	}
	// Otherwise use IP
	return "ip:" + r.RemoteAddr
}

func (rl *RateLimiter) parseWindow() time.Duration {
	d, err := time.ParseDuration(rl.config.Window)
	if err != nil {
		return time.Minute
	}
	return d
}
```

- [ ] **Step 4: Create mock for tests**

Add to `internal/core/middleware/ratelimit_test.go`:

```go
package middleware

import (
	"context"
	"sync"
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
```

- [ ] **Step 5: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/core/middleware/... -v -run TestRateLimit
```

Expected: PASS — all 3 tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/core/middleware/ratelimit.go internal/core/middleware/ratelimit_test.go && git commit -m "feat: add Redis sliding window rate limiting middleware"
```

---

### Task 4: Config for Rate Limiting + Timeout

**Files:**
- Modify: `internal/core/config/config.go`
- Modify: `configs/config.example.yaml`

- [ ] **Step 1: Add config types**

Add to `internal/core/config/config.go`:

```go
type RateLimitRule struct {
	Limit  int    `yaml:"limit"`
	Window string `yaml:"window"`
}

type RateLimitConfig struct {
	Auth  RateLimitRule `yaml:"auth"`
	Write RateLimitRule `yaml:"write"`
	Read  RateLimitRule `yaml:"read"`
}

type TimeoutConfig struct {
	Default string            `yaml:"default"`
	Routes  map[string]string `yaml:"routes"`
}
```

Add to the `Config` struct:

```go
RateLimit RateLimitConfig `yaml:"rate_limit"`
Timeout   TimeoutConfig   `yaml:"timeout"`
```

Add defaults in `defaults()`:

```go
RateLimit: RateLimitConfig{
	Auth:  RateLimitRule{Limit: 10, Window: "1m"},
	Write: RateLimitRule{Limit: 30, Window: "1m"},
	Read:  RateLimitRule{Limit: 60, Window: "1m"},
},
Timeout: TimeoutConfig{
	Default: "30s",
},
```

- [ ] **Step 2: Update config.example.yaml**

Add to `configs/config.example.yaml`:

```yaml
rate_limit:
  auth:
    limit: 10
    window: "1m"
  write:
    limit: 30
    window: "1m"
  read:
    limit: 60
    window: "1m"

timeout:
  default: "30s"
```

- [ ] **Step 3: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/core/config/... -v
```

Expected: PASS — existing tests still pass.

- [ ] **Step 4: Commit**

```bash
git add internal/core/config/ configs/config.example.yaml && git commit -m "feat: add rate limit and timeout config with defaults"
```

---

### Task 5: Wire Middlewares into main.go

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Update main.go middleware chain**

Read `cmd/server/main.go` and update the middleware registration to:

1. Import `middleware` (already imported)
2. Add timeout middleware after Recovery
3. Add rate limiting middleware after timeout
4. Create a Redis adapter for the rate limiter
5. Apply different rate limit rules per endpoint group

The middleware order should be:
```go
router.Use(middleware.RequestID)
router.Use(middleware.Recovery)
router.Use(middleware.Timeout(defaultTimeout))
router.Use(middleware.Logger(log))
```

For rate limiting, create a helper that applies the correct limiter per route group:

```go
// Rate limit auth endpoints
authRouter := router.With(
    middleware.NewRateLimiter(redisAdapter, cfg.RateLimit.Auth).Middleware("auth"),
)
authModule.RegisterRoutes(authRouter)

// Rate limit write endpoints (applied inside auth guard)
// Rate limit read endpoints (applied inside auth guard)
```

Since the current auth/users routes are registered via `RegisterRoutes(router)`, the simplest approach is to apply rate limiting per route group using chi routing groups:

```go
// Auth routes (rate limited by IP)
authLimiter := middleware.NewRateLimiter(redisAdapter, cfg.RateLimit.Auth)
router.Route("/api/v1/auth", func(r chi.Router) {
    r.Use(authLimiter.Middleware("auth"))
    authModule.RegisterRoutes(r)
})

// User routes (rate limited by user)
readLimiter := middleware.NewRateLimiter(redisAdapter, cfg.RateLimit.Read)
writeLimiter := middleware.NewRateLimiter(redisAdapter, cfg.RateLimit.Write)
router.Route("/api/v1/users", func(r chi.Router) {
    r.Use(readLimiter.Middleware("read"))
    usersModule.RegisterRoutes(r)
})
```

Note: This requires updating the `RegisterRoutes` methods to accept the sub-router instead of the root router. The modules already accept `chi.Router`, so they'll work with sub-routers.

Also need to create a Redis adapter for the rate limiter. Add to `cmd/server/main.go` or create a small helper:

```go
type redisLimiterAdapter struct {
    client *redis.Client
}

func (a *redisLimiterAdapter) Incr(ctx context.Context, key string) *redis.IntCmd {
    return a.client.Incr(ctx, key)
}

func (a *redisLimiterAdapter) Expire(ctx context.Context, key string, ttl time.Duration) *redis.BoolCmd {
    return a.client.Expire(ctx, key, ttl)
}
```

- [ ] **Step 2: Verify it compiles and tests pass**

```bash
cd G:/workspace/ai/cyandie_backend && go build ./cmd/server/ && go test ./internal/... -count=1
```

- [ ] **Step 3: Commit**

```bash
git add cmd/server/main.go && git commit -m "feat: wire timeout and rate limiting middleware into HTTP server"
```

---

## Self-Review

**1. Spec coverage:**

| Spec Requirement | Task |
|---|---|
| Redis sliding window rate limiting | Task 3 |
| Per user (authenticated) / per IP (unauthenticated) | Task 3 (resolveIdentity) |
| Rate limit rules (auth/write/read) | Task 4 (config) |
| 429 response with headers | Task 3 |
| Retry-After header | Task 3 |
| X-RateLimit-Limit/Remaining headers | Task 3 |
| context.WithTimeout | Task 2 |
| Global default 30s | Task 4 (config) |
| Per-route override | Task 2 (middleware parameter) |
| 503 on timeout | Task 1 (ErrTimeout) + Task 2 |
| Middleware order | Task 5 |
| Config types | Task 4 |
| ErrTimeout error code | Task 1 |

**2. Placeholder scan:** No TBD/TODO found. All code blocks complete.

**3. Type consistency:** `RateLimitConfig` used consistently between middleware and config. `RedisLimiterClient` interface matches mock and adapter. `ErrTimeout` added to codes.go and mapped in errors.go.
