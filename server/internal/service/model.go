package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidID = errors.New("invalid id format")

// TxBeginner abstracts the ability to begin a database transaction.
// *pgxpool.Pool implements this interface.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Member struct {
	ID          string
	Username    string
	DisplayName string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Gate struct {
	ID               string
	Name             string
	StatusTTLSeconds int32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
