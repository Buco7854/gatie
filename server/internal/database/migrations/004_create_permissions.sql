-- +goose Up
CREATE TABLE permissions (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    member_id       UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    gate_id         UUID NOT NULL REFERENCES gates(id) ON DELETE CASCADE,
    can_open        BOOLEAN NOT NULL DEFAULT false,
    can_close       BOOLEAN NOT NULL DEFAULT false,
    can_view_status BOOLEAN NOT NULL DEFAULT false,
    can_manage      BOOLEAN NOT NULL DEFAULT false,
    UNIQUE (member_id, gate_id)
);

-- +goose Down
DROP TABLE permissions;
