package service

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidID = errors.New("invalid id format")

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
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
