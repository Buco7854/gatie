-- +goose Up
CREATE TABLE gate_actions (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    gate_id         UUID NOT NULL REFERENCES gates(id) ON DELETE CASCADE,
    action_type     VARCHAR(10) NOT NULL CHECK (action_type IN ('OPEN', 'CLOSE', 'STATUS')),
    transport_type  VARCHAR(10) NOT NULL CHECK (transport_type IN ('MQTT', 'HTTP', 'NONE')),
    config          JSONB NOT NULL DEFAULT '{}',
    UNIQUE (gate_id, action_type)
);

-- +goose Down
DROP TABLE gate_actions;
