package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreatePermissionParams struct {
	MemberID      pgtype.UUID
	GateID        pgtype.UUID
	CanOpen       bool
	CanClose      bool
	CanViewStatus bool
	CanManage     bool
}

func (q *Queries) CreatePermission(ctx context.Context, arg CreatePermissionParams) (Permission, error) {
	row := q.db.QueryRow(ctx,
		`INSERT INTO permissions (member_id, gate_id, can_open, can_close, can_view_status, can_manage)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, member_id, gate_id, can_open, can_close, can_view_status, can_manage`,
		arg.MemberID, arg.GateID, arg.CanOpen, arg.CanClose, arg.CanViewStatus, arg.CanManage,
	)
	var p Permission
	err := row.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage)
	return p, MapError(err)
}

type DeletePermissionParams struct {
	MemberID pgtype.UUID
	GateID   pgtype.UUID
}

func (q *Queries) DeletePermission(ctx context.Context, arg DeletePermissionParams) error {
	_, err := q.db.Exec(ctx,
		`DELETE FROM permissions WHERE member_id = $1 AND gate_id = $2`,
		arg.MemberID, arg.GateID,
	)
	return MapError(err)
}

type GetPermissionParams struct {
	MemberID pgtype.UUID
	GateID   pgtype.UUID
}

func (q *Queries) GetPermission(ctx context.Context, arg GetPermissionParams) (Permission, error) {
	row := q.db.QueryRow(ctx,
		`SELECT id, member_id, gate_id, can_open, can_close, can_view_status, can_manage
		FROM permissions WHERE member_id = $1 AND gate_id = $2`,
		arg.MemberID, arg.GateID,
	)
	var p Permission
	err := row.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage)
	return p, MapError(err)
}

func (q *Queries) ListPermissionsByGate(ctx context.Context, gateID pgtype.UUID) ([]Permission, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, member_id, gate_id, can_open, can_close, can_view_status, can_manage
		FROM permissions WHERE gate_id = $1`, gateID,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage); err != nil {
			return nil, MapError(err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (q *Queries) ListPermissionsByMember(ctx context.Context, memberID pgtype.UUID) ([]Permission, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, member_id, gate_id, can_open, can_close, can_view_status, can_manage
		FROM permissions WHERE member_id = $1`, memberID,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage); err != nil {
			return nil, MapError(err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

type UpdatePermissionParams struct {
	MemberID      pgtype.UUID
	GateID        pgtype.UUID
	CanOpen       bool
	CanClose      bool
	CanViewStatus bool
	CanManage     bool
}

func (q *Queries) UpdatePermission(ctx context.Context, arg UpdatePermissionParams) (Permission, error) {
	row := q.db.QueryRow(ctx,
		`UPDATE permissions SET can_open = $3, can_close = $4, can_view_status = $5, can_manage = $6
		WHERE member_id = $1 AND gate_id = $2
		RETURNING id, member_id, gate_id, can_open, can_close, can_view_status, can_manage`,
		arg.MemberID, arg.GateID, arg.CanOpen, arg.CanClose, arg.CanViewStatus, arg.CanManage,
	)
	var p Permission
	err := row.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage)
	return p, MapError(err)
}
