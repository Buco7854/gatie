package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
)

type MemberRepository interface {
	CountMembers(ctx context.Context) (int64, error)
	ListMembers(ctx context.Context, arg repository.ListParams) ([]repository.Member, error)
	GetMemberByID(ctx context.Context, id string) (repository.Member, error)
	CreateMember(ctx context.Context, arg repository.CreateMemberParams) (repository.Member, error)
	PatchMember(ctx context.Context, arg repository.PatchMemberParams) (repository.Member, error)
	DeleteMember(ctx context.Context, id string) error
	CountMembersByRoleForUpdate(ctx context.Context, role string) (int64, error)
}

var (
	ErrMemberNotFound = errors.New("member not found")
	ErrSelfDelete     = errors.New("cannot delete your own account")
	ErrSelfRoleChange = errors.New("cannot change your own role")
	ErrLastAdmin      = errors.New("cannot delete the last admin")
	ErrUsernameExists = errors.New("username already taken")
)

type MemberService struct {
	repo    MemberRepository
	beginTx func(ctx context.Context) (MemberRepository, Tx, error)
}

func NewMemberService(repo MemberRepository, beginTx func(ctx context.Context) (MemberRepository, Tx, error)) *MemberService {
	return &MemberService{repo: repo, beginTx: beginTx}
}

type MemberPage struct {
	Members []Member
	Total   int64
}

type CreateMemberInput struct {
	Username    string
	DisplayName string
	Password    string
	Role        string
}

type UpdateMemberInput struct {
	Username           *string
	DisplayName        *string
	SetDisplayNameNull bool
	Role               *string
	CallerID           string
}

func (s *MemberService) ListMembers(ctx context.Context, page, perPage int) (*MemberPage, error) {
	total, err := s.repo.CountMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting members: %w", err)
	}

	rows, err := s.repo.ListMembers(ctx, repository.ListParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}

	members := make([]Member, len(rows))
	for i, r := range rows {
		members[i] = toMember(r)
	}

	return &MemberPage{Members: members, Total: total}, nil
}

func (s *MemberService) GetMember(ctx context.Context, id string) (*Member, error) {
	row, err := s.repo.GetMemberByID(ctx, id)
	if err != nil {
		return nil, mapMemberError(err, "getting member")
	}

	m := toMember(row)
	return &m, nil
}

func (s *MemberService) CreateMember(ctx context.Context, input CreateMemberInput) (*Member, error) {
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	var displayName *string
	if input.DisplayName != "" {
		displayName = &input.DisplayName
	}

	row, err := s.repo.CreateMember(ctx, repository.CreateMemberParams{
		Username:     input.Username,
		DisplayName:  displayName,
		PasswordHash: hash,
		Role:         input.Role,
	})
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("creating member: %w", err)
	}

	m := toMember(row)
	return &m, nil
}

func (s *MemberService) UpdateMember(ctx context.Context, id string, input UpdateMemberInput) (*Member, error) {
	if input.Role != nil && input.CallerID == id {
		current, err := s.repo.GetMemberByID(ctx, id)
		if err != nil {
			return nil, mapMemberError(err, "getting member")
		}
		if *input.Role != current.Role {
			return nil, ErrSelfRoleChange
		}
	}

	params := repository.PatchMemberParams{ID: id}

	if input.Username != nil {
		params.Username = input.Username
	}
	if input.SetDisplayNameNull {
		params.SetDisplayNameNull = true
	} else if input.DisplayName != nil {
		params.DisplayName = input.DisplayName
	}
	if input.Role != nil {
		params.Role = input.Role
	}

	row, err := s.repo.PatchMember(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrUsernameExists
		}
		return nil, mapMemberError(err, "updating member")
	}

	m := toMember(row)
	return &m, nil
}

func (s *MemberService) DeleteMember(ctx context.Context, id string, callerID string) error {
	if id == callerID {
		return ErrSelfDelete
	}

	qtx, tx, err := s.beginTx(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := qtx.GetMemberByID(ctx, id)
	if err != nil {
		return mapMemberError(err, "getting member")
	}

	if row.Role == RoleAdmin {
		adminCount, err := qtx.CountMembersByRoleForUpdate(ctx, RoleAdmin)
		if err != nil {
			return fmt.Errorf("counting admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	if err := qtx.DeleteMember(ctx, id); err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}

	return tx.Commit(ctx)
}

func mapMemberError(err error, fallback string) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return ErrInvalidID
	case errors.Is(err, repository.ErrNotFound):
		return ErrMemberNotFound
	default:
		return fmt.Errorf("%s: %w", fallback, err)
	}
}

func toMember(r repository.Member) Member {
	displayName := ""
	if r.DisplayName != nil {
		displayName = *r.DisplayName
	}
	return Member{
		ID:          r.ID,
		Username:    r.Username,
		DisplayName: displayName,
		Role:        r.Role,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
