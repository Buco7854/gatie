package repository

import "time"

type ListParams struct {
	Limit  int32
	Offset int32
}

type Member struct {
	ID           string
	Username     string
	DisplayName  *string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateMemberParams struct {
	Username     string
	DisplayName  *string
	PasswordHash string
	Role         string
}

type PatchMemberParams struct {
	ID                 string
	Username           *string
	DisplayName        *string
	SetDisplayNameNull bool
	Role               *string
}

type Gate struct {
	ID               string
	Name             string
	GateTokenHash    string
	StatusTtlSeconds int32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreateGateParams struct {
	Name             string
	GateTokenHash    string
	StatusTtlSeconds int32
}

type PatchGateParams struct {
	ID               string
	Name             *string
	StatusTtlSeconds *int32
}

type UpdateGateTokenParams struct {
	ID            string
	GateTokenHash string
}

type RefreshToken struct {
	ID        string
	MemberID  string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type CreateRefreshTokenParams struct {
	MemberID  string
	TokenHash string
	ExpiresAt time.Time
}
