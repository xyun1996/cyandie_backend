# Cyandie Backend Framework Design

Date: 2026-05-11

## Overview

Cyandie is a modular, extensible, self-hostable backend framework for games, apps, communities, and real-time interactive services. It provides account management, platform identity, TCP chat, leaderboards, friends, and admin capabilities through a plugin-based architecture.

## Key Design Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Architecture | Core + Plugin | Modules are optional, boundaries are clear, fits "universal framework" |
| Language | Go | Performance, concurrency, strong typing |
| Database | PostgreSQL + Redis | PG for JSONB/extensibility, Redis for caching/leaderboard/presence |
| Data access | sqlc | Type-safe SQL, no ORM overhead, generated code |
| Migrations | goose | SQL-first, widely used |
| HTTP router | chi | Lightweight, idiomatic, middleware-friendly |
| Protobuf tooling | buf | Modern protobuf management, lint + generate |
| TCP protocol | TLV + Protobuf | Extensible, self-describing length, cross-language serialization |
| Auth | JWT (access + refresh) | Stateless access, Redis-backed refresh for revocation |
| Deployment | Modular monolith | Single process, modules communicate via interfaces, future split possible |

## Architecture

### Core + Plugin Model

The framework is split into a **Core** (always present) and **Plugins** (optional, registered at startup).

Core subsystems:

| Subsystem | Responsibility |
|-----------|---------------|
| Config | YAML + env var loading, hot reload |
| Database | PG connection pool + goose migrations |
| Cache | Redis connection + abstract cache interface |
| Server | HTTP (chi), gRPC, TCP server lifecycle |
| Logger | Structured logging (slog) |
| Errors | Unified error codes + HTTP/gRPC/TCP error mapping |
| Middleware | Auth, rate limiting, CORS, request logging |

Plugin registration:

```go
type Module interface {
    Name() string
    Dependencies() []string
    RegisterServices(reg *ServiceRegistry)
    RegisterRoutes(router chi.Router)
    RegisterTCPHandlers(handler *TCPHandler)
    OnStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}

app := core.NewApp(cfg)
app.Register(auth.NewModule())
app.Register(users.NewModule())
app.Register(chat.NewModule())        // optional
app.Register(leaderboard.NewModule()) // optional
app.Run()
```

Modules communicate through `ServiceRegistry` — they register interfaces and depend on interfaces, not implementations.

### Core Modules (always present)

- **auth**: Login, token issuance, session management, account binding
- **users**: User profiles, status, queries
- **platforms**: Platform adapter registry, OAuth/payment/profile interfaces

### Plugin Modules (optional)

- **chat**: TCP real-time messaging, rooms, message history
- **leaderboard**: Redis Sorted Set ranking, score submission, board config
- **friends**: Friend requests, friend list, block list, online presence
- **achievements**: Achievement definitions, progress, unlock events
- **inventory**: Items, currency, asset ledger
- **payments**: Orders, receipt verification, webhooks
- **admin**: RBAC, audit logs, user management

## Project Directory

```text
cyandie/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── core/
│   │   ├── app.go
│   │   ├── module.go
│   │   ├── registry.go
│   │   ├── config/
│   │   ├── database/
│   │   ├── cache/
│   │   ├── server/
│   │   ├── logger/
│   │   ├── errors/
│   │   └── middleware/
│   ├── auth/
│   ├── users/
│   ├── platforms/
│   ├── chat/
│   ├── leaderboard/
│   ├── friends/
│   ├── achievements/
│   ├── inventory/
│   ├── payments/
│   └── admin/
├── api/
│   └── proto/
│       ├── auth/v1/
│       ├── user/v1/
│       ├── chat/v1/
│       └── leaderboard/v1/
├── sdk/
│   ├── go/
│   └── ts/
├── web/
├── deployments/
│   ├── docker/
│   │   └── Dockerfile
│   └── docker-compose.yml
├── migrations/
│   ├── core/
│   ├── chat/
│   ├── leaderboard/
│   └── friends/
├── scripts/
├── docs/
├── go.mod
├── go.sum
├── buf.gen.yaml
└── sqlc.yaml
```

## Module Interfaces

### Auth

```go
type AuthService interface {
    Login(ctx context.Context, req LoginRequest) (*TokenPair, error)
    Register(ctx context.Context, req RegisterRequest) (UserID, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
    ValidateToken(ctx context.Context, accessToken string) (*Claims, error)
    RevokeSession(ctx context.Context, sessionID string) error
}
```

Login methods: username/password, email OTP, phone OTP, third-party OAuth.

