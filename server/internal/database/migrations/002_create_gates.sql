-- +goose Up
CREATE TABLE gates (
    id                UUID PRIMARY KEY DEFAULT uuidv7(),
    name              VARCHAR(100) NOT NULL,
    gate_token_hash   TEXT NOT NULL,
    status_ttl_seconds INT NOT NULL DEFAULT 60,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE gates;
