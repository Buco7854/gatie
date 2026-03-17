package repository

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("unique constraint violation")
	ErrInvalidID = errors.New("invalid id format")
)
