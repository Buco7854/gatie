package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type attachScheduleToMemberGateParams struct {
	MemberID   pgtype.UUID
	GateID     pgtype.UUID
	ScheduleID pgtype.UUID
}

func (r *Repository) AttachScheduleToMemberGate(ctx context.Context, arg attachScheduleToMemberGateParams) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO member_gate_schedules (member_id, gate_id, schedule_id)
		VALUES ($1, $2, $3)`,
		arg.MemberID, arg.GateID, arg.ScheduleID,
	)
	return mapError(err)
}

type createScheduleParams struct {
	Name       string
	Scope      string
	OwnerID    pgtype.UUID
	Expression []byte
}

func (r *Repository) CreateSchedule(ctx context.Context, arg createScheduleParams) (scheduleRow, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO schedules (name, scope, owner_id, expression)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, scope, owner_id, expression, created_at, updated_at`,
		arg.Name, arg.Scope, arg.OwnerID, arg.Expression,
	)
	var s scheduleRow
	err := row.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt)
	return s, mapError(err)
}

func (r *Repository) DeleteSchedule(ctx context.Context, id pgtype.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	return mapError(err)
}

type detachScheduleFromMemberGateParams struct {
	MemberID   pgtype.UUID
	GateID     pgtype.UUID
	ScheduleID pgtype.UUID
}

func (r *Repository) DetachScheduleFromMemberGate(ctx context.Context, arg detachScheduleFromMemberGateParams) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM member_gate_schedules
		WHERE member_id = $1 AND gate_id = $2 AND schedule_id = $3`,
		arg.MemberID, arg.GateID, arg.ScheduleID,
	)
	return mapError(err)
}

func (r *Repository) GetScheduleByID(ctx context.Context, id pgtype.UUID) (scheduleRow, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, scope, owner_id, expression, created_at, updated_at
		FROM schedules WHERE id = $1`, id,
	)
	var s scheduleRow
	err := row.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt)
	return s, mapError(err)
}

type listAdminSchedulesParams struct {
	Limit  int32
	Offset int32
}

func (r *Repository) ListAdminSchedules(ctx context.Context, arg listAdminSchedulesParams) ([]scheduleRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, scope, owner_id, expression, created_at, updated_at
		FROM schedules WHERE scope = 'ADMIN'
		ORDER BY created_at ASC LIMIT $1 OFFSET $2`,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []scheduleRow
	for rows.Next() {
		var s scheduleRow
		if err := rows.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, mapError(err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

type listPersonalSchedulesParams struct {
	OwnerID pgtype.UUID
	Limit   int32
	Offset  int32
}

func (r *Repository) ListPersonalSchedules(ctx context.Context, arg listPersonalSchedulesParams) ([]scheduleRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, scope, owner_id, expression, created_at, updated_at
		FROM schedules WHERE scope = 'PERSONAL' AND owner_id = $1
		ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		arg.OwnerID, arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []scheduleRow
	for rows.Next() {
		var s scheduleRow
		if err := rows.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, mapError(err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

type listSchedulesForMemberGateParams struct {
	MemberID pgtype.UUID
	GateID   pgtype.UUID
}

func (r *Repository) ListSchedulesForMemberGate(ctx context.Context, arg listSchedulesForMemberGateParams) ([]scheduleRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.name, s.scope, s.owner_id, s.expression, s.created_at, s.updated_at
		FROM schedules s
		JOIN member_gate_schedules mgs ON mgs.schedule_id = s.id
		WHERE mgs.member_id = $1 AND mgs.gate_id = $2`,
		arg.MemberID, arg.GateID,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []scheduleRow
	for rows.Next() {
		var s scheduleRow
		if err := rows.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, mapError(err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

type updateScheduleParams struct {
	ID         pgtype.UUID
	Name       string
	Expression []byte
}

func (r *Repository) UpdateSchedule(ctx context.Context, arg updateScheduleParams) (scheduleRow, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE schedules SET name = $2, expression = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, scope, owner_id, expression, created_at, updated_at`,
		arg.ID, arg.Name, arg.Expression,
	)
	var s scheduleRow
	err := row.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt)
	return s, mapError(err)
}
