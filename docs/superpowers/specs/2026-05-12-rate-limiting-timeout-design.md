# Cyandie Rate Limiting and Request Timeout Design

Date: 2026-05-12

## Overview

Add rate limiting and request timeout middleware to all HTTP APIs. Rate limiting uses Redis sliding window counters, scoped per user (authenticated) or per IP (unauthenticated). Request timeout uses Go context deadlines with a global default and per-route overrides.

## Key Design Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Rate limit algorithm | Redis sliding window | Smooth, supports multi-instance, zero new dependencies |
| Rate limit scope | Per user (authenticated), per IP (unauthenticated) | Prevents single-user abuse, protects auth endpoints |
| Timeout mechanism | context.WithTimeout in chi middleware | Go idiom, propagates to all downstream calls |
| Default timeout | 30s global, per-route override | Covers most APIs, slow endpoints can opt out |

## Rate Limiting Middleware

### Algorithm

Redis sliding window counter. For each request:
1. Determine identity: user_id from context (authenticated) or IP from request (unauthenticated)
2. Determine endpoint group: auth, write, or read
3. Redis key: `ratelimit:{identity}:{group}:{window_timestamp}`
4. INCR key, EXPIRE to window size
5. Sum counts across recent windows (e.g. 60 windows of 1s = 1 minute)
6. If sum > limit → reject with 429

### Rate Limit Rules

Configurable via YAML:

```yaml
rate_limit:
  auth:
    limit: 10        # requests per window
    window: "1m"     # window duration
  write:
    limit: 30
    window: "1m"
  read:
    limit: 60
    window: "1m"
```

Default rules (no config required):

| Endpoint Group | Limit | Window | Scope |
|---------------|-------|--------|-------|
| auth (login, register, refresh) | 10 | 1min | IP |
| write (POST, PUT, DELETE with auth) | 30 | 1min | User |
| read (GET with auth) | 60 | 1min | User |

### Response Headers

Every response includes rate limit info:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1715433600
```

When rate limited (429):

```
Retry-After: 30
```

### Response Body (429)

```json
{
  "ok": false,
  "error": {
    "code": "RATE_LIMITED",
    "message": "rate limit exceeded, retry after 30s",
    "requestId": "req_..."
  }
}
```

### Endpoint Group Detection

The middleware determines the group from the request:

- **auth**: Path starts with `/api/v1/auth/`
- **write**: Authenticated + method is POST/PUT/DELETE/PATCH
- **read**: Authenticated + method is GET

## Request Timeout Middleware

### Mechanism

chi middleware that wraps each request with `context.WithTimeout`:

```go
func Timeout(duration time.Duration) func(http.Handler) http.Handler
```

### Default and Override

Global default (30s) applied in main.go:

```go
router.Use(middleware.Timeout(30 * time.Second))
```

Per-route override:

```go
router.With(middleware.Timeout(60 * time.Second)).Post("/api/v1/chat/rooms/{id}/messages", h.sendMessage)
```

### Timeout Behavior

1. Context deadline set on request start
2. All downstream operations (DB, Redis, gRPC) respect context cancellation
3. On timeout, middleware returns 503:

```json
{
  "ok": false,
  "error": {
    "code": "TIMEOUT",
    "message": "request timeout",
    "requestId": "req_..."
  }
}
```

4. Logger records timeout with path, duration, user_id

### Middleware Order

```go
router.Use(middleware.RequestID)
router.Use(middleware.Recovery)
router.Use(middleware.Timeout(30 * time.Second))
router.Use(middleware.RateLimit(rateLimiter))
router.Use(middleware.Logger(log))
```

Timeout after Recovery so panics from timeout are caught. Before Logger so timeout is logged. Before RateLimit so rate limit Redis calls also respect the timeout.

## Configuration

Add to `internal/core/config/config.go`:

```go
type RateLimitConfig struct {
    Auth  RateLimitRule `yaml:"auth"`
    Write RateLimitRule `yaml:"write"`
    Read  RateLimitRule `yaml:"read"`
}

type RateLimitRule struct {
    Limit  int    `yaml:"limit"`
    Window string `yaml:"window"`
}

type TimeoutConfig struct {
    Default string            `yaml:"default"`
    Routes  map[string]string `yaml:"routes"`
}
```

Add to Config struct:

```go
RateLimit RateLimitConfig `yaml:"rate_limit"`
Timeout   TimeoutConfig   `yaml:"timeout"`
```

## Error Codes

Add to `internal/core/errors/codes.go`:

```go
const (
    ErrRateLimited = "RATE_LIMITED"
    ErrTimeout     = "TIMEOUT"
)
```

`ErrRateLimited` already exists. `ErrTimeout` is new.

## Observability

- Rate limit rejections logged with identity, endpoint group, current count
- Timeout events logged with path, duration, user_id
- Prometheus metrics:
  - `cyandie_ratelimit_rejected_total{group, identity_type}`
  - `cyandie_request_timeout_total{path, method}`
