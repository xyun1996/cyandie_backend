-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: SearchUsers :many
SELECT * FROM users
WHERE (username ILIKE CAST($1 AS varchar) OR display_name ILIKE CAST($1 AS varchar))
AND status = 'active'
ORDER BY username
LIMIT $2 OFFSET $3;

-- name: CreateUser :one
INSERT INTO users (username, email, display_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET display_name = COALESCE(sqlc.arg(display_name), display_name),
    avatar_url = COALESCE(sqlc.arg(avatar_url), avatar_url),
    metadata = COALESCE(sqlc.arg(metadata), metadata),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users SET status = sqlc.arg(status), updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;
