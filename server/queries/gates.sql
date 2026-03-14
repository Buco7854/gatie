-- name: CreateGate :one
INSERT INTO gates (name, gate_token_hash, status_ttl_seconds)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetGateByID :one
SELECT * FROM gates
WHERE id = $1;

-- name: ListGates :many
SELECT * FROM gates
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: CountGates :one
SELECT count(*) FROM gates;

-- name: UpdateGate :one
UPDATE gates
SET name = $2, status_ttl_seconds = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateGateToken :one
UPDATE gates
SET gate_token_hash = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteGate :exec
DELETE FROM gates WHERE id = $1;
