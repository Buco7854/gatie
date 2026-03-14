-- name: CreatePermission :one
INSERT INTO permissions (member_id, gate_id, can_open, can_close, can_view_status, can_manage)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions
WHERE member_id = $1 AND gate_id = $2;

-- name: ListPermissionsByMember :many
SELECT * FROM permissions
WHERE member_id = $1;

-- name: ListPermissionsByGate :many
SELECT * FROM permissions
WHERE gate_id = $1;

-- name: UpdatePermission :one
UPDATE permissions
SET can_open = $3, can_close = $4, can_view_status = $5, can_manage = $6
WHERE member_id = $1 AND gate_id = $2
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE member_id = $1 AND gate_id = $2;
