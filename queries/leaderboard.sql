-- name: GetLeaderboardConfig :one
SELECT * FROM leaderboard_configs WHERE code = $1;

-- name: ListLeaderboardConfigs :many
SELECT * FROM leaderboard_configs ORDER BY name;

-- name: CreateLeaderboardConfig :one
INSERT INTO leaderboard_configs (code, name, update_strategy, max_entries, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateScore :one
INSERT INTO leaderboard_scores (board_id, user_id, score)
VALUES ($1, $2, $3)
RETURNING *;
