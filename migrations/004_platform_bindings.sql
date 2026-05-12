-- +goose Up
CREATE TABLE platform_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(32) NOT NULL,
    platform_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL DEFAULT '',
    refresh_token TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_platform_bindings_platform_user ON platform_bindings(platform, platform_user_id);
CREATE INDEX idx_platform_bindings_user_id ON platform_bindings(user_id);

-- +goose Down
DROP TABLE IF EXISTS platform_bindings;
