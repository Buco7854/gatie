-- name: CreateGateAction :one
INSERT INTO gate_actions (gate_id, action_type, transport_type, config)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGateAction :one
SELECT * FROM gate_actions
WHERE gate_id = $1 AND action_type = $2;

-- name: ListGateActionsByGate :many
SELECT * FROM gate_actions
WHERE gate_id = $1;

-- name: UpdateGateAction :one
UPDATE gate_actions
SET transport_type = $3, config = $4
WHERE gate_id = $1 AND action_type = $2
RETURNING *;

-- name: DeleteGateAction :exec
DELETE FROM gate_actions
WHERE gate_id = $1 AND action_type = $2;
