package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gatie-io/gatie-server/internal/repository"
)

type AuthorizationService struct {
	repo repository.AuthorizationRepository
}

func NewAuthorizationService(repo repository.AuthorizationRepository) *AuthorizationService {
	return &AuthorizationService{repo: repo}
}

func (s *AuthorizationService) Can(ctx context.Context, memberRoleID, memberID, gateID, permission string) (bool, error) {
	globalPerms, err := s.repo.GetRolePermissions(ctx, memberRoleID)
	if err != nil {
		return false, fmt.Errorf("getting global role permissions: %w", err)
	}

	if slices.Contains(globalPerms, PermAll) || slices.Contains(globalPerms, permission) {
		return true, nil
	}

	membership, err := s.repo.GetGateMembership(ctx, gateID, memberID)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("getting gate membership: %w", err)
	}

	localPerms, err := s.repo.GetRolePermissions(ctx, membership.RoleID)
	if err != nil {
		return false, fmt.Errorf("getting local role permissions: %w", err)
	}

	return slices.Contains(localPerms, PermAll) || slices.Contains(localPerms, permission), nil
}

func isNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}
