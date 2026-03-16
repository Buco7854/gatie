package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

func (q *Queries) CountMembers(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, `SELECT count(*) FROM members`)
	var count int64
	err := row.Scan(&count)
	return count, MapError(err)
}

func (q *Queries) CountMembersByRole(ctx context.Context, role string) (int64, error) {
	row := q.db.QueryRow(ctx, `SELECT count(*) FROM members WHERE role = $1`, role)
	var count int64
	err := row.Scan(&count)
	return count, MapError(err)
}

type CreateMemberParams struct {
	Username     string
	DisplayName  pgtype.Text
	PasswordHash string
	Role         string
}

func (q *Queries) CreateMember(ctx context.Context, arg CreateMemberParams) (Member, error) {
	row := q.db.QueryRow(ctx,
		`INSERT INTO members (username, display_name, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, display_name, password_hash, role, created_at, updated_at`,
		arg.Username, arg.DisplayName, arg.PasswordHash, arg.Role,
	)
	var m Member
	err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	return m, MapError(err)
}

func (q *Queries) DeleteMember(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, `DELETE FROM members WHERE id = $1`, id)
	return MapError(err)
}

func (q *Queries) GetMemberByID(ctx context.Context, id pgtype.UUID) (Member, error) {
	row := q.db.QueryRow(ctx,
		`SELECT id, username, display_name, password_hash, role, created_at, updated_at
		FROM members WHERE id = $1`, id,
	)
	var m Member
	err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	return m, MapError(err)
}

func (q *Queries) GetMemberByUsername(ctx context.Context, username string) (Member, error) {
	row := q.db.QueryRow(ctx,
		`SELECT id, username, display_name, password_hash, role, created_at, updated_at
		FROM members WHERE username = $1`, username,
	)
	var m Member
	err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	return m, MapError(err)
}

type ListMembersParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListMembers(ctx context.Context, arg ListMembersParams) ([]Member, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, username, display_name, password_hash, role, created_at, updated_at
		FROM members ORDER BY created_at ASC LIMIT $1 OFFSET $2`,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.Role, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

type UpdateMemberParams struct {
	ID          pgtype.UUID
	Username    string
	DisplayName pgtype.Text
	Role        string
}

func (q *Queries) UpdateMember(ctx context.Context, arg UpdateMemberParams) (Member, error) {
	row := q.db.QueryRow(ctx,
		`UPDATE members SET username = $2, display_name = $3, role = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, username, display_name, password_hash, role, created_at, updated_at`,
		arg.ID, arg.Username, arg.DisplayName, arg.Role,
	)
	var m Member
	err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	return m, MapError(err)
}
