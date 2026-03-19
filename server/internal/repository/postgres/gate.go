package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/repository"
)

type GateRepository struct{ base }

func NewGateRepository(pool *pgxpool.Pool) *GateRepository {
	return &GateRepository{base{db: pool, pool: pool}}
}

func (r *GateRepository) BeginTx(ctx context.Context) (repository.GateRepository, error) {
	b, err := r.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	return &GateRepository{b}, nil
}

func (r *GateRepository) CountGates(ctx context.Context) (int64, error) {
	row := r.db.QueryRow(ctx, `SELECT count(*) FROM gates`)
	var count int64
	err := row.Scan(&count)
	return count, mapError(err)
}

func (r *GateRepository) ListGates(ctx context.Context, arg repository.ListParams) ([]repository.Gate, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at
		FROM gates ORDER BY created_at ASC LIMIT $1 OFFSET $2`,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var out []repository.Gate
	for rows.Next() {
		var g gateRow
		if err := rows.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, mapError(err)
		}
		out = append(out, toRepoGate(g))
	}
	return out, rows.Err()
}

func (r *GateRepository) GetGateByID(ctx context.Context, id string) (repository.Gate, error) {
	return queryGateByID(ctx, r.db, id)
}

func (r *GateRepository) CreateGate(ctx context.Context, arg repository.CreateGateParams) (repository.Gate, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO gates (name, gate_token_hash, status_ttl_seconds)
		VALUES ($1, $2, $3)
		RETURNING id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at`,
		arg.Name, arg.GateTokenHash, arg.StatusTtlSeconds,
	)
	var g gateRow
	if err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return repository.Gate{}, mapError(err)
	}
	return toRepoGate(g), nil
}

func (r *GateRepository) PatchGate(ctx context.Context, arg repository.PatchGateParams) (repository.Gate, error) {
	uid, err := parseUUID(arg.ID)
	if err != nil {
		return repository.Gate{}, err
	}

	setClauses := []string{}
	args := []any{uid}
	i := 2

	if arg.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", i))
		args = append(args, *arg.Name)
		i++
	}
	if arg.StatusTtlSeconds != nil {
		setClauses = append(setClauses, fmt.Sprintf("status_ttl_seconds = $%d", i))
		args = append(args, *arg.StatusTtlSeconds)
		i++
	}

	if len(setClauses) == 0 {
		return r.GetGateByID(ctx, arg.ID)
	}

	query := fmt.Sprintf(
		`UPDATE gates SET %s, updated_at = now() WHERE id = $1
		RETURNING id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at`,
		strings.Join(setClauses, ", "),
	)

	row := r.db.QueryRow(ctx, query, args...)
	var g gateRow
	if err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return repository.Gate{}, mapError(err)
	}
	return toRepoGate(g), nil
}

func (r *GateRepository) DeleteGate(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	row := r.db.QueryRow(ctx, `DELETE FROM gates WHERE id = $1 RETURNING id`, uid)
	var deleted pgtype.UUID
	return mapError(row.Scan(&deleted))
}

func (r *GateRepository) UpdateGateToken(ctx context.Context, arg repository.UpdateGateTokenParams) (repository.Gate, error) {
	uid, err := parseUUID(arg.ID)
	if err != nil {
		return repository.Gate{}, err
	}
	row := r.db.QueryRow(ctx,
		`UPDATE gates SET gate_token_hash = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at`,
		uid, arg.GateTokenHash,
	)
	var g gateRow
	if err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return repository.Gate{}, mapError(err)
	}
	return toRepoGate(g), nil
}
