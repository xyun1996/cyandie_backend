# Cyandie Phase 2: Platform Adapters Design

Date: 2026-05-12

## Overview

Platform Adapters module enables third-party platform identity integration. Users can log in via WeChat OAuth (web), bind/unbind platform accounts, and the system maps platform identities to Cyandie user accounts. The module uses capability-based interfaces (OAuthProvider, PaymentProvider) so new platforms can be added without modifying core code.

## Key Design Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Interface model | Capability-based (OAuthProvider, PaymentProvider) | Platforms implement only the capabilities they support |
| WeChat scope | Web authorization only (PC scan login) | Most common, other flows deferred |
| Login flow | Direct third-party login + post-login binding | Both flows supported for flexibility |
| Token storage | Plaintext in DB, protected by access control | Confirmed in earlier design review |
| Platform registry | In-memory map, populated at startup | Small number of platforms, no dynamic loading needed |

## OAuth Flow

### Direct Third-Party Login

1. Client requests `GET /api/v1/platforms/{name}/auth-url` → server returns WeChat OAuth URL
2. User authorizes on WeChat, WeChat redirects to callback URL with `code`
3. Client sends `POST /api/v1/platforms/{name}/callback` with `code`
4. Server exchanges code for access_token + user info via WeChat API
5. Server looks up `platform_bindings` table for `wechat:{openid}`
   - Found → generate JWT for existing user, return tokens
   - Not found → create new user + credential + platform_binding, generate JWT, return tokens

### Post-Login Binding

1. Authenticated user sends `POST /api/v1/platforms/{name}/bind` with `code`
2. Server exchanges code for access_token + user info
3. Server creates `platform_bindings` record linking platform identity to current user
4. If platform identity already bound to another user → return `PLATFORM_BINDING_EXISTS` error

### Unbinding

1. Authenticated user sends `DELETE /api/v1/platforms/{name}/bind`
2. Server deletes `platform_bindings` record
3. If user has no other credentials (no password, no other platforms) → return error, must set password first

## Interfaces

```go
type OAuthProvider interface {
    Name() string
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
    GetUserInfo(ctx context.Context, token *OAuthToken) (*PlatformUser, error)
}

type OAuthToken struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Scope        string
}

type PlatformUser struct {
    PlatformID   string // OpenID for WeChat
    DisplayName  string
    AvatarURL    string
    RawData      map[string]any
}
```

## Platform Registry

```go
type PlatformRegistry struct {
    oauthProviders map[string]OAuthProvider
}

func (r *PlatformRegistry) RegisterOAuth(provider OAuthProvider) {
    r.oauthProviders[provider.Name()] = provider
}

func (r *PlatformRegistry) GetOAuth(name string) (OAuthProvider, bool) {
    p, ok := r.oauthProviders[name]
    return p, ok
}
```

Registered at startup via module `RegisterServices()`.

## WeChat OAuth Provider

WeChat web authorization flow:

1. Build auth URL: `https://open.weixin.qq.com/connect/qrconnect?appid={APPID}&redirect_uri={REDIRECT_URI}&response_type=code&scope=snsapi_login&state={STATE}`
2. Exchange code: `POST https://api.weixin.qq.com/sns/oauth2/access_token?appid={APPID}&secret={SECRET}&code={CODE}&grant_type=authorization_code`
3. Response contains: `access_token`, `openid`, `expires_in`
4. Get user info: `GET https://api.weixin.qq.com/sns/userinfo?access_token={TOKEN}&openid={OPENID}`
5. Response contains: `nickname`, `headimgurl`, `openid`

Configuration:

```yaml
platforms:
  wechat:
    app_id: "${WECHAT_APP_ID}"
    app_secret: "${WECHAT_APP_SECRET}"
    redirect_uri: "${WECHAT_REDIRECT_URI}"
```

## HTTP API

```
GET    /api/v1/platforms                    # List available platforms
GET    /api/v1/platforms/{name}/auth-url    # Get OAuth URL for platform
POST   /api/v1/platforms/{name}/callback    # OAuth callback (login or register)
POST   /api/v1/platforms/{name}/bind        # Bind platform to current user (auth required)
DELETE /api/v1/platforms/{name}/bind         # Unbind platform from current user (auth required)
GET    /api/v1/platforms/bindings            # List current user's platform bindings (auth required)
```

### Response Examples

GET /api/v1/platforms:
```json
{
  "ok": true,
  "data": [
    {"name": "wechat", "capabilities": ["oauth"], "authUrl": "/api/v1/platforms/wechat/auth-url"}
  ]
}
```

GET /api/v1/platforms/wechat/auth-url:
```json
{
  "ok": true,
  "data": {
    "url": "https://open.weixin.qq.com/connect/qrconnect?appid=...&state=abc123",
    "state": "abc123"
  }
}
```

POST /api/v1/platforms/wechat/callback:
```json
{
  "ok": true,
  "data": {
    "accessToken": "eyJ...",
    "refreshToken": "ref_...",
    "isNewUser": true,
    "userId": "uuid"
  }
}
```

## State Management

OAuth `state` parameter prevents CSRF attacks. Stored in Redis with 10-minute TTL:

- Key: `oauth_state:{state_value}`
- Value: `{"platform":"wechat","created_at":1715433600}`
- On callback, verify state exists and matches platform, then delete

## Database

Uses existing `platform_bindings` table from migration 002. No new migrations needed.

## Error Codes

Already defined in `internal/core/errors/codes.go`:
- `ErrPlatformNotSupported` — platform name not in registry
- `ErrPlatformAuthFailed` — code exchange or user info fetch failed
- `ErrPlatformBindingExists` — platform identity already bound to another user

New error codes needed:
- `ErrPlatformNotBound` — trying to unbind a platform that isn't bound
- `ErrLastCredential` — cannot unbind, user would have no way to log in

## Security

- OAuth state parameter stored in Redis, verified and deleted on callback
- WeChat app_secret never exposed to client (server-side code exchange only)
- Platform access_tokens stored in DB (plaintext, protected by DB access control)
- Callback endpoint rate-limited (auth group: 10/min/IP)
- Bind/unbind endpoints require authentication

## Observability

- Platform OAuth flow logged: platform, success/failure, new user vs existing
- WeChat API call latency and error rate logged
- Platform binding/unbinding events logged with user_id and platform