OTP delivery: Auth module defines `OTPNotifier` interface. Implementations inject at startup (e.g. SMTP email, SMS gateway). Framework ships with a log-based notifier for development.

Token strategy: JWT access (15min) + Redis-backed refresh (7d, revocable).

### Users

```go
type UserService interface {
    GetUser(ctx context.Context, id UserID) (*User, error)
    UpdateProfile(ctx context.Context, id UserID, req UpdateProfileRequest) error
    SearchUsers(ctx context.Context, query string, page Pagination) ([]*User, error)
}
```

### Platforms

Platforms use capability-based interfaces, not per-platform monoliths:

```go
type OAuthProvider interface {
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
    GetUserInfo(ctx context.Context, token string) (*PlatformUser, error)
}

type PaymentProvider interface {
    CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error)
    VerifyCallback(ctx context.Context, data CallbackData) error
}
```

A platform provider implements whichever capabilities it supports. Registry lookup: `platforms.GetOAuthProvider("wechat")`.

MVP providers: WeChat (OAuth), username/password (built-in auth).

### Chat (plugin)

- TCP long connection with TLV + Protobuf protocol
- HTTP REST for message history
- Dependencies: UserService, AuthService
- If not registered, TCP server does not start

### Leaderboard (plugin)

- Redis Sorted Set for real-time ranking
- PostgreSQL for board config and score snapshots (historical record, not primary query path)
- Dependencies: UserService

### Friends (plugin)

- Friend requests, friend list, block list
- Online presence via Redis
- Dependencies: UserService, AuthService

## Data Model

### PostgreSQL Core Tables

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(64) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(32) UNIQUE,
    display_name VARCHAR(128),
    avatar_url VARCHAR(512),
    status VARCHAR(16) NOT NULL DEFAULT 'active',  -- active/banned/deleted
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Credentials
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(16) NOT NULL,  -- password/email/phone/oauth
    identifier VARCHAR(255) NOT NULL,
    secret_hash VARCHAR(255),
    verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Platform bindings
CREATE TABLE platform_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    platform VARCHAR(32) NOT NULL,
    platform_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(platform, platform_user_id)
);
```

### PostgreSQL Plugin Tables

```sql
-- Chat rooms (chat plugin)
CREATE TABLE chat_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(16) NOT NULL,  -- private/group/channel
    name VARCHAR(128),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE chat_room_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(16) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(room_id, user_id)
);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id),
    sender_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    type VARCHAR(16) NOT NULL DEFAULT 'text',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Friends (friends plugin)
CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    friend_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending/accepted/blocked
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Leaderboard config (leaderboard plugin)
CREATE TABLE leaderboard_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL,
    update_strategy VARCHAR(16) NOT NULL DEFAULT 'highest',  -- highest/lowest/last
    max_entries INT DEFAULT 1000,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE leaderboard_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL REFERENCES leaderboard_configs(id),
    user_id UUID NOT NULL REFERENCES users(id),
    score DOUBLE PRECISION NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Redis Key Patterns

| Key Pattern | Purpose | TTL |
|-------------|---------|-----|
| `session:{id}` | Refresh token session | 7d |
| `otp:{email\|phone}` | OTP verification code | 5min |
| `leaderboard:{id}` | Sorted Set ranking | per-board |
| `chat:online:{user_id}` | Online presence | heartbeat renewal |
| `cache:user:{id}` | User profile cache | 1h |
| `ratelimit:{endpoint}:{user_id}` | Rate limiting | per-rule |

### Migration Strategy

- goose manages SQL migrations
- Core migrations in `migrations/core/`
- Plugin migrations in `migrations/{module}/`
- Plugin `OnStart()` checks and executes its own migrations

## Protocols

### HTTP REST (chi)

```
/api/v1/auth/login          POST
/api/v1/auth/register       POST
/api/v1/auth/refresh        POST
/api/v1/auth/logout         POST

/api/v1/users/me            GET/PUT
/api/v1/users/{id}          GET
/api/v1/users/search        GET

/api/v1/platforms/{name}/auth-url    GET
/api/v1/platforms/{name}/callback    POST

/api/v1/chat/rooms          GET/POST          # plugin
/api/v1/chat/rooms/{id}     GET
/api/v1/chat/rooms/{id}/messages  GET/POST

/api/v1/leaderboard/{id}    GET              # plugin
/api/v1/leaderboard/{id}/submit  POST

/api/v1/friends             GET/POST          # plugin
/api/v1/friends/{id}/accept  PUT
/api/v1/friends/{id}/reject  DELETE
```

Response envelope:

```json
{
  "ok": true,
  "data": {}
}
```

Error envelope:

