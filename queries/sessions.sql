-- name: CreateSession :one
INSERT INTO user_sessions (user_id, refresh_token_hash, device_id, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM user_sessions WHERE id = $1;

-- name: RevokeSession :one
UPDATE user_sessions SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL
RETURNING *;

-- name: RevokeSessionsByUserID :many
UPDATE user_sessions SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL
RETURNING *;
