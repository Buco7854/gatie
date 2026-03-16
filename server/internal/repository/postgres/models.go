package postgres

import "github.com/jackc/pgx/v5/pgtype"

type Member struct {
	ID           pgtype.UUID        `json:"id"`
	Username     string             `json:"username"`
	DisplayName  pgtype.Text        `json:"display_name"`
	PasswordHash string             `json:"password_hash"`
	Role         string             `json:"role"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}

type Gate struct {
	ID               pgtype.UUID        `json:"id"`
	Name             string             `json:"name"`
	GateTokenHash    string             `json:"gate_token_hash"`
	StatusTtlSeconds int32              `json:"status_ttl_seconds"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
}

type RefreshToken struct {
	ID        pgtype.UUID        `json:"id"`
	MemberID  pgtype.UUID        `json:"member_id"`
	TokenHash string             `json:"token_hash"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type GateAction struct {
	ID            pgtype.UUID `json:"id"`
	GateID        pgtype.UUID `json:"gate_id"`
	ActionType    string      `json:"action_type"`
	TransportType string      `json:"transport_type"`
	Config        []byte      `json:"config"`
}

type Permission struct {
	ID            pgtype.UUID `json:"id"`
	MemberID      pgtype.UUID `json:"member_id"`
	GateID        pgtype.UUID `json:"gate_id"`
	CanOpen       bool        `json:"can_open"`
	CanClose      bool        `json:"can_close"`
	CanViewStatus bool        `json:"can_view_status"`
	CanManage     bool        `json:"can_manage"`
}

type Schedule struct {
	ID         pgtype.UUID        `json:"id"`
	Name       string             `json:"name"`
	Scope      string             `json:"scope"`
	OwnerID    pgtype.UUID        `json:"owner_id"`
	Expression []byte             `json:"expression"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
}

type MemberGateSchedule struct {
	MemberID   pgtype.UUID `json:"member_id"`
	GateID     pgtype.UUID `json:"gate_id"`
	ScheduleID pgtype.UUID `json:"schedule_id"`
}
