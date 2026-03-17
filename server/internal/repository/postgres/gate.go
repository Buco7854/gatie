package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

func (q *Queries) CountGates(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, `SELECT count(*) FROM gates`)
	var count int64
	err := row.Scan(&count)
	return count, MapError(err)
}

type CreateGateParams struct {
	Name             string
	GateTokenHash    string
	StatusTtlSeconds int32
}

func (q *Queries) CreateGate(ctx context.Context, arg CreateGateParams) (Gate, error) {
	row := q.db.QueryRow(ctx,
		`INSERT INTO gates (name, gate_token_hash, status_ttl_seconds)
		VALUES ($1, $2, $3)
		RETURNING id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at`,
		arg.Name, arg.GateTokenHash, arg.StatusTtlSeconds,
	)
	var g Gate
	err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt)
	return g, MapError(err)
}

func (q *Queries) DeleteGate(ctx context.Context, id pgtype.UUID) error {
	row := q.db.QueryRow(ctx, `DELETE FROM gates WHERE id = $1 RETURNING id`, id)
	var deleted pgtype.UUID
	return MapError(row.Scan(&deleted))
}

func (q *Queries) GetGateByID(ctx context.Context, id pgtype.UUID) (Gate, error) {
	row := q.db.QueryRow(ctx,
		`SELECT id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at
		FROM gates WHERE id = $1`, id,
	)
	var g Gate
	err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt)
	return g, MapError(err)
}

type ListGatesParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListGates(ctx context.Context, arg ListGatesParams) ([]Gate, error) {
	rows, err := q.db.Query(ctx,
		`SELECT id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at
		FROM gates ORDER BY created_at ASC LIMIT $1 OFFSET $2`,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	items := []Gate{}
	for rows.Next() {
		var g Gate
		if err := rows.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		items = append(items, g)
	}
	return items, rows.Err()
}

type UpdateGateParams struct {
	ID               pgtype.UUID
	Name             string
	StatusTtlSeconds int32
}

func (q *Queries) UpdateGate(ctx context.Context, arg UpdateGateParams) (Gate, error) {
	row := q.db.QueryRow(ctx,
		`UPDATE gates SET name = $2, status_ttl_seconds = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at`,
		arg.ID, arg.Name, arg.StatusTtlSeconds,
	)
	var g Gate
	err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt)
	return g, MapError(err)
}

type UpdateGateTokenParams struct {
	ID            pgtype.UUID
	GateTokenHash string
}

func (q *Queries) UpdateGateToken(ctx context.Context, arg UpdateGateTokenParams) (Gate, error) {
	row := q.db.QueryRow(ctx,
		`UPDATE gates SET gate_token_hash = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at`,
		arg.ID, arg.GateTokenHash,
	)
	var g Gate
	err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt)
	return g, MapError(err)
}
