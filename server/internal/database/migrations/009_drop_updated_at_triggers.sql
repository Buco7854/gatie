-- +goose Up

-- Remove auto-update triggers: updated_at is set explicitly in SQL queries.
DROP TRIGGER IF EXISTS trg_members_updated_at ON members;
DROP TRIGGER IF EXISTS trg_gates_updated_at ON gates;
DROP TRIGGER IF EXISTS trg_schedules_updated_at ON schedules;
DROP FUNCTION IF EXISTS set_updated_at();

-- Fix index direction: ListGates queries ORDER BY created_at ASC.
DROP INDEX IF EXISTS idx_gates_created_at;
CREATE INDEX idx_gates_created_at ON gates(created_at ASC);

-- +goose Down

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

CREATE TRIGGER trg_schedules_updated_at
    BEFORE UPDATE ON schedules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP INDEX IF EXISTS idx_gates_created_at;
CREATE INDEX idx_gates_created_at ON gates(created_at DESC);
