package postgres

import "github.com/gatie-io/gatie-server/internal/repository"

var (
	_ repository.AuthRepository           = (*AuthRepository)(nil)
	_ repository.MemberRepository         = (*MemberRepository)(nil)
	_ repository.GateRepository           = (*GateRepository)(nil)
	_ repository.RoleRepository           = (*RoleRepository)(nil)
	_ repository.GateMembershipRepository = (*GateMembershipRepository)(nil)
	_ repository.AuthorizationRepository  = (*AuthorizationRepository)(nil)
)
