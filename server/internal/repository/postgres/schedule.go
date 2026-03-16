package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type AttachScheduleToMemberGateParams struct {
	MemberID   pgtype.UUID
	GateID     pgtype.UUID
	ScheduleID pgtype.UUID
}

func (q *Queries) AttachScheduleToMemberGate(ctx context.Context, arg AttachScheduleToMemberGateParams) error {
	_, err := q.db.Exec(ctx,
		`INSERT INTO member_gate_schedules (member_id, gate_id, schedule_id)
		VALUES ($1, $2, $3)`,
		arg.MemberID, arg.GateID, arg.ScheduleID,
	)
	return MapError(err)
}

type CreateScheduleParams struct {
	Name       string
	Scope      string
	OwnerID    pgtype.UUID
	Expression []byte
}

func (q *Queries) CreateSchedule(ctx context.Context, arg CreateScheduleParams) (Schedule, error) {
	row := q.db.QueryRow(ctx,
		`INSERT INTO schedules (name, scope, owner_id, expression)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, scope, owner_id, expression, created_at, updated_at`,
		arg.Name, arg.Scope, arg.OwnerID, arg.Expression,
	)
	var s Schedule
	err := row.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt)
	return s, MapError(err)
}

func (q *Queries) DeleteSchedule(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	return MapError(err)
}

type DetachScheduleFromMemberGateParams struct {
	MemberID   pgtype.UUID
	GateID     pgtype.UUID
	ScheduleID pgtype.UUID
}

func (q *Queries) DetachScheduleFromMemberGate(ctx context.Context, arg DetachScheduleFromMemberGateParams) error {
	_, err := q.db.Exec(ctx,
		`DELETE FROM member_gate_schedules
		WHERE member_id = $1 AND gate_id = $2 AND schedule_id = $3`,
		arg.MemberID, arg.GateID, arg.ScheduleID,
	)
	return MapError(err)
}

func (q *Queries) GetScheduleByID(ctx context.Context, id pgtype.UUID) (Schedule, error) {
	row := q.db.QueryRow(ctx,
		`SELECT id, name, scope, owner_id, expression, created_at, updated_at
		FROM schedules WHERE id = $1`, id,
	)
	var s Schedule
	err := row.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt)
	return s, MapError(err)
}

type ListAdminSchedulesParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListAdminSchedules(ctx context.Context, arg ListAdminSchedulesParams) ([]Schedule, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, name, scope, owner_id, expression, created_at, updated_at
		FROM schedules WHERE scope = 'ADMIN'
		ORDER BY created_at ASC LIMIT $1 OFFSET $2`,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []Schedule{}
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

type ListPersonalSchedulesParams struct {
	OwnerID pgtype.UUID
	Limit   int32
	Offset  int32
}

func (q *Queries) ListPersonalSchedules(ctx context.Context, arg ListPersonalSchedulesParams) ([]Schedule, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, name, scope, owner_id, expression, created_at, updated_at
		FROM schedules WHERE scope = 'PERSONAL' AND owner_id = $1
		ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		arg.OwnerID, arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []Schedule{}
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

type ListSchedulesForMemberGateParams struct {
	MemberID pgtype.UUID
	GateID   pgtype.UUID
}

func (q *Queries) ListSchedulesForMemberGate(ctx context.Context, arg ListSchedulesForMemberGateParams) ([]Schedule, error) {
	rows, err := q.db.Query(ctx,
		`SELECT s.id, s.name, s.scope, s.owner_id, s.expression, s.created_at, s.updated_at
		FROM schedules s
		JOIN member_gate_schedules mgs ON mgs.schedule_id = s.id
		WHERE mgs.member_id = $1 AND mgs.gate_id = $2`,
		arg.MemberID, arg.GateID,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []Schedule{}
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

type UpdateScheduleParams struct {
	ID         pgtype.UUID
	Name       string
	Expression []byte
}

func (q *Queries) UpdateSchedule(ctx context.Context, arg UpdateScheduleParams) (Schedule, error) {
	row := q.db.QueryRow(ctx,
		`UPDATE schedules SET name = $2, expression = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, scope, owner_id, expression, created_at, updated_at`,
		arg.ID, arg.Name, arg.Expression,
	)
	var s Schedule
	err := row.Scan(&s.ID, &s.Name, &s.Scope, &s.OwnerID, &s.Expression, &s.CreatedAt, &s.UpdatedAt)
	return s, MapError(err)
}
