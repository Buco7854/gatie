-- +goose Up

-- Missing index: gate_memberships by gate_id (FK lookup)
CREATE INDEX idx_gate_memberships_gate_id ON gate_memberships(gate_id);

-- Missing index: gates pagination by created_at
CREATE INDEX idx_gates_created_at ON gates(created_at DESC);

-- Missing index: schedules by owner_id (FK lookup)
CREATE INDEX idx_schedules_owner_id ON schedules(owner_id);

-- Missing index: member_gate_schedules component columns (FK lookup for JOINs)
CREATE INDEX idx_member_gate_schedules_gate_id ON member_gate_schedules(gate_id);
CREATE INDEX idx_member_gate_schedules_schedule_id ON member_gate_schedules(schedule_id);

-- Missing index: refresh_tokens by expires_at (cleanup queries)
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Missing CHECK constraint: gates.status_ttl_seconds must be positive
ALTER TABLE gates ADD CONSTRAINT chk_gates_status_ttl_seconds
    CHECK (status_ttl_seconds >= 1 AND status_ttl_seconds <= 86400);

-- Missing trigger: auto-update schedules.updated_at
CREATE TRIGGER trg_schedules_updated_at
    BEFORE UPDATE ON schedules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_schedules_updated_at ON schedules;
ALTER TABLE gates DROP CONSTRAINT IF EXISTS chk_gates_status_ttl_seconds;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_member_gate_schedules_schedule_id;
DROP INDEX IF EXISTS idx_member_gate_schedules_gate_id;
DROP INDEX IF EXISTS idx_schedules_owner_id;
DROP INDEX IF EXISTS idx_gates_created_at;
DROP INDEX IF EXISTS idx_gate_memberships_gate_id;
