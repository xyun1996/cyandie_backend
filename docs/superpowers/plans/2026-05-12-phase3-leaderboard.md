# Phase 3: Leaderboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Leaderboard module with Redis Sorted Set ranking, score submission, ranking queries, and board configuration.

**Architecture:** Leaderboard is a plugin module. Redis Sorted Sets store real-time rankings (primary query path). PostgreSQL stores board config and score snapshots. The module registers with the app via Module interface and exposes HTTP routes.

**Tech Stack:** Go, chi v5, sqlc, Redis (go-redis/v9), slog

---

### Task 1: Database Migration + sqlc Queries

**Files:**
- Create: `migrations/004_leaderboard.sql`
- Create: `queries/leaderboard.sql`
- Regenerate: `internal/db/`

- [ ] **Step 1: Create migration**

Create `migrations/004_leaderboard.sql`:

```sql
-- +goose Up
CREATE TABLE leaderboard_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL,
    update_strategy VARCHAR(16) NOT NULL DEFAULT 'highest',
    max_entries INT DEFAULT 1000,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE leaderboard_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL REFERENCES leaderboard_configs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score DOUBLE PRECISION NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_scores_board_user ON leaderboard_scores(board_id, user_id);

-- +goose Down
DROP TABLE IF EXISTS leaderboard_scores;
DROP TABLE IF EXISTS leaderboard_configs;
```

- [ ] **Step 2: Create queries**

Create `queries/leaderboard.sql`:

```sql
-- name: GetLeaderboardConfig :one
SELECT * FROM leaderboard_configs WHERE code = $1;

-- name: ListLeaderboardConfigs :many
SELECT * FROM leaderboard_configs ORDER BY name;

-- name: CreateLeaderboardConfig :one
INSERT INTO leaderboard_configs (code, name, update_strategy, max_entries, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateScore :one
INSERT INTO leaderboard_scores (board_id, user_id, score)
VALUES ($1, $2, $3)
RETURNING *;
```

- [ ] **Step 3: Regenerate sqlc**

```bash
cd G:/workspace/ai/cyandie_backend
sqlc generate
go build ./internal/db/...
```

- [ ] **Step 4: Commit**

```bash
git add migrations/ queries/ internal/db/ && git commit -m "feat: add leaderboard migrations and sqlc queries"
```

---

### Task 2: Leaderboard Service

**Files:**
- Create: `internal/leaderboard/service.go`
- Create: `internal/leaderboard/service_test.go`

- [ ] **Step 1: Create service**

Create `internal/leaderboard/service.go`:

```go
package leaderboard

import (
	"context"
	"fmt"
	"time"

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
```

- [ ] **Step 2: Create test with mock**

Create `internal/leaderboard/service_test.go`:

```go
package leaderboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
		UpdateStrategy: "highest", MaxEntries: 100, Metadata: json.RawMessage(`{}`),
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
```

- [ ] **Step 3: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/leaderboard/... -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/leaderboard/ && git commit -m "feat: add LeaderboardService with Redis Sorted Set ranking"
```

---

### Task 3: HTTP Handlers + Module + Wire

**Files:**
- Create: `internal/leaderboard/handler.go`
- Create: `internal/leaderboard/handler_test.go`
- Create: `internal/leaderboard/module.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Create handler**

Create `internal/leaderboard/handler.go`:

```go
package leaderboard

import (
	"encoding/json"
	"net/http"
	"strconv"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *LeaderboardService
}

func NewHandler(svc *LeaderboardService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/api/v1/leaderboard/{code}", h.getRanking)
	router.Post("/api/v1/leaderboard/{code}/submit", h.submitScore)
	router.Get("/api/v1/leaderboard/{code}/me", h.getMyRank)
}

func (h *Handler) getRanking(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	entries, err := h.svc.GetRanking(r.Context(), code, limit, offset)
	if err != nil {
		writeLBError(w, err)
		return
	}
	writeLBJSON(w, http.StatusOK, entries)
}

func (h *Handler) submitScore(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}

	code := chi.URLParam(r, "code")
	var req struct {
		Score float64 `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	if err := h.svc.SubmitScore(r.Context(), code, userID, req.Score); err != nil {
		writeLBError(w, err)
		return
	}
	writeLBJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getMyRank(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}

	code := chi.URLParam(r, "code")
	entry, err := h.svc.GetUserRank(r.Context(), code, userID)
	if err != nil {
		writeLBError(w, err)
		return
	}
	writeLBJSON(w, http.StatusOK, entry)
}

func writeLBJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writeLBError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*coreerrors.AppError); ok {
		appErr.WriteHTTP(w)
		return
	}
	coreerrors.New(coreerrors.ErrInternal, "internal error").WriteHTTP(w)
}
```

- [ ] **Step 2: Create handler test**

Create `internal/leaderboard/handler_test.go`:

```go
package leaderboard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_GetRanking(t *testing.T) {
	svc := newTestLBService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard/test-board?limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SubmitScore(t *testing.T) {
	svc := newTestLBService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/leaderboard/test-board/submit", nil)
	req.Header.Set("X-User-ID", "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should get 400 since no body, but at least not 500
	if rec.Code == http.StatusInternalServerError {
		t.Errorf("expected not 500, got %d", rec.Code)
	}
}
```

- [ ] **Step 3: Create module**

Create `internal/leaderboard/module.go`:

```go
package leaderboard

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler *Handler
	service *LeaderboardService
}

func NewModule(queries db.Querier, redis redisClient) *Module {
	service := NewLeaderboardService(queries, redis)
	handler := NewHandler(service)
	return &Module{handler: handler, service: service}
}

func (m *Module) Name() string { return "leaderboard" }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("leaderboard", m.service)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	m.handler.RegisterRoutes(router)
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }
```

Note: `redisClient` is the interface defined in service.go. The `cache.Cache` struct embeds `*redis.Client` which implements all needed methods. We need a thin adapter or use the redis.Client directly.

- [ ] **Step 4: Update main.go**

Read `cmd/server/main.go` and add:
- Import `"github.com/cyandie/backend/internal/leaderboard"`
- Create leaderboard module using `rdb.Client` as the redis client
- Register with app and add routes

The `redisClient` interface in service.go needs ZAdd, ZRevRank, ZRevRangeWithScores, ZScore, ZCard. The `*redis.Client` already implements all of these, so we can pass it directly.

```go
leaderboardModule := leaderboard.NewModule(queries, rdb.Client)
app.Register(leaderboardModule)

router.Route("/api/v1/leaderboard", func(r chi.Router) {
    readLimiter := middleware.NewRateLimiter(redisAdapter, middleware.RateLimitConfig(cfg.RateLimit.Read))
    r.Use(readLimiter.Middleware("read"))
    leaderboardModule.RegisterRoutes(r)
})
```

- [ ] **Step 5: Verify**

```bash
cd G:/workspace/ai/cyandie_backend && go build ./cmd/server/ && go test ./internal/... -count=1
```

- [ ] **Step 6: Commit**

```bash
git add internal/leaderboard/ cmd/server/main.go && git commit -m "feat: add Leaderboard module with Redis Sorted Set ranking and HTTP API"
```

---

## Self-Review

**1. Spec coverage:** Redis Sorted Set ranking ✅, Score submission ✅, Ranking queries ✅, Board config (PG) ✅, HTTP API ✅, Module wiring ✅

**2. Placeholder scan:** No TBD/TODO.

**3. Type consistency:** `LeaderboardService`, `Handler`, `Module`, `RankEntry`, `redisClient` — all consistent.
