package service

import (
	"errors"
	"time"
)

var (
	ErrInvalidID      = errors.New("invalid id format")
	ErrNothingToUpdate = errors.New("no fields to update")
)

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
