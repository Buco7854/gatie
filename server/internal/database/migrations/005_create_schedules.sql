-- +goose Up
CREATE TABLE schedules (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL,
    scope       VARCHAR(10) NOT NULL CHECK (scope IN ('ADMIN', 'PERSONAL')),
    owner_id    UUID REFERENCES members(id) ON DELETE CASCADE,
    expression  JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT schedules_personal_requires_owner
        CHECK (scope = 'ADMIN' OR owner_id IS NOT NULL)
);

CREATE TABLE member_gate_schedules (
    member_id   UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    gate_id     UUID NOT NULL REFERENCES gates(id) ON DELETE CASCADE,
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    PRIMARY KEY (member_id, gate_id, schedule_id)
);

-- +goose Down
DROP TABLE member_gate_schedules;
DROP TABLE schedules;
