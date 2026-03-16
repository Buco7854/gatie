package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateGateActionParams struct {
	GateID        pgtype.UUID
	ActionType    string
	TransportType string
	Config        []byte
}

func (q *Queries) CreateGateAction(ctx context.Context, arg CreateGateActionParams) (GateAction, error) {
	row := q.db.QueryRow(ctx,
		`INSERT INTO gate_actions (gate_id, action_type, transport_type, config)
		VALUES ($1, $2, $3, $4)
		RETURNING id, gate_id, action_type, transport_type, config`,
		arg.GateID, arg.ActionType, arg.TransportType, arg.Config,
	)
	var ga GateAction
	err := row.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config)
	return ga, MapError(err)
}

type DeleteGateActionParams struct {
	GateID     pgtype.UUID
	ActionType string
}

func (q *Queries) DeleteGateAction(ctx context.Context, arg DeleteGateActionParams) error {
	_, err := q.db.Exec(ctx,
		`DELETE FROM gate_actions WHERE gate_id = $1 AND action_type = $2`,
		arg.GateID, arg.ActionType,
	)
	return MapError(err)
}

type GetGateActionParams struct {
	GateID     pgtype.UUID
	ActionType string
}

func (q *Queries) GetGateAction(ctx context.Context, arg GetGateActionParams) (GateAction, error) {
	row := q.db.QueryRow(ctx,
		`SELECT id, gate_id, action_type, transport_type, config
		FROM gate_actions WHERE gate_id = $1 AND action_type = $2`,
		arg.GateID, arg.ActionType,
	)
	var ga GateAction
	err := row.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config)
	return ga, MapError(err)
}

func (q *Queries) ListGateActionsByGate(ctx context.Context, gateID pgtype.UUID) ([]GateAction, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, gate_id, action_type, transport_type, config
		FROM gate_actions WHERE gate_id = $1`, gateID,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []GateAction{}
	for rows.Next() {
		var ga GateAction
		if err := rows.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config); err != nil {
			return nil, MapError(err)
		}
		items = append(items, ga)
	}
	return items, rows.Err()
}

type UpdateGateActionParams struct {
	GateID        pgtype.UUID
	ActionType    string
	TransportType string
	Config        []byte
}

func (q *Queries) UpdateGateAction(ctx context.Context, arg UpdateGateActionParams) (GateAction, error) {
	row := q.db.QueryRow(ctx,
		`UPDATE gate_actions SET transport_type = $3, config = $4
		WHERE gate_id = $1 AND action_type = $2
		RETURNING id, gate_id, action_type, transport_type, config`,
		arg.GateID, arg.ActionType, arg.TransportType, arg.Config,
	)
	var ga GateAction
	err := row.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config)
	return ga, MapError(err)
}
