-- +goose Up
CREATE TABLE block_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_block_unique ON block_relations(blocker_id, blocked_id);
CREATE INDEX idx_block_blocked ON block_relations(blocked_id);

-- +goose Down
DROP TABLE IF EXISTS block_relations;
