-- name: CreateChatRoom :one
INSERT INTO chat_rooms (type, name, metadata)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetChatRoom :one
SELECT * FROM chat_rooms WHERE id = $1;

-- name: AddRoomMember :one
INSERT INTO chat_room_members (room_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: RemoveRoomMember :one
DELETE FROM chat_room_members WHERE room_id = $1 AND user_id = $2
RETURNING *;

-- name: GetRoomMembers :many
SELECT * FROM chat_room_members WHERE room_id = $1;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (room_id, sender_id, content, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetChatMessages :many
SELECT * FROM chat_messages
WHERE room_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListRoomsByUser :many
SELECT cr.id, cr.type, cr.name, cr.metadata, cr.created_at
FROM chat_rooms cr
JOIN chat_room_members crm ON cr.id = crm.room_id
WHERE crm.user_id = $1
ORDER BY cr.created_at DESC;
