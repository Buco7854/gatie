-- +goose Up

-- Missing indexes on foreign keys for JOIN performance
CREATE INDEX idx_gate_actions_gate_id ON gate_actions(gate_id);

-- Auto-update updated_at on row modification
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_members_updated_at
    BEFORE UPDATE ON members
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_gates_updated_at
    BEFORE UPDATE ON gates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_gates_updated_at ON gates;
DROP TRIGGER IF EXISTS trg_members_updated_at ON members;
DROP FUNCTION IF EXISTS set_updated_at();
DROP INDEX IF EXISTS idx_gate_actions_gate_id;
