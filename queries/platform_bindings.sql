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
