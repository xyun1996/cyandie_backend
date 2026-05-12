-- name: GetAdminByUsername :one
SELECT * FROM admin_users WHERE username = $1;

-- name: CreateAdminUser :one
INSERT INTO admin_users (username, password_hash, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (operator_id, action, target_type, target_id, before_value, after_value, reason, ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2;
