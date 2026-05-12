-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(64) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(32) UNIQUE,
    display_name VARCHAR(128),
    avatar_url VARCHAR(512),
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(16) NOT NULL,
    identifier VARCHAR(255) NOT NULL,
    secret_hash VARCHAR(255),
    verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_credentials_user_id ON credentials(user_id);
CREATE INDEX idx_credentials_identifier ON credentials(identifier);
CREATE UNIQUE INDEX idx_credentials_type_identifier ON credentials(type, identifier);

-- +goose Down
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS users;
