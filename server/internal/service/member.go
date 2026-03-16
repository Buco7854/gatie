package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/convert"
	"github.com/gatie-io/gatie-server/internal/repository"
	"github.com/gatie-io/gatie-server/internal/repository/postgres"
)

var (
	ErrMemberNotFound = errors.New("member not found")
	ErrSelfDelete     = errors.New("cannot delete your own account")
	ErrSelfRoleChange = errors.New("cannot change your own role")
	ErrLastAdmin      = errors.New("cannot delete the last admin")
	ErrUsernameExists = errors.New("username already taken")
)

type MemberService struct {
	queries *postgres.Queries
	pool    *pgxpool.Pool
}

func NewMemberService(queries *postgres.Queries, pool *pgxpool.Pool) *MemberService {
	return &MemberService{queries: queries, pool: pool}
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

	rows, err := s.queries.ListMembers(ctx, postgres.ListMembersParams{
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
	uid, err := convert.ParseUUID(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	row, err := s.queries.GetMemberByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("getting member: %w", err)
	}

	m := toMember(row)
	return &m, nil
}

func (s *MemberService) CreateMember(ctx context.Context, input CreateMemberInput) (*Member, error) {
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	displayName := pgtype.Text{}
	if input.DisplayName != "" {
		displayName = pgtype.Text{String: input.DisplayName, Valid: true}
	}

	row, err := s.queries.CreateMember(ctx, postgres.CreateMemberParams{
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
	uid, err := convert.ParseUUID(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	if input.CallerID == id {
		current, err := s.queries.GetMemberByID(ctx, uid)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
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

	row, err := s.queries.UpdateMember(ctx, postgres.UpdateMemberParams{
		ID:          uid,
		Username:    input.Username,
		DisplayName: displayName,
		Role:        input.Role,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrMemberNotFound
		}
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("updating member: %w", err)
	}

	m := toMember(row)
	return &m, nil
}

func (s *MemberService) DeleteMember(ctx context.Context, id string, callerID string) error {
	uid, err := convert.ParseUUID(id)
	if err != nil {
		return ErrInvalidID
	}

	if id == callerID {
		return ErrSelfDelete
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	row, err := qtx.GetMemberByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrMemberNotFound
		}
		return fmt.Errorf("getting member: %w", err)
	}

	if row.Role == "ADMIN" {
		adminCount, err := qtx.CountMembersByRoleForUpdate(ctx, "ADMIN")
		if err != nil {
			return fmt.Errorf("counting admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	if err := qtx.DeleteMember(ctx, uid); err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}

	return tx.Commit(ctx)
}

func toMember(r postgres.Member) Member {
	displayName := ""
	if r.DisplayName.Valid {
		displayName = r.DisplayName.String
	}
	return Member{
		ID:          convert.UUIDToString(r.ID),
		Username:    r.Username,
		DisplayName: displayName,
		Role:        r.Role,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}
}
