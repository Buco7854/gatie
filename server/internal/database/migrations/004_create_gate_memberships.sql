-- +goose Up
CREATE TABLE gate_memberships (
    gate_id    UUID NOT NULL REFERENCES gates(id) ON DELETE CASCADE,
    member_id  UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    role_id    VARCHAR(20) NOT NULL REFERENCES roles(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (gate_id, member_id)
);

CREATE INDEX idx_gate_memberships_member_id ON gate_memberships(member_id);

-- +goose Down
DROP TABLE gate_memberships;
