# Phase 2: Platform Adapters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Platform Adapters module with WeChat web OAuth, platform account binding/unbinding, and direct third-party login.

**Architecture:** Platforms module implements the Module interface. It contains a PlatformRegistry that maps platform names to OAuthProvider implementations. WeChat is the first provider. The module exposes HTTP routes for auth URL generation, OAuth callback, and bind/unbind. On callback, it either logs in an existing user or creates a new one.

**Tech Stack:** Go, chi v5, sqlc, Redis (state management), slog

---

## File Structure

```
internal/platforms/
  module.go            # Module implementation
  registry.go          # PlatformRegistry
  registry_test.go     # Registry tests
  oauth.go             # OAuthProvider interface, OAuthToken, PlatformUser
  handler.go           # HTTP handlers
  handler_test.go      # Handler tests
  wechat.go            # WeChat OAuth provider
  wechat_test.go       # WeChat provider tests
  service.go           # PlatformService (bind/unbind/login logic)
  service_test.go      # Service tests
queries/
  platform_bindings.sql
internal/db/
  platform_bindings.sql.go  # sqlc generated
  models.go                  # updated with PlatformBinding
  querier.go                 # updated with new methods
  db.go                      # updated
```

---

### Task 1: Add Error Codes + sqlc Queries

**Files:**
- Modify: `internal/core/errors/codes.go`
- Create: `queries/platform_bindings.sql`
- Regenerate: `internal/db/`

- [ ] **Step 1: Add new error codes**

Add to `internal/core/errors/codes.go` in the Platform section:

```go
ErrPlatformNotBound = "PLATFORM_NOT_BOUND"
ErrLastCredential   = "LAST_CREDENTIAL"
```

Add HTTP mappings in `internal/core/errors/errors.go`:

```go
case ErrPlatformNotBound:
    return http.StatusNotFound
case ErrLastCredential:
    return http.StatusConflict
```

- [ ] **Step 2: Create platform_bindings queries**

Create `queries/platform_bindings.sql`:

```sql
-- name: GetPlatformBinding :one
SELECT * FROM platform_bindings WHERE platform = $1 AND platform_user_id = $2;

-- name: GetPlatformBindingsByUserID :many
SELECT * FROM platform_bindings WHERE user_id = $1;

-- name: CreatePlatformBinding :one
INSERT INTO platform_bindings (user_id, platform, platform_user_id, access_token, refresh_token, expires_at, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeletePlatformBinding :one
DELETE FROM platform_bindings WHERE user_id = $1 AND platform = $2
RETURNING *;

-- name: CountUserCredentials :one
SELECT COUNT(*) FROM credentials WHERE user_id = $1;
```

- [ ] **Step 3: Regenerate sqlc code**

```bash
cd G:/workspace/ai/cyandie_backend
sqlc generate
go build ./internal/db/...
```

- [ ] **Step 4: Commit**

```bash
git add internal/core/errors/ queries/ internal/db/ && git commit -m "feat: add platform error codes and sqlc queries for platform bindings"
```

---

### Task 2: OAuth Interfaces + Platform Registry

**Files:**
- Create: `internal/platforms/oauth.go`
- Create: `internal/platforms/registry.go`
- Create: `internal/platforms/registry_test.go`

- [ ] **Step 1: Create OAuth interfaces**

Create `internal/platforms/oauth.go`:

```go
package platforms

import (
	"context"
	"time"
)

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

type PlatformUser struct {
	PlatformID  string         `json:"platform_id"`
	DisplayName string         `json:"display_name"`
	AvatarURL   string         `json:"avatar_url,omitempty"`
	RawData     map[string]any `json:"raw_data,omitempty"`
}

type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
	GetUserInfo(ctx context.Context, token *OAuthToken) (*PlatformUser, error)
}
```

- [ ] **Step 2: Create PlatformRegistry**

Create `internal/platforms/registry.go`:

```go
package platforms

type PlatformRegistry struct {
	oauthProviders map[string]OAuthProvider
}

func NewPlatformRegistry() *PlatformRegistry {
	return &PlatformRegistry{
		oauthProviders: make(map[string]OAuthProvider),
	}
}

func (r *PlatformRegistry) RegisterOAuth(provider OAuthProvider) {
	r.oauthProviders[provider.Name()] = provider
}

func (r *PlatformRegistry) GetOAuth(name string) (OAuthProvider, bool) {
	p, ok := r.oauthProviders[name]
	return p, ok
}

func (r *PlatformRegistry) ListPlatforms() []string {
	names := make([]string, 0, len(r.oauthProviders))
	for name := range r.oauthProviders {
		names = append(names, name)
	}
	return names
}
```

