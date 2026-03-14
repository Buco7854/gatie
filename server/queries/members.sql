-- name: CreateMember :one
INSERT INTO members (username, display_name, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMemberByUsername :one
SELECT * FROM members
WHERE username = $1;

-- name: GetMemberByID :one
SELECT * FROM members
WHERE id = $1;

-- name: CountMembers :one
SELECT count(*) FROM members;

-- name: ListMembers :many
SELECT * FROM members
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: UpdateMember :one
UPDATE members
SET username = $2, display_name = $3, role = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMember :exec
DELETE FROM members WHERE id = $1;
