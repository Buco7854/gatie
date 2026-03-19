package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
)

type AuthService struct {
	repo repository.AuthRepository
	jwt  *auth.JWTManager
}

var (
	ErrSetupAlreadyCompleted = errors.New("setup already completed")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrRefreshTokenExpired   = errors.New("refresh token expired")
)

func NewAuthService(repo repository.AuthRepository, jwt *auth.JWTManager) *AuthService {
	return &AuthService{repo: repo, jwt: jwt}
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
	count, err := s.repo.CountMembers(ctx)
	if err != nil {
		return false, fmt.Errorf("counting members: %w", err)
	}
	return count == 0, nil
}

func (s *AuthService) Setup(ctx context.Context, input SetupInput) (*AuthResult, error) {
	count, err := s.repo.CountMembers(ctx)
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

	row, err := s.repo.CreateMember(ctx, repository.CreateMemberParams{
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
	row, err := s.repo.GetMemberByUsername(ctx, input.Username)
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

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rt, err := tx.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if err := tx.DeleteRefreshToken(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("revoking old token: %w", err)
	}

	if rt.ExpiresAt.Before(time.Now()) {
		tx.Commit(ctx)
		return nil, ErrRefreshTokenExpired
	}

	row, err := tx.GetMemberByID(ctx, rt.MemberID)
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

	_, err = tx.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		MemberID:  rt.MemberID,
		TokenHash: refreshHash,
		ExpiresAt: expiresAt,
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

	rt, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("looking up refresh token: %w", err)
	}

	if err := s.repo.DeleteRefreshToken(ctx, rt.ID); err != nil {
		return fmt.Errorf("deleting refresh token: %w", err)
	}
	return nil
}

func (s *AuthService) generateTokens(ctx context.Context, row repository.Member) (*AuthResult, error) {
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

	_, err = s.repo.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		MemberID:  row.ID,
		TokenHash: refreshHash,
		ExpiresAt: expiresAt,
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
