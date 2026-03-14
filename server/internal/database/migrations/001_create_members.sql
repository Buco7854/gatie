-- +goose Up
CREATE TABLE members (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    username      VARCHAR(100) UNIQUE NOT NULL,
    display_name  VARCHAR(200),
    password_hash TEXT NOT NULL,
    role          VARCHAR(10) NOT NULL DEFAULT 'MEMBER' CHECK (role IN ('ADMIN', 'MEMBER')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE members;
