-- name: GetCredentialByTypeIdentifier :one
SELECT * FROM credentials WHERE type = $1 AND identifier = $2;

-- name: CreateCredential :one
INSERT INTO credentials (user_id, type, identifier, secret_hash, verified)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCredentialsByUserID :many
SELECT * FROM credentials WHERE user_id = $1;
