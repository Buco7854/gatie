package repository

import "context"

type AuthRepository interface {
	BeginTx(ctx context.Context) (AuthRepository, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CountMembers(ctx context.Context) (int64, error)
	CreateMember(ctx context.Context, arg CreateMemberParams) (Member, error)
	GetMemberByUsername(ctx context.Context, username string) (Member, error)
	GetMemberByID(ctx context.Context, id string) (Member, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id string) error
	CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error)
}

type MemberRepository interface {
	BeginTx(ctx context.Context) (MemberRepository, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CountMembers(ctx context.Context) (int64, error)
	ListMembers(ctx context.Context, arg ListParams) ([]Member, error)
	GetMemberByID(ctx context.Context, id string) (Member, error)
	CreateMember(ctx context.Context, arg CreateMemberParams) (Member, error)
	PatchMember(ctx context.Context, arg PatchMemberParams) (Member, error)
	DeleteMember(ctx context.Context, id string) error
	CountMembersByRoleForUpdate(ctx context.Context, role string) (int64, error)
}

type GateRepository interface {
	BeginTx(ctx context.Context) (GateRepository, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CountGates(ctx context.Context) (int64, error)
	ListGates(ctx context.Context, arg ListParams) ([]Gate, error)
	GetGateByID(ctx context.Context, id string) (Gate, error)
	CreateGate(ctx context.Context, arg CreateGateParams) (Gate, error)
	PatchGate(ctx context.Context, arg PatchGateParams) (Gate, error)
	DeleteGate(ctx context.Context, id string) error
	UpdateGateToken(ctx context.Context, arg UpdateGateTokenParams) (Gate, error)
}

type RoleRepository interface {
	BeginTx(ctx context.Context) (RoleRepository, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	ListRoles(ctx context.Context) ([]Role, error)
	CreateRole(ctx context.Context, id, description string) (Role, error)
	UpdateRole(ctx context.Context, id, description string) (Role, error)
	DeleteRole(ctx context.Context, id string) error
	RoleInUse(ctx context.Context, roleID string) (bool, error)
	GetRolePermissions(ctx context.Context, roleID string) ([]string, error)
	DeleteRolePermissions(ctx context.Context, roleID string) error
	AddRolePermission(ctx context.Context, roleID, permissionID string) error
	ListPermissions(ctx context.Context) ([]Permission, error)
	CreatePermission(ctx context.Context, id, description string) (Permission, error)
	UpdatePermission(ctx context.Context, id, description string) (Permission, error)
	DeletePermission(ctx context.Context, id string) error
	PermissionInUse(ctx context.Context, permissionID string) (bool, error)
}

type GateMembershipRepository interface {
	ListGateMemberships(ctx context.Context, gateID string) ([]GateMembershipWithMember, error)
	GetGateMembership(ctx context.Context, gateID, memberID string) (GateMembership, error)
	CreateGateMembership(ctx context.Context, arg CreateGateMembershipParams) (GateMembership, error)
	UpdateGateMembership(ctx context.Context, arg UpdateGateMembershipParams) (GateMembership, error)
	DeleteGateMembership(ctx context.Context, gateID, memberID string) error
	GetGateByID(ctx context.Context, id string) (Gate, error)
	GetMemberByID(ctx context.Context, id string) (Member, error)
}

type AuthorizationRepository interface {
	GetRolePermissions(ctx context.Context, roleID string) ([]string, error)
	GetGateMembership(ctx context.Context, gateID, memberID string) (GateMembership, error)
}