```json
{
  "ok": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "requestId": "req_..."
  }
}
```

### gRPC (Protobuf + buf)

- Each module has its own proto package under `api/proto/`
- Used for internal service communication and future microservice split
- buf manages lint + code generation
- Generated code committed to repo

### TCP Protocol (TLV + Protobuf)

Frame format:

```
+--------+--------+---------+
| Type   | Length | Value   |
| 2B     | 4B     | NB      |
+--------+--------+---------+

Type:   uint16, big-endian, message type
Length: uint32, big-endian, value byte count
Value:  Protobuf-encoded message body
```

Message types:

```protobuf
enum MessageType {
  // Connection management
  AUTH          = 0x0001;
  AUTH_OK       = 0x0002;
  AUTH_REJECTED = 0x0003;
  HEARTBEAT     = 0x0010;
  HEARTBEAT_ACK = 0x0011;

  // Chat
  JOIN_ROOM     = 0x0100;
  JOIN_ROOM_OK  = 0x0101;
  LEAVE_ROOM    = 0x0102;
  LEAVE_ROOM_OK = 0x0103;
  SEND_MSG      = 0x0110;
  RECV_MSG      = 0x0111;
  MSG_REJECTED  = 0x0112;

  // Error
  ERROR         = 0xFFFF;
}
```

Connection flow:

1. Client opens TCP connection
2. Client sends AUTH message with access token
3. Server validates token and user status
4. Server replies AUTH_OK or AUTH_REJECTED
5. Authenticated clients exchange chat messages
6. Client sends HEARTBEAT every 30s
7. Server replies HEARTBEAT_ACK
8. Server disconnects after 90s without heartbeat

## Roadmap

### Phase 0: Foundation (2-3 weeks)

- Go project scaffold (cmd/internal/api/sdk directory structure)
- Core framework: App lifecycle, Module interface, ServiceRegistry
- Config system (YAML + env vars)
- PostgreSQL connection + goose migrations
- Redis connection
- Unified error code system
- Structured logging (slog)
- HTTP server (chi) + public middleware
- Docker Compose local dev environment
- Health check endpoint

### Phase 1: Auth + Users (2-3 weeks)

- Username/password registration and login
- Email OTP login
- JWT access + refresh token
- Session management (Redis)
- User profile CRUD
- Auth middleware
- Unit tests + API integration tests

### Phase 2: Platform Adapters (2-3 weeks)

- Platform module core: OAuthProvider interface + registry
- WeChat adapter (OAuth + user info)
- Platform account binding/unbinding
- Platform adapter contract tests

### Phase 3: Leaderboard (1-2 weeks)

- Redis Sorted Set ranking
- Score submission + ranking queries
- Board configuration
- Admin board config API

### Phase 4: Chat (2-3 weeks)

- TCP server + TLV + Protobuf protocol
- Connection auth + heartbeat
- Private + group chat
- Message persistence (PostgreSQL)
- Message history query (HTTP REST)

### Phase 5: Admin + Friends (2-3 weeks)

- Admin RBAC + audit logs
- User ban/unban
- Friend requests/list/block list
- Online presence (Redis)

### Phase 6: SDK + Docs (1-2 weeks)

- Go SDK
- TypeScript SDK
- OpenAPI documentation
- Usage examples

## Module Boundaries

- `auth` calls `users` (create/query) and `platforms` (verify identity). Does not write leaderboard, chat, or inventory data.
- `users` is called by auth, chat, leaderboard, friends, admin. Does not issue tokens or verify platform identity.
- `platforms` provides capability interfaces. Does not create local users, issue JWT, or write business data.
- `chat` calls `users` (query status) and `auth` (validate connections). Does not modify user identity or write leaderboard data.
- `leaderboard` calls `users` (fill user info). Does not verify platform identity or send chat messages.
- `friends` calls `users` and `auth`. Does not send chat messages or write leaderboard data.
- `admin` orchestrates multiple modules through public service interfaces. Must log audit for high-risk operations.
- `core` (shared) does not depend on any business module.

## Security

- Refresh tokens are revocable (Redis-backed)
- Secrets, platform credentials, webhook secrets never in source code
- Platform callbacks must verify signature or origin
- Admin API uses RBAC
- High-risk admin operations record actor, target, reason, before/after values
- User ban affects auth, chat, leaderboard submission
- All public APIs use `/api/v1` prefix and unified error format
- Authenticated endpoints require Auth Guard middleware

## Observability

- Structured logging with request ID / trace ID
- API latency metrics
- TCP connection metrics
- Redis leaderboard operation metrics
- Platform API call success rate and latency
- Admin audit logs