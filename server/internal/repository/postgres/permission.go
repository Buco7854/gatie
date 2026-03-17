package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type createPermissionParams struct {
	MemberID      pgtype.UUID
	GateID        pgtype.UUID
	CanOpen       bool
	CanClose      bool
	CanViewStatus bool
	CanManage     bool
}

func (r *Repository) CreatePermission(ctx context.Context, arg createPermissionParams) (permissionRow, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO permissions (member_id, gate_id, can_open, can_close, can_view_status, can_manage)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, member_id, gate_id, can_open, can_close, can_view_status, can_manage`,
		arg.MemberID, arg.GateID, arg.CanOpen, arg.CanClose, arg.CanViewStatus, arg.CanManage,
	)
	var p permissionRow
	err := row.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage)
	return p, mapError(err)
}

type deletePermissionParams struct {
	MemberID pgtype.UUID
	GateID   pgtype.UUID
}

func (r *Repository) DeletePermission(ctx context.Context, arg deletePermissionParams) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM permissions WHERE member_id = $1 AND gate_id = $2`,
		arg.MemberID, arg.GateID,
	)
	return mapError(err)
}

type getPermissionParams struct {
	MemberID pgtype.UUID
	GateID   pgtype.UUID
}

func (r *Repository) GetPermission(ctx context.Context, arg getPermissionParams) (permissionRow, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, member_id, gate_id, can_open, can_close, can_view_status, can_manage
		FROM permissions WHERE member_id = $1 AND gate_id = $2`,
		arg.MemberID, arg.GateID,
	)
	var p permissionRow
	err := row.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage)
	return p, mapError(err)
}

func (r *Repository) ListPermissionsByGate(ctx context.Context, gateID pgtype.UUID) ([]permissionRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, member_id, gate_id, can_open, can_close, can_view_status, can_manage
		FROM permissions WHERE gate_id = $1`, gateID,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []permissionRow
	for rows.Next() {
		var p permissionRow
		if err := rows.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage); err != nil {
			return nil, mapError(err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *Repository) ListPermissionsByMember(ctx context.Context, memberID pgtype.UUID) ([]permissionRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, member_id, gate_id, can_open, can_close, can_view_status, can_manage
		FROM permissions WHERE member_id = $1`, memberID,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []permissionRow
	for rows.Next() {
		var p permissionRow
		if err := rows.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage); err != nil {
			return nil, mapError(err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

type updatePermissionParams struct {
	MemberID      pgtype.UUID
	GateID        pgtype.UUID
	CanOpen       bool
	CanClose      bool
	CanViewStatus bool
	CanManage     bool
}

func (r *Repository) UpdatePermission(ctx context.Context, arg updatePermissionParams) (permissionRow, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE permissions SET can_open = $3, can_close = $4, can_view_status = $5, can_manage = $6
		WHERE member_id = $1 AND gate_id = $2
		RETURNING id, member_id, gate_id, can_open, can_close, can_view_status, can_manage`,
		arg.MemberID, arg.GateID, arg.CanOpen, arg.CanClose, arg.CanViewStatus, arg.CanManage,
	)
	var p permissionRow
	err := row.Scan(&p.ID, &p.MemberID, &p.GateID, &p.CanOpen, &p.CanClose, &p.CanViewStatus, &p.CanManage)
	return p, mapError(err)
}
