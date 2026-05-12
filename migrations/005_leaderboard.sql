-- +goose Up
CREATE TABLE leaderboard_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL,
    update_strategy VARCHAR(16) NOT NULL DEFAULT 'highest',
    max_entries INT DEFAULT 1000,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE leaderboard_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL REFERENCES leaderboard_configs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score DOUBLE PRECISION NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_scores_board_user ON leaderboard_scores(board_id, user_id);

-- +goose Down
DROP TABLE IF EXISTS leaderboard_scores;
DROP TABLE IF EXISTS leaderboard_configs;
