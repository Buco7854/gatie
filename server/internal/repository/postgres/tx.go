package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/service"
)

var (
	_ service.AuthRepository   = (*Repository)(nil)
	_ service.MemberRepository = (*Repository)(nil)
	_ service.GateRepository   = (*Repository)(nil)
)

func NewTxFactory(pool *pgxpool.Pool) func(ctx context.Context) (*Repository, service.Tx, error) {
	return func(ctx context.Context) (*Repository, service.Tx, error) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return nil, nil, err
		}
		return NewRepository(tx), tx, nil
	}
}
