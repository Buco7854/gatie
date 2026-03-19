package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gatie-io/gatie-server/internal/repository"
)

type GateMembershipRepository interface {
	ListGateMemberships(ctx context.Context, gateID string) ([]repository.GateMembershipWithMember, error)
	GetGateMembership(ctx context.Context, gateID, memberID string) (repository.GateMembership, error)
	CreateGateMembership(ctx context.Context, arg repository.CreateGateMembershipParams) (repository.GateMembership, error)
	UpdateGateMembership(ctx context.Context, arg repository.UpdateGateMembershipParams) (repository.GateMembership, error)
	DeleteGateMembership(ctx context.Context, gateID, memberID string) error
	GetGateByID(ctx context.Context, id string) (repository.Gate, error)
	GetMemberByID(ctx context.Context, id string) (repository.Member, error)
}

var (
	ErrGateMembershipNotFound = errors.New("gate membership not found")
	ErrGateMembershipExists   = errors.New("member already has access to this gate")
)

type GateMembershipService struct {
	repo GateMembershipRepository
}

func NewGateMembershipService(repo GateMembershipRepository) *GateMembershipService {
	return &GateMembershipService{repo: repo}
}

type GateMember struct {
	GateID      string
	MemberID    string
	Username    string
	DisplayName string
	RoleID      string
	CreatedAt   time.Time
}

type CreateGateMembershipInput struct {
	GateID   string
	MemberID string
	RoleID   string
}

type UpdateGateMembershipInput struct {
	RoleID string
}

func (s *GateMembershipService) ListGateMembers(ctx context.Context, gateID string) ([]GateMember, error) {
	if _, err := s.repo.GetGateByID(ctx, gateID); err != nil {
		return nil, mapGateError(err, "getting gate")
	}

	rows, err := s.repo.ListGateMemberships(ctx, gateID)
	if err != nil {
		return nil, fmt.Errorf("listing gate memberships: %w", err)
	}

	members := make([]GateMember, len(rows))
	for i, r := range rows {
		displayName := ""
		if r.DisplayName != nil {
			displayName = *r.DisplayName
		}
		members[i] = GateMember{
			GateID:      r.GateID,
			MemberID:    r.MemberID,
			Username:    r.Username,
			DisplayName: displayName,
			RoleID:      r.RoleID,
			CreatedAt:   r.CreatedAt,
		}
	}
	return members, nil
}

func (s *GateMembershipService) AddGateMember(ctx context.Context, input CreateGateMembershipInput) (*GateMember, error) {
	if _, err := s.repo.GetGateByID(ctx, input.GateID); err != nil {
		return nil, mapGateError(err, "getting gate")
	}

	member, err := s.repo.GetMemberByID(ctx, input.MemberID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("getting member: %w", err)
	}

	gm, err := s.repo.CreateGateMembership(ctx, repository.CreateGateMembershipParams{
		GateID:   input.GateID,
		MemberID: input.MemberID,
		RoleID:   input.RoleID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrGateMembershipExists
		}
		return nil, fmt.Errorf("creating gate membership: %w", err)
	}

	displayName := ""
	if member.DisplayName != nil {
		displayName = *member.DisplayName
	}

	return &GateMember{
		GateID:      gm.GateID,
		MemberID:    gm.MemberID,
		Username:    member.Username,
		DisplayName: displayName,
		RoleID:      gm.RoleID,
		CreatedAt:   gm.CreatedAt,
	}, nil
}

func (s *GateMembershipService) UpdateGateMember(ctx context.Context, gateID, memberID string, input UpdateGateMembershipInput) (*GateMember, error) {
	gm, err := s.repo.UpdateGateMembership(ctx, repository.UpdateGateMembershipParams{
		GateID:   gateID,
		MemberID: memberID,
		RoleID:   input.RoleID,
	})
	if err != nil {
		return nil, mapGateMembershipError(err, "updating gate membership")
	}

	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("getting member: %w", err)
	}

	displayName := ""
	if member.DisplayName != nil {
		displayName = *member.DisplayName
	}

	return &GateMember{
		GateID:      gm.GateID,
		MemberID:    gm.MemberID,
		Username:    member.Username,
		DisplayName: displayName,
		RoleID:      gm.RoleID,
		CreatedAt:   gm.CreatedAt,
	}, nil
}

func (s *GateMembershipService) RemoveGateMember(ctx context.Context, gateID, memberID string) error {
	if err := s.repo.DeleteGateMembership(ctx, gateID, memberID); err != nil {
		return mapGateMembershipError(err, "deleting gate membership")
	}
	return nil
}

func mapGateMembershipError(err error, fallback string) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return ErrInvalidID
	case errors.Is(err, repository.ErrNotFound):
		return ErrGateMembershipNotFound
	default:
		return fmt.Errorf("%s: %w", fallback, err)
	}
}

func mapGateError(err error, fallback string) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return ErrInvalidID
	case errors.Is(err, repository.ErrNotFound):
		return ErrGateNotFound
	default:
		return fmt.Errorf("%s: %w", fallback, err)
	}
}
