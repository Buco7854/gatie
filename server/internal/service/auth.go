package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
)

type AuthService struct {
	queries *repository.Queries
	jwt     *auth.JWTManager
}

func NewAuthService(queries *repository.Queries, jwt *auth.JWTManager) *AuthService {
	return &AuthService{queries: queries, jwt: jwt}
}

type SetupInput struct {
	Username string
	Password string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	Member       repository.Member
}

func (s *AuthService) Setup(ctx context.Context, input SetupInput) (*AuthResult, error) {
	count, err := s.queries.CountMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting members: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("setup already completed")
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	member, err := s.queries.CreateMember(ctx, repository.CreateMemberParams{
		Username:     input.Username,
		PasswordHash: hash,
		Role:         "ADMIN",
	})
	if err != nil {
		return nil, fmt.Errorf("creating admin: %w", err)
	}

	return s.generateTokens(ctx, member)
}

type LoginInput struct {
	Username string
	Password string
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	member, err := s.queries.GetMemberByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !auth.CheckPassword(input.Password, member.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return s.generateTokens(ctx, member)
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*AuthResult, error) {
	tokenHash := auth.HashToken(rawToken)

	rt, err := s.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if rt.ExpiresAt.Time.Before(time.Now()) {
		s.queries.DeleteRefreshToken(ctx, rt.ID)
		return nil, fmt.Errorf("refresh token expired")
	}

	// Rotation : supprimer l'ancien, en créer un nouveau
	if err := s.queries.DeleteRefreshToken(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("revoking old token: %w", err)
	}

	member, err := s.queries.GetMemberByID(ctx, rt.MemberID)
	if err != nil {
		return nil, fmt.Errorf("member not found")
	}

	return s.generateTokens(ctx, member)
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	tokenHash := auth.HashToken(rawToken)

	rt, err := s.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil
	}

	return s.queries.DeleteRefreshToken(ctx, rt.ID)
}

func (s *AuthService) generateTokens(ctx context.Context, member repository.Member) (*AuthResult, error) {
	memberID := uuidToString(member.ID)

	accessToken, err := s.jwt.GenerateAccessToken(memberID, member.Role, member.Username)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshHash := auth.HashToken(rawRefresh)
	expiresAt := time.Now().Add(s.jwt.RefreshDuration())

	_, err = s.queries.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		MemberID:  member.ID,
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

func uuidToString(u pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
