-- +goose Up
CREATE TABLE refresh_tokens (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    member_id     UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    token_hash    TEXT NOT NULL UNIQUE,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_member_id ON refresh_tokens(member_id);

-- +goose Down
DROP TABLE refresh_tokens;
