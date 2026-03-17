package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type createGateActionParams struct {
	GateID        pgtype.UUID
	ActionType    string
	TransportType string
	Config        []byte
}

func (r *Repository) CreateGateAction(ctx context.Context, arg createGateActionParams) (gateActionRow, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO gate_actions (gate_id, action_type, transport_type, config)
		VALUES ($1, $2, $3, $4)
		RETURNING id, gate_id, action_type, transport_type, config`,
		arg.GateID, arg.ActionType, arg.TransportType, arg.Config,
	)
	var ga gateActionRow
	err := row.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config)
	return ga, mapError(err)
}

type deleteGateActionParams struct {
	GateID     pgtype.UUID
	ActionType string
}

func (r *Repository) DeleteGateAction(ctx context.Context, arg deleteGateActionParams) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM gate_actions WHERE gate_id = $1 AND action_type = $2`,
		arg.GateID, arg.ActionType,
	)
	return mapError(err)
}

type getGateActionParams struct {
	GateID     pgtype.UUID
	ActionType string
}

func (r *Repository) GetGateAction(ctx context.Context, arg getGateActionParams) (gateActionRow, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, gate_id, action_type, transport_type, config
		FROM gate_actions WHERE gate_id = $1 AND action_type = $2`,
		arg.GateID, arg.ActionType,
	)
	var ga gateActionRow
	err := row.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config)
	return ga, mapError(err)
}

func (r *Repository) ListGateActionsByGate(ctx context.Context, gateID pgtype.UUID) ([]gateActionRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, gate_id, action_type, transport_type, config
		FROM gate_actions WHERE gate_id = $1`, gateID,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []gateActionRow
	for rows.Next() {
		var ga gateActionRow
		if err := rows.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config); err != nil {
			return nil, mapError(err)
		}
		items = append(items, ga)
	}
	return items, rows.Err()
}

type updateGateActionParams struct {
	GateID        pgtype.UUID
	ActionType    string
	TransportType string
	Config        []byte
}

func (r *Repository) UpdateGateAction(ctx context.Context, arg updateGateActionParams) (gateActionRow, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE gate_actions SET transport_type = $3, config = $4
		WHERE gate_id = $1 AND action_type = $2
		RETURNING id, gate_id, action_type, transport_type, config`,
		arg.GateID, arg.ActionType, arg.TransportType, arg.Config,
	)
	var ga gateActionRow
	err := row.Scan(&ga.ID, &ga.GateID, &ga.ActionType, &ga.TransportType, &ga.Config)
	return ga, mapError(err)
}
