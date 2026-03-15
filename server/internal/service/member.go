package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
)

var (
	ErrMemberNotFound  = errors.New("member not found")
	ErrSelfDelete      = errors.New("cannot delete your own account")
	ErrSelfRoleChange  = errors.New("cannot change your own role")
	ErrLastAdmin       = errors.New("cannot delete the last admin")
	ErrUsernameExists  = errors.New("username already taken")
)

type MemberService struct {
	queries *repository.Queries
}

func NewMemberService(queries *repository.Queries) *MemberService {
	return &MemberService{queries: queries}
}

type MemberPage struct {
	Members []repository.Member
	Total   int64
}

type CreateMemberInput struct {
	Username    string
	DisplayName string
	Password    string
	Role        string
}

type UpdateMemberInput struct {
	Username    string
	DisplayName string
	Role        string
	CallerID    string
}

func (s *MemberService) ListMembers(ctx context.Context, page, perPage int) (*MemberPage, error) {
	total, err := s.queries.CountMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting members: %w", err)
	}

	members, err := s.queries.ListMembers(ctx, repository.ListMembersParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}

	return &MemberPage{Members: members, Total: total}, nil
}

func (s *MemberService) GetMember(ctx context.Context, id pgtype.UUID) (*repository.Member, error) {
	member, err := s.queries.GetMemberByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("getting member: %w", err)
	}
	return &member, nil
}

func (s *MemberService) CreateMember(ctx context.Context, input CreateMemberInput) (*repository.Member, error) {
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	displayName := pgtype.Text{}
	if input.DisplayName != "" {
		displayName = pgtype.Text{String: input.DisplayName, Valid: true}
	}

	member, err := s.queries.CreateMember(ctx, repository.CreateMemberParams{
		Username:     input.Username,
		DisplayName:  displayName,
		PasswordHash: hash,
		Role:         input.Role,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("creating member: %w", err)
	}

	return &member, nil
}

func (s *MemberService) UpdateMember(ctx context.Context, id pgtype.UUID, input UpdateMemberInput) (*repository.Member, error) {
	if input.CallerID == uuidToString(id) {
		current, err := s.queries.GetMemberByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrMemberNotFound
			}
			return nil, fmt.Errorf("getting member: %w", err)
		}
		if input.Role != current.Role {
			return nil, ErrSelfRoleChange
		}
	}

	displayName := pgtype.Text{}
	if input.DisplayName != "" {
		displayName = pgtype.Text{String: input.DisplayName, Valid: true}
	}

	member, err := s.queries.UpdateMember(ctx, repository.UpdateMemberParams{
		ID:          id,
		Username:    input.Username,
		DisplayName: displayName,
		Role:        input.Role,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMemberNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("updating member: %w", err)
	}

	return &member, nil
}

func (s *MemberService) DeleteMember(ctx context.Context, id pgtype.UUID, callerID string) error {
	member, err := s.queries.GetMemberByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMemberNotFound
		}
		return fmt.Errorf("getting member: %w", err)
	}

	if uuidToString(member.ID) == callerID {
		return ErrSelfDelete
	}

	if member.Role == "ADMIN" {
		adminCount, err := s.queries.CountMembersByRole(ctx, "ADMIN")
		if err != nil {
			return fmt.Errorf("counting admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	return s.queries.DeleteMember(ctx, id)
}
