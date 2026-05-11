# Cyandie Inter-Service Communication Design

Date: 2026-05-11

## Overview

Defines how Cyandie communicates with other internal business services (game servers, match servers, room services, etc.). The design supports bidirectional communication: other services call Cyandie for data and operations, and Cyandie pushes events to other services.

## Key Design Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Synchronous calls | gRPC + Protobuf | Type-safe, high-performance, consistent with existing proto definitions |
| Event push | Redis Pub/Sub | Zero new dependencies (Redis already in stack), sufficient for low-volume push |
| Service auth | Pre-shared service tokens | Simple for small number of services, no mTLS overhead |
| Event format | Unified JSON envelope | Consistent parsing, extensible |

## gRPC Service-to-Service Calls

### Cyandie Exposes (Other Services → Cyandie)

```protobuf
// api/proto/auth/v1/auth.proto
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RevokeSession(RevokeSessionRequest) returns (RevokeSessionResponse);
}

// api/proto/user/v1/user.proto
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc SearchUsers(SearchUsersRequest) returns (SearchUsersResponse);
  rpc UpdateUserStatus(UpdateUserStatusRequest) returns (UpdateUserStatusResponse);
}

// api/proto/leaderboard/v1/leaderboard.proto
service LeaderboardService {
  rpc SubmitScore(SubmitScoreRequest) returns (SubmitScoreResponse);
  rpc GetRanking(GetRankingRequest) returns (GetRankingResponse);
}

// api/proto/chat/v1/chat.proto
service ChatService {
  rpc SendSystemMessage(SendSystemMessageRequest) returns (SendSystemMessageResponse);
}
```

### Cyandie Calls Out (Cyandie → Other Services)

External service addresses are injected via configuration:

```yaml
services:
  game_server:
    addr: "localhost:50051"
  match_server:
    addr: "localhost:50052"
```

Cyandie creates gRPC clients for configured services at startup. Modules resolve clients via ServiceRegistry:

```go
type ExternalServiceClient interface {
    Name() string
    Connect(addr string) error
    Close() error
}
```

### Service Token Authentication

All gRPC inter-service calls carry `authorization: Bearer <service-token>` in gRPC metadata.

Configuration:

```yaml
auth:
  service_tokens:
    - name: "game-server"
      token: "${GAME_SERVER_TOKEN}"
    - name: "match-server"
      token: "${MATCH_SERVER_TOKEN}"
```

Server-side interceptor flow:
1. Extract token from `authorization` metadata
2. Look up token in configured `service_tokens`
3. Valid → write service name into context (`caller: game-server`)
4. Invalid → return `UNAUTHENTICATED` gRPC error

Client-side interceptor automatically attaches the configured service token to outgoing calls.

HTTP API authentication is separate (JWT user tokens). gRPC service auth and HTTP user auth are independent paths.

## Redis Pub/Sub Event Push

### Event Channels

Channel naming: `cyandie:{module}:{event}`

| Channel | Trigger | Payload Fields |
|---------|---------|---------------|
| `cyandie:auth:user_login` | User login success | `user_id`, `session_id`, `ip`, `timestamp` |
| `cyandie:auth:user_logout` | User logout | `user_id`, `session_id`, `timestamp` |
| `cyandie:user:status_changed` | User status change | `user_id`, `old_status`, `new_status`, `timestamp` |
| `cyandie:user:banned` | User banned | `user_id`, `reason`, `operator_id`, `timestamp` |
| `cyandie:leaderboard:score_updated` | Leaderboard score update | `board_id`, `user_id`, `score`, `rank`, `timestamp` |
| `cyandie:chat:system_message` | System message sent | `room_id`, `content`, `type`, `timestamp` |

### Event Envelope

All events use a unified JSON envelope:

```json
{
  "id": "evt_01234567-89ab-cdef",
  "type": "cyandie:auth:user_login",
  "timestamp": "2026-05-11T12:00:00Z",
  "data": {
    "user_id": "uuid",
    "session_id": "uuid",
    "ip": "1.2.3.4"
  }
}
```

### Publishing

Modules publish events through `EventBus`:

```go
type EventBus interface {
    Publish(ctx context.Context, channel string, event Event) error
}

type Event struct {
    ID        string         `json:"id"`
    Type      string         `json:"type"`
    Timestamp time.Time      `json:"timestamp"`
    Data      map[string]any `json:"data"`
}
```

EventBus implementation uses Redis PUBLISH command. The interface allows future migration to Redis Streams (XADD/XREAD) without changing module code.

### Subscribing

Other services subscribe using standard Redis SUBSCRIBE:

```
SUBSCRIBE cyandie:*           -- all events
SUBSCRIBE cyandie:auth:*      -- auth events only
SUBSCRIBE cyandie:user:banned -- specific event
```

### Upgrade Path to Redis Streams

If persistence or replay is needed in the future:

1. Change EventBus implementation from PUBLISH to XADD
2. Consumers switch from SUBSCRIBE to XREADGROUP
3. Module code (calling EventBus.Publish) does not change
4. Add consumer group management to EventBus interface

## Security

- Service tokens are read from environment variables, never committed to source
- gRPC interceptor rejects unauthenticated calls with UNAUTHENTICATED
- Event payloads do not contain secrets (tokens, passwords, hashes)
- Banned user events are published so other services can immediately enforce bans
- Audit log records which service made each gRPC call (via caller context)

## Observability

- gRPC calls logged with caller service name, method, duration, status
- Event publishes logged with channel, event ID
- gRPC interceptor records metrics: call count, latency, error rate per service
- EventBus records metrics: publish count, publish errors per channel
