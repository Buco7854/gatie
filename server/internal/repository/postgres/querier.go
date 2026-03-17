package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Querier is the interface implemented by *Queries. Services should depend
// on this interface rather than the concrete type, enabling unit testing
// without a real database.
type Querier interface {
	WithTx(tx pgx.Tx) Querier

	// Members
	CountMembers(ctx context.Context) (int64, error)
	CountMembersByRole(ctx context.Context, role string) (int64, error)
	CountMembersByRoleForUpdate(ctx context.Context, role string) (int64, error)
	CreateMember(ctx context.Context, arg CreateMemberParams) (Member, error)
	DeleteMember(ctx context.Context, id pgtype.UUID) error
	GetMemberByID(ctx context.Context, id pgtype.UUID) (Member, error)
	GetMemberByUsername(ctx context.Context, username string) (Member, error)
	ListMembers(ctx context.Context, arg ListMembersParams) ([]Member, error)
	PatchMember(ctx context.Context, arg PatchMemberParams) (Member, error)

	// Gates
	CountGates(ctx context.Context) (int64, error)
	CreateGate(ctx context.Context, arg CreateGateParams) (Gate, error)
	DeleteGate(ctx context.Context, id pgtype.UUID) error
	GetGateByID(ctx context.Context, id pgtype.UUID) (Gate, error)
	ListGates(ctx context.Context, arg ListGatesParams) ([]Gate, error)
	PatchGate(ctx context.Context, arg PatchGateParams) (Gate, error)
	UpdateGateToken(ctx context.Context, arg UpdateGateTokenParams) (Gate, error)

	// Gate Actions
	CreateGateAction(ctx context.Context, arg CreateGateActionParams) (GateAction, error)
	DeleteGateAction(ctx context.Context, arg DeleteGateActionParams) error
	GetGateAction(ctx context.Context, arg GetGateActionParams) (GateAction, error)
	ListGateActionsByGate(ctx context.Context, gateID pgtype.UUID) ([]GateAction, error)
	UpdateGateAction(ctx context.Context, arg UpdateGateActionParams) (GateAction, error)

	// Permissions
	CreatePermission(ctx context.Context, arg CreatePermissionParams) (Permission, error)
	DeletePermission(ctx context.Context, arg DeletePermissionParams) error
	GetPermission(ctx context.Context, arg GetPermissionParams) (Permission, error)
	ListPermissionsByGate(ctx context.Context, gateID pgtype.UUID) ([]Permission, error)
	ListPermissionsByMember(ctx context.Context, memberID pgtype.UUID) ([]Permission, error)
	UpdatePermission(ctx context.Context, arg UpdatePermissionParams) (Permission, error)

	// Refresh Tokens
	CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id pgtype.UUID) error
	DeleteRefreshTokensByMember(ctx context.Context, memberID pgtype.UUID) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error)

	// Schedules
	AttachScheduleToMemberGate(ctx context.Context, arg AttachScheduleToMemberGateParams) error
	CreateSchedule(ctx context.Context, arg CreateScheduleParams) (Schedule, error)
	DeleteSchedule(ctx context.Context, id pgtype.UUID) error
	DetachScheduleFromMemberGate(ctx context.Context, arg DetachScheduleFromMemberGateParams) error
	GetScheduleByID(ctx context.Context, id pgtype.UUID) (Schedule, error)
	ListAdminSchedules(ctx context.Context, arg ListAdminSchedulesParams) ([]Schedule, error)
	ListPersonalSchedules(ctx context.Context, arg ListPersonalSchedulesParams) ([]Schedule, error)
	ListSchedulesForMemberGate(ctx context.Context, arg ListSchedulesForMemberGateParams) ([]Schedule, error)
	UpdateSchedule(ctx context.Context, arg UpdateScheduleParams) (Schedule, error)
}

// compile-time check
var _ Querier = (*Queries)(nil)
