package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type base struct {
	db   DBTX
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func (b *base) Commit(ctx context.Context) error {
	if b.tx == nil {
		return nil
	}
	return b.tx.Commit(ctx)
}

func (b *base) Rollback(ctx context.Context) error {
	if b.tx == nil {
		return nil
	}
	return b.tx.Rollback(ctx)
}

func (b *base) beginTx(ctx context.Context) (base, error) {
	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return base{}, err
	}
	return base{db: tx, pool: b.pool, tx: tx}, nil
}