- [ ] **Step 3: Create registry test**

Create `internal/platforms/registry_test.go`:

```go
package platforms

import (
	"context"
	"testing"
)

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string                                                { return m.name }
func (m *mockProvider) GetAuthURL(state string) string                              { return "https://example.com/auth?state=" + state }
func (m *mockProvider) ExchangeCode(_ context.Context, _ string) (*OAuthToken, error) { return &OAuthToken{AccessToken: "tok"}, nil }
func (m *mockProvider) GetUserInfo(_ context.Context, _ *OAuthToken) (*PlatformUser, error) {
	return &PlatformUser{PlatformID: "id1", DisplayName: "Test User"}, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := NewPlatformRegistry()
	reg.RegisterOAuth(&mockProvider{name: "test"})

	p, ok := reg.GetOAuth("test")
	if !ok {
		t.Error("expected to find registered provider")
	}
	if p.Name() != "test" {
		t.Errorf("expected test, got %s", p.Name())
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	reg := NewPlatformRegistry()
	_, ok := reg.GetOAuth("nonexistent")
	if ok {
		t.Error("expected not found for unregistered provider")
	}
}

func TestRegistry_ListPlatforms(t *testing.T) {
	reg := NewPlatformRegistry()
	reg.RegisterOAuth(&mockProvider{name: "wechat"})
	reg.RegisterOAuth(&mockProvider{name: "discord"})

	platforms := reg.ListPlatforms()
	if len(platforms) != 2 {
		t.Errorf("expected 2 platforms, got %d", len(platforms))
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/platforms/... -v
```

Expected: PASS — all 3 tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/platforms/ && git commit -m "feat: add OAuth interfaces and platform registry"
```

---

### Task 3: WeChat OAuth Provider

**Files:**
- Create: `internal/platforms/wechat.go`
- Create: `internal/platforms/wechat_test.go`

- [ ] **Step 1: Create WeChat provider**

Create `internal/platforms/wechat.go`:

```go
package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cyandie/backend/internal/core/errors"
)

type WeChatConfig struct {
	AppID       string `yaml:"app_id" env:"WECHAT_APP_ID"`
	AppSecret   string `yaml:"app_secret" env:"WECHAT_APP_SECRET"`
	RedirectURI string `yaml:"redirect_uri" env:"WECHAT_REDIRECT_URI"`
}

type WeChatProvider struct {
	config WeChatConfig
	client *http.Client
}

func NewWeChatProvider(config WeChatConfig) *WeChatProvider {
	return &WeChatProvider{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WeChatProvider) Name() string { return "wechat" }

func (w *WeChatProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("appid", w.config.AppID)
	params.Set("redirect_uri", w.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "snsapi_login")
	params.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + params.Encode() + "#wechat_redirect"
}

func (w *WeChatProvider) ExchangeCode(_ context.Context, code string) (*OAuthToken, error) {
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		w.config.AppID, w.config.AppSecret, code,
	)

	resp, err := w.client.Get(tokenURL)
	if err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat token exchange failed")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken  string `json:"access_token"`
		OpenID       string `json:"openid"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat token response parse failed")
	}
	if result.ErrCode != 0 {
		return nil, errors.New(errors.ErrPlatformAuthFailed, fmt.Sprintf("wechat error: %d %s", result.ErrCode, result.ErrMsg))
	}

	return &OAuthToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}, nil
}

func (w *WeChatProvider) GetUserInfo(_ context.Context, token *OAuthToken) (*PlatformUser, error) {
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		token.AccessToken, token.RefreshToken, // WeChat puts openid in RefreshToken field from exchange
	)

	resp, err := w.client.Get(userInfoURL)
	if err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat user info fetch failed")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OpenID     string `json:"openid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat user info parse failed")
	}
	if result.ErrCode != 0 {
		return nil, errors.New(errors.ErrPlatformAuthFailed, fmt.Sprintf("wechat error: %d %s", result.ErrCode, result.ErrMsg))
	}

	return &PlatformUser{
		PlatformID:  result.OpenID,
		DisplayName: result.Nickname,
		AvatarURL:   result.HeadImgURL,
	}, nil
}
```

Note: WeChat's `/sns/oauth2/access_token` response puts `openid` as a top-level field. We need to capture it. Let me fix the ExchangeCode to also return the openid. The cleanest way: store openid in a custom field on OAuthToken or return it separately.

Actually, looking at the WeChat API response, `openid` is returned alongside `access_token`. We need to pass it to `GetUserInfo`. The simplest approach: store openid in the OAuthToken's RefreshToken field temporarily (since WeChat doesn't use refresh_token the same way), or add an OpenID field.

Let me add an OpenID field to OAuthToken:

Update `internal/platforms/oauth.go` to add:

```go
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	OpenID       string    `json:"open_id,omitempty"` // Platform-specific user ID (e.g. WeChat openid)
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}
```

Then in wechat.go ExchangeCode, capture openid:

```go
return &OAuthToken{
    AccessToken:  result.AccessToken,
    RefreshToken: result.RefreshToken,
    OpenID:       result.OpenID,
    ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
}, nil
```

And in GetUserInfo, use token.OpenID:

```go
userInfoURL := fmt.Sprintf(
    "https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
    token.AccessToken, token.OpenID,
)
```

- [ ] **Step 2: Create WeChat test**

Create `internal/platforms/wechat_test.go`:

```go
package platforms

