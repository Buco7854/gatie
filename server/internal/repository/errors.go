package repository

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("unique constraint violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrInvalidID           = errors.New("invalid id format")
)
