package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
)

type AuthRepository interface {
	CountMembers(ctx context.Context) (int64, error)
	CreateMember(ctx context.Context, arg repository.CreateMemberParams) (repository.Member, error)
	GetMemberByUsername(ctx context.Context, username string) (repository.Member, error)
	GetMemberByID(ctx context.Context, id string) (repository.Member, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (repository.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id string) error
	CreateRefreshToken(ctx context.Context, arg repository.CreateRefreshTokenParams) (repository.RefreshToken, error)
}

type AuthService struct {
	repo AuthRepository
	tx   repository.Transactor
	jwt  *auth.JWTManager
}

var (
	ErrSetupAlreadyCompleted = errors.New("setup already completed")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrRefreshTokenExpired   = errors.New("refresh token expired")
)

func NewAuthService(repo AuthRepository, tx repository.Transactor, jwt *auth.JWTManager) *AuthService {
	return &AuthService{repo: repo, tx: tx, jwt: jwt}
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

	var result *AuthResult
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		rt, err := s.repo.GetRefreshTokenByHash(txCtx, tokenHash)
		if err != nil {
			return ErrInvalidRefreshToken
		}

		if err := s.repo.DeleteRefreshToken(txCtx, rt.ID); err != nil {
			return fmt.Errorf("revoking old token: %w", err)
		}

		if rt.ExpiresAt.Before(time.Now()) {
			return ErrRefreshTokenExpired
		}

		row, err := s.repo.GetMemberByID(txCtx, rt.MemberID)
		if err != nil {
			return fmt.Errorf("getting member for refresh: %w", err)
		}

		member := toMember(row)
		accessToken, err := s.jwt.GenerateAccessToken(member.ID, member.Role, member.Username)
		if err != nil {
			return fmt.Errorf("generating access token: %w", err)
		}

		rawRefresh, err := auth.GenerateRefreshToken()
		if err != nil {
			return err
		}

		refreshHash := auth.HashToken(rawRefresh)
		expiresAt := time.Now().Add(s.jwt.RefreshDuration())

		_, err = s.repo.CreateRefreshToken(txCtx, repository.CreateRefreshTokenParams{
			MemberID:  rt.MemberID,
			TokenHash: refreshHash,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return fmt.Errorf("storing refresh token: %w", err)
		}

		result = &AuthResult{
			AccessToken:  accessToken,
			RefreshToken: rawRefresh,
			Member:       member,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
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
