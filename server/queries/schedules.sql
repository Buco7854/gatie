-- name: CreateSchedule :one
INSERT INTO schedules (name, scope, owner_id, expression)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetScheduleByID :one
SELECT * FROM schedules
WHERE id = $1;

-- name: ListAdminSchedules :many
SELECT * FROM schedules
WHERE scope = 'ADMIN'
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: ListPersonalSchedules :many
SELECT * FROM schedules
WHERE scope = 'PERSONAL' AND owner_id = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: UpdateSchedule :one
UPDATE schedules
SET name = $2, expression = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM schedules WHERE id = $1;

-- name: AttachScheduleToMemberGate :exec
INSERT INTO member_gate_schedules (member_id, gate_id, schedule_id)
VALUES ($1, $2, $3);

-- name: DetachScheduleFromMemberGate :exec
DELETE FROM member_gate_schedules
WHERE member_id = $1 AND gate_id = $2 AND schedule_id = $3;

-- name: ListSchedulesForMemberGate :many
SELECT s.* FROM schedules s
JOIN member_gate_schedules mgs ON mgs.schedule_id = s.id
WHERE mgs.member_id = $1 AND mgs.gate_id = $2;
