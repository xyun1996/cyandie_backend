# Phase 5: Admin + Friends Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement Admin module (RBAC, audit logs, user management) and Friends module (friend requests, friend list, online presence).

**Architecture:** Admin is a core module with independent user system (admin_users table). Friends is a plugin module. Both register via Module interface.

**Tech Stack:** Go, chi v5, sqlc, Redis, bcrypt, slog

---

### Task 1: Database Migrations + sqlc

- Create `migrations/006_admin_friends.sql`
- Create `queries/admin.sql` and `queries/friends.sql`
- Regenerate sqlc

Migration:
```sql
-- +goose Up
CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'admin',
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id UUID REFERENCES admin_users(id),
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id VARCHAR(255) NOT NULL,
    before_value JSONB,
    after_value JSONB,
    reason TEXT,
    ip VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_friendships_user ON friendships(user_id);
CREATE INDEX idx_friendships_friend ON friendships(friend_id);

-- +goose Down
DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS admin_users;
```

Commit: `git add migrations/ queries/ internal/db/ && git commit -m "feat: add admin and friends migrations and sqlc queries"`

---

### Task 2: Admin Service + Handlers + Module

Create `internal/admin/` with:
- `service.go` — AdminService (login, ban/unban, list users, audit log)
- `handler.go` — HTTP handlers under `/api/v1/admin/`
- `module.go` — Module interface

Admin login uses bcrypt + JWT (separate from user JWT). Admin endpoints require admin auth middleware.

Key endpoints:
```
POST /api/v1/admin/login
GET  /api/v1/admin/users
PUT  /api/v1/admin/users/{id}/status
GET  /api/v1/admin/audit-logs
```

Commit: `git add internal/admin/ && git commit -m "feat: add Admin module with RBAC and audit logs"`

---

### Task 3: Friends Service + Handlers + Module

Create `internal/friends/` with:
- `service.go` — FriendsService (send request, accept/reject, list friends, online presence)
- `handler.go` — HTTP handlers under `/api/v1/friends`
- `module.go` — Module interface

Key endpoints:
```
POST   /api/v1/friends/request
PUT    /api/v1/friends/{id}/accept
DELETE /api/v1/friends/{id}/reject
GET    /api/v1/friends
GET    /api/v1/friends/online
```

Online presence tracked in Redis: `friend:online:{user_id}` with heartbeat TTL.

Commit: `git add internal/friends/ && git commit -m "feat: add Friends module with requests and online presence"`

---

### Task 4: Wire into main.go

Update `cmd/server/main.go` to register admin and friends modules with appropriate rate limiting.

Commit: `git add cmd/server/main.go && git commit -m "feat: wire admin and friends modules into app"`