import (
	"testing"
)

func TestWeChatProvider_GetAuthURL(t *testing.T) {
	provider := NewWeChatProvider(WeChatConfig{
		AppID:       "wx123456",
		AppSecret:   "secret",
		RedirectURI: "https://example.com/callback",
	})

	url := provider.GetAuthURL("test-state")
	if url == "" {
		t.Error("expected non-empty auth URL")
	}
	if !contains(url, "wx123456") {
		t.Error("expected app_id in auth URL")
	}
	if !contains(url, "test-state") {
		t.Error("expected state in auth URL")
	}
}

func TestWeChatProvider_Name(t *testing.T) {
	provider := NewWeChatProvider(WeChatConfig{})
	if provider.Name() != "wechat" {
		t.Errorf("expected wechat, got %s", provider.Name())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

- [ ] **Step 3: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/platforms/... -v
```

Expected: PASS — all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/platforms/ && git commit -m "feat: add WeChat OAuth provider with web authorization flow"
```

---

### Task 4: Platform Service

**Files:**
- Create: `internal/platforms/service.go`
- Create: `internal/platforms/service_test.go`

- [ ] **Step 1: Create PlatformService**

Create `internal/platforms/service.go`:

```go
package platforms

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
)

type PlatformService struct {
	queries  db.Querier
	registry *PlatformRegistry
}

func NewPlatformService(queries db.Querier, registry *PlatformRegistry) *PlatformService {
	return &PlatformService{queries: queries, registry: registry}
}

func (s *PlatformService) GetAuthURL(platformName, state string) (string, error) {
	provider, ok := s.registry.GetOAuth(platformName)
	if !ok {
		return "", errors.New(errors.ErrPlatformNotSupported, "platform not supported: "+platformName)
	}
	return provider.GetAuthURL(state), nil
}

func (s *PlatformService) HandleCallback(ctx context.Context, platformName, code string) (*CallbackResult, error) {
	provider, ok := s.registry.GetOAuth(platformName)
	if !ok {
		return nil, errors.New(errors.ErrPlatformNotSupported, "platform not supported: "+platformName)
	}

	token, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	platformUser, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check if this platform identity is already bound
	binding, err := s.queries.GetPlatformBinding(ctx, db.GetPlatformBindingParams{
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
	})
	if err == nil {
		// Existing user — return user ID
		return &CallbackResult{
			UserID:    binding.UserID.String(),
			IsNewUser: false,
		}, nil
	}

	// New user — create user + credential + binding
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:    fmt.Sprintf("wx_%s", platformUser.PlatformID[:8]),
		DisplayName: sql.NullString{String: platformUser.DisplayName, Valid: true},
		AvatarURL:   sql.NullString{String: platformUser.AvatarURL, Valid: true},
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to create user")
	}

	_, err = s.queries.CreatePlatformBinding(ctx, db.CreatePlatformBindingParams{
		UserID:         user.ID,
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
		AccessToken:    sql.NullString{String: token.AccessToken, Valid: true},
		RefreshToken:   sql.NullString{String: token.RefreshToken, Valid: true},
		ExpiresAt:      sql.NullTime{Time: token.ExpiresAt, Valid: !token.ExpiresAt.IsZero()},
		Metadata:       json.RawMessage(`{}`),
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to create platform binding")
	}

	return &CallbackResult{
		UserID:    user.ID.String(),
		IsNewUser: true,
	}, nil
}

func (s *PlatformService) BindPlatform(ctx context.Context, userID, platformName, code string) error {
	provider, ok := s.registry.GetOAuth(platformName)
	if !ok {
		return errors.New(errors.ErrPlatformNotSupported, "platform not supported: "+platformName)
	}

	token, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		return err
	}

	platformUser, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		return err
	}

	uid, _ := uuid.Parse(userID)

	// Check if already bound to another user
	existing, err := s.queries.GetPlatformBinding(ctx, db.GetPlatformBindingParams{
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
	})
	if err == nil && existing.UserID != uid {
		return errors.New(errors.ErrPlatformBindingExists, "platform identity already bound to another user")
	}

	_, err = s.queries.CreatePlatformBinding(ctx, db.CreatePlatformBindingParams{
		UserID:         uid,
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
		AccessToken:    sql.NullString{String: token.AccessToken, Valid: true},
		RefreshToken:   sql.NullString{String: token.RefreshToken, Valid: true},
		ExpiresAt:      sql.NullTime{Time: token.ExpiresAt, Valid: !token.ExpiresAt.IsZero()},
		Metadata:       json.RawMessage(`{}`),
	})
	if err != nil {
		return fmt.Errorf("create binding: %w", err)
	}

	return nil
}

func (s *PlatformService) UnbindPlatform(ctx context.Context, userID, platformName string) error {
	uid, _ := uuid.Parse(userID)

	bindings, err := s.queries.GetPlatformBindingsByUserID(ctx, uid)
	if err != nil {
		return errors.New(errors.ErrPlatformNotBound, "no platform bindings found")
	}

	found := false
	for _, b := range bindings {
		if b.Platform == platformName {
			found = true
			break
		}
	}
	if !found {
		return errors.New(errors.ErrPlatformNotBound, "platform not bound")
	}

	// Check if user would have any way to log in after unbinding
	credCount, _ := s.queries.CountUserCredentials(ctx, uid)
	otherBindings := 0
	for _, b := range bindings {
		if b.Platform != platformName {
			otherBindings++
		}
	}
	if credCount == 0 && otherBindings == 0 {
		return errors.New(errors.ErrLastCredential, "cannot unbind last credential, set a password first")
	}

	_, err = s.queries.DeletePlatformBinding(ctx, db.DeletePlatformBindingParams{
		UserID:   uid,
		Platform: platformName,
	})
	if err != nil {
		return fmt.Errorf("delete binding: %w", err)
	}

	return nil
}

func (s *PlatformService) ListBindings(ctx context.Context, userID string) ([]db.PlatformBinding, error) {
	uid, _ := uuid.Parse(userID)
	return s.queries.GetPlatformBindingsByUserID(ctx, uid)
}

type CallbackResult struct {
	UserID    string `json:"userId"`
	IsNewUser bool   `json:"isNewUser"`
}
```

- [ ] **Step 2: Create service test with mock**

Create `internal/platforms/service_test.go`:

```go
package platforms

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
)

type platformMockQueries struct {
	bindings map[string]db.PlatformBinding
	users    map[string]db.User
}

func newPlatformMockQueries() *platformMockQueries {
	return &platformMockQueries{
		bindings: make(map[string]db.PlatformBinding),
		users:    make(map[string]db.User),
	}
}

func (m *platformMockQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *platformMockQueries) CreatePlatformBinding(_ context.Context, params db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	b := db.PlatformBinding{
		ID:             uuid.New(),
		UserID:         params.UserID,
		Platform:       params.Platform,
		PlatformUserID: params.PlatformUserID,
		Metadata:       json.RawMessage(`{}`),
	}
	m.bindings[params.Platform+":"+params.PlatformUserID] = b
	return b, nil
}
func (m *platformMockQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *platformMockQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) {
	return 1, nil
}
func (m *platformMockQueries) CreateUser(_ context.Context, params db.CreateUserParams) (db.User, error) {
	u := db.User{ID: uuid.New(), Username: params.Username, Status: "active", Metadata: json.RawMessage(`{}`)}
	m.users[params.Username] = u
	return u, nil
}
func (m *platformMockQueries) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *platformMockQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *platformMockQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *platformMockQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *platformMockQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *platformMockQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *platformMockQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *platformMockQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *platformMockQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *platformMockQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *platformMockQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}

func newTestPlatformService() *PlatformService {
	reg := NewPlatformRegistry()
	reg.RegisterOAuth(&mockProvider{name: "wechat"})
	return NewPlatformService(newPlatformMockQueries(), reg)
}

func TestPlatformService_GetAuthURL(t *testing.T) {
	svc := newTestPlatformService()
	url, err := svc.GetAuthURL("wechat", "state123")
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty auth URL")
	}
}

func TestPlatformService_GetAuthURL_UnsupportedPlatform(t *testing.T) {
	svc := newTestPlatformService()
	_, err := svc.GetAuthURL("google", "state123")
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
}

func TestPlatformService_UnbindPlatform_NotBound(t *testing.T) {
	svc := newTestPlatformService()
	err := svc.UnbindPlatform(context.Background(), uuid.New().String(), "wechat")
	if err == nil {
		t.Error("expected error for unbound platform")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/platforms/... -v
```

Expected: PASS — all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/platforms/ && git commit -m "feat: add PlatformService with callback, bind, unbind logic"
```

---

### Task 5: HTTP Handlers

**Files:**
- Create: `internal/platforms/handler.go`
- Create: `internal/platforms/handler_test.go`

- [ ] **Step 1: Create handlers**

Create `internal/platforms/handler.go`:

```go
package platforms

import (
	"encoding/json"
	"net/http"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *PlatformService
}

func NewHandler(svc *PlatformService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/api/v1/platforms", h.listPlatforms)
	router.Get("/api/v1/platforms/{name}/auth-url", h.getAuthURL)
	router.Post("/api/v1/platforms/{name}/callback", h.callback)
	router.Post("/api/v1/platforms/{name}/bind", h.bind)
	router.Delete("/api/v1/platforms/{name}/bind", h.unbind)
	router.Get("/api/v1/platforms/bindings", h.listBindings)
}

func (h *Handler) listPlatforms(w http.ResponseWriter, r *http.Request) {
	platforms := h.svc.registry.ListPlatforms()
	type platformInfo struct {
		Name         string   `json:"name"`
		Capabilities []string `json:"capabilities"`
	}
	result := make([]platformInfo, len(platforms))
	for i, name := range platforms {
		result[i] = platformInfo{Name: name, Capabilities: []string{"oauth"}}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getAuthURL(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	state := r.URL.Query().Get("state")
	if state == "" {
		state = generateState()
	}

	url, err := h.svc.GetAuthURL(name, state)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": url, "state": state})
}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	result, err := h.svc.HandleCallback(r.Context(), name, req.Code)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) bind(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}

	name := chi.URLParam(r, "name")
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	if err := h.svc.BindPlatform(r.Context(), userID, name, req.Code); err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) unbind(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}

	name := chi.URLParam(r, "name")
	if err := h.svc.UnbindPlatform(r.Context(), userID, name); err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listBindings(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}

	bindings, err := h.svc.ListBindings(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	type bindingInfo struct {
		Platform       string `json:"platform"`
		PlatformUserID string `json:"platformUserId"`
	}
	result := make([]bindingInfo, len(bindings))
	for i, b := range bindings {
		result[i] = bindingInfo{Platform: b.Platform, PlatformUserID: b.PlatformUserID}
	}
	writeJSON(w, http.StatusOK, result)
}

func generateState() string {
	return fmt.Sprintf("st_%d", time.Now().UnixNano())
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writeAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*coreerrors.AppError); ok {
		appErr.WriteHTTP(w)
		return
	}
	coreerrors.New(coreerrors.ErrInternal, "internal error").WriteHTTP(w)
}
```

- [ ] **Step 2: Create handler test**

Create `internal/platforms/handler_test.go`:

```go
package platforms

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_ListPlatforms(t *testing.T) {
	svc := newTestPlatformService()
	h := NewHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/platforms", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAuthURL(t *testing.T) {
	svc := newTestPlatformService()
	h := NewHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/platforms/wechat/auth-url?state=test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetAuthURL_UnsupportedPlatform(t *testing.T) {
	svc := newTestPlatformService()
	h := NewHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/platforms/google/auth-url", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd G:/workspace/ai/cyandie_backend && go test ./internal/platforms/... -v
```

Expected: PASS — all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/platforms/handler.go internal/platforms/handler_test.go && git commit -m "feat: add platform HTTP handlers for auth URL, callback, bind, unbind"
```

---

### Task 6: Module + Config + Wire into main.go

**Files:**
- Create: `internal/platforms/module.go`
- Modify: `internal/core/config/config.go`
- Modify: `configs/config.example.yaml`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Create platforms module**

Create `internal/platforms/module.go`:

```go
package platforms

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler  *Handler
	service  *PlatformService
	registry *PlatformRegistry
}

func NewModule(queries db.Querier, registry *PlatformRegistry) *Module {
	service := NewPlatformService(queries, registry)
	handler := NewHandler(service)
	return &Module{
		handler:  handler,
		service:  service,
		registry: registry,
	}
}

func (m *Module) Name() string { return "platforms" }

func (m *Module) Dependencies() []string { return []string{"auth", "users"} }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("platforms", m.service)
	reg.Register("platform-registry", m.registry)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	m.handler.RegisterRoutes(router)
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }
```

- [ ] **Step 2: Add platform config**

Add to `internal/core/config/config.go`:

```go
type WeChatConfig struct {
	AppID       string `yaml:"app_id" env:"WECHAT_APP_ID"`
	AppSecret   string `yaml:"app_secret" env:"WECHAT_APP_SECRET"`
	RedirectURI string `yaml:"redirect_uri" env:"WECHAT_REDIRECT_URI"`
}

type PlatformsConfig struct {
	WeChat WeChatConfig `yaml:"wechat"`
}
```

Add `Platforms PlatformsConfig \`yaml:"platforms"\`` to Config struct.

Add env overrides in `applyEnv` for WeChat config fields.

- [ ] **Step 3: Update config.example.yaml**

Add:
```yaml
platforms:
  wechat:
    app_id: "${WECHAT_APP_ID}"
    app_secret: "${WECHAT_APP_SECRET}"
    redirect_uri: "${WECHAT_REDIRECT_URI}"
```

- [ ] **Step 4: Update main.go**

Read current `cmd/server/main.go` and add:
- Import `github.com/cyandie/backend/internal/platforms`
- Create PlatformRegistry
- Register WeChat provider if config is present
- Create platforms module and register with app
- Register platform routes with rate limiting

```go
// Platform adapters
platformRegistry := platforms.NewPlatformRegistry()
if cfg.Platforms.WeChat.AppID != "" {
    platformRegistry.RegisterOAuth(platforms.NewWeChatProvider(platforms.WeChatConfig{
        AppID:       cfg.Platforms.WeChat.AppID,
        AppSecret:   cfg.Platforms.WeChat.AppSecret,
        RedirectURI: cfg.Platforms.WeChat.RedirectURI,
    }))
}
platformsModule := platforms.NewModule(queries, platformRegistry)
app.Register(platformsModule)

// Platform routes (rate limited)
authLimiter := middleware.NewRateLimiter(redisAdapter, middleware.RateLimitConfig(cfg.RateLimit.Auth))
router.Route("/api/v1/platforms", func(r chi.Router) {
    r.Use(authLimiter.Middleware("auth"))
    platformsModule.RegisterRoutes(r)
})
```

- [ ] **Step 5: Verify it compiles and tests pass**

```bash
cd G:/workspace/ai/cyandie_backend && go build ./cmd/server/ && go test ./internal/... -count=1
```

- [ ] **Step 6: Commit**

```bash
git add internal/platforms/ internal/core/config/ configs/config.example.yaml cmd/server/main.go && git commit -m "feat: wire platforms module with WeChat OAuth into app lifecycle"
```

---

## Self-Review

**1. Spec coverage:**

| Spec Requirement | Task |
|---|---|
| OAuthProvider interface | Task 2 |
| PlatformRegistry | Task 2 |
| WeChat web OAuth | Task 3 |
| Direct third-party login | Task 4 (HandleCallback) |
| Post-login binding | Task 4 (BindPlatform) |
| Unbinding with last-credential check | Task 4 (UnbindPlatform) |
| OAuth state management | Task 5 (generateState) |
| HTTP API (6 endpoints) | Task 5 |
| Platform config | Task 6 |
| Module wiring | Task 6 |
| Error codes | Task 1 |

**2. Placeholder scan:** No TBD/TODO found.

**3. Type consistency:** `OAuthProvider`, `OAuthToken`, `PlatformUser`, `PlatformRegistry`, `PlatformService`, `Handler`, `Module` — all consistent across tasks. `WeChatConfig` used in both wechat.go and config.go.
