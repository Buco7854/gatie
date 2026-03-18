package postgres

import "github.com/jackc/pgx/v5/pgtype"

type memberRow struct {
	ID           pgtype.UUID
	Username     string
	DisplayName  pgtype.Text
	PasswordHash string
	RoleID       string
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
}

type gateRow struct {
	ID               pgtype.UUID
	Name             string
	GateTokenHash    string
	StatusTtlSeconds int32
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
}

type refreshTokenRow struct {
	ID        pgtype.UUID
	MemberID  pgtype.UUID
	TokenHash string
	ExpiresAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
}

type gateMembershipRow struct {
	GateID    pgtype.UUID
	MemberID  pgtype.UUID
	RoleID    string
	CreatedAt pgtype.Timestamptz
}
