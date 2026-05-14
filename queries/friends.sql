-- name: CreateFriendship :one
INSERT INTO friendships (user_id, friend_id, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFriendship :one
SELECT * FROM friendships WHERE id = $1;

-- name: GetFriendshipByUsers :one
SELECT * FROM friendships
WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1);

-- name: UpdateFriendshipStatus :one
UPDATE friendships SET status = $1, updated_at = now()
WHERE id = $2
RETURNING *;

-- name: ListFriends :many
SELECT * FROM friendships
WHERE (user_id = $1 OR friend_id = $1) AND status = 'accepted'
ORDER BY updated_at DESC;

-- name: ListPendingRequests :many
SELECT * FROM friendships
WHERE friend_id = $1 AND status = 'pending'
ORDER BY created_at DESC;

-- name: DeleteFriendship :one
DELETE FROM friendships WHERE id = $1 RETURNING *;

-- name: CreateBlockRelation :one
INSERT INTO block_relations (blocker_id, blocked_id, reason) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteBlockRelation :one
DELETE FROM block_relations WHERE blocker_id = $1 AND blocked_id = $2 RETURNING *;

-- name: ListBlockedUsers :many
SELECT * FROM block_relations WHERE blocker_id = $1 ORDER BY created_at DESC;

-- name: IsBlockedBy :one
SELECT id FROM block_relations WHERE blocker_id = $1 AND blocked_id = $2;

-- name: DeleteFriendshipByUsers :one
DELETE FROM friendships WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1) RETURNING *;
