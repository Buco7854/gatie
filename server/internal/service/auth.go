package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
	"github.com/gatie-io/gatie-server/internal/repository/postgres"
)

type AuthService struct {
	queries postgres.Querier
	pool    TxBeginner
	jwt     *auth.JWTManager
}

var (
	ErrSetupAlreadyCompleted = errors.New("setup already completed")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrRefreshTokenExpired   = errors.New("refresh token expired")
)

func NewAuthService(queries postgres.Querier, pool TxBeginner, jwt *auth.JWTManager) *AuthService {
	return &AuthService{queries: queries, pool: pool, jwt: jwt}
}

type SetupInput struct {
	Username string
	Password string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	Member       Member
}

func (s *AuthService) NeedsSetup(ctx context.Context) (bool, error) {
	count, err := s.queries.CountMembers(ctx)
	if err != nil {
		return false, fmt.Errorf("counting members: %w", err)
	}
	return count == 0, nil
}

func (s *AuthService) Setup(ctx context.Context, input SetupInput) (*AuthResult, error) {
	count, err := s.queries.CountMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting members: %w", err)
	}
	if count > 0 {
		return nil, ErrSetupAlreadyCompleted
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	row, err := s.queries.CreateMember(ctx, postgres.CreateMemberParams{
		Username:     input.Username,
		PasswordHash: hash,
		Role:         RoleAdmin,
	})
	if err != nil {
		return nil, fmt.Errorf("creating admin: %w", err)
	}

	return s.generateTokens(ctx, row)
}

type LoginInput struct {
	Username string
	Password string
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	row, err := s.queries.GetMemberByUsername(ctx, input.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPassword(input.Password, row.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, row)
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*AuthResult, error) {
	tokenHash := auth.HashToken(rawToken)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	rt, err := qtx.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if err := qtx.DeleteRefreshToken(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("revoking old token: %w", err)
	}

	if rt.ExpiresAt.Time.Before(time.Now()) {
		tx.Commit(ctx)
		return nil, ErrRefreshTokenExpired
	}

	row, err := qtx.GetMemberByID(ctx, rt.MemberID)
	if err != nil {
		return nil, fmt.Errorf("getting member for refresh: %w", err)
	}

	member := toMember(row)
	accessToken, err := s.jwt.GenerateAccessToken(member.ID, member.Role, member.Username)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshHash := auth.HashToken(rawRefresh)
	expiresAt := time.Now().Add(s.jwt.RefreshDuration())

	_, err = qtx.CreateRefreshToken(ctx, postgres.CreateRefreshTokenParams{
		MemberID:  rt.MemberID,
		TokenHash: refreshHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		Member:       member,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	tokenHash := auth.HashToken(rawToken)

	rt, err := s.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("looking up refresh token: %w", err)
	}

	return s.queries.DeleteRefreshToken(ctx, rt.ID)
}

func (s *AuthService) generateTokens(ctx context.Context, row postgres.Member) (*AuthResult, error) {
	member := toMember(row)

	accessToken, err := s.jwt.GenerateAccessToken(member.ID, member.Role, member.Username)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshHash := auth.HashToken(rawRefresh)
	expiresAt := time.Now().Add(s.jwt.RefreshDuration())

	_, err = s.queries.CreateRefreshToken(ctx, postgres.CreateRefreshTokenParams{
		MemberID:  row.ID,
		TokenHash: refreshHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		Member:       member,
	}, nil
}
