package postgres

import "github.com/jackc/pgx/v5/pgtype"

type memberRow struct {
	ID           pgtype.UUID
	Username     string
	DisplayName  pgtype.Text
	PasswordHash string
	Role         string
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

type gateActionRow struct {
	ID            pgtype.UUID
	GateID        pgtype.UUID
	ActionType    string
	TransportType string
	Config        []byte
}

type permissionRow struct {
	ID            pgtype.UUID
	MemberID      pgtype.UUID
	GateID        pgtype.UUID
	CanOpen       bool
	CanClose      bool
	CanViewStatus bool
	CanManage     bool
}

type scheduleRow struct {
	ID         pgtype.UUID
	Name       string
	Scope      string
	OwnerID    pgtype.UUID
	Expression []byte
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
}
