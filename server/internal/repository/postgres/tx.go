package postgres

import "github.com/gatie-io/gatie-server/internal/service"

var (
	_ service.AuthRepository           = (*AuthRepository)(nil)
	_ service.MemberRepository         = (*MemberRepository)(nil)
	_ service.GateRepository           = (*GateRepository)(nil)
	_ service.RoleRepository           = (*RoleRepository)(nil)
	_ service.GateMembershipRepository = (*GateMembershipRepository)(nil)
	_ service.AuthorizationRepository  = (*AuthorizationRepository)(nil)
)
