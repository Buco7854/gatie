package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
)

var ErrGateNotFound = errors.New("gate not found")

type GateService struct {
	queries *repository.Queries
}

func NewGateService(queries *repository.Queries) *GateService {
	return &GateService{queries: queries}
}

type GatePage struct {
	Gates []repository.Gate
	Total int64
}

type CreateGateInput struct {
	Name             string
	StatusTTLSeconds int32
}

type UpdateGateInput struct {
	Name             string
	StatusTTLSeconds int32
}

type GateWithToken struct {
	Gate  repository.Gate
	Token string // plain token, shown once
}

func (s *GateService) ListGates(ctx context.Context, page, perPage int) (*GatePage, error) {
	total, err := s.queries.CountGates(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting gates: %w", err)
	}

	gates, err := s.queries.ListGates(ctx, repository.ListGatesParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("listing gates: %w", err)
	}

	return &GatePage{Gates: gates, Total: total}, nil
}

func (s *GateService) GetGate(ctx context.Context, id pgtype.UUID) (*repository.Gate, error) {
	gate, err := s.queries.GetGateByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGateNotFound
		}
		return nil, fmt.Errorf("getting gate: %w", err)
	}
	return &gate, nil
}

func (s *GateService) CreateGate(ctx context.Context, input CreateGateInput) (*GateWithToken, error) {
	plainToken, hash, err := generateToken()
	if err != nil {
		return nil, err
	}

	ttl := input.StatusTTLSeconds
	if ttl <= 0 {
		ttl = 60
	}

	gate, err := s.queries.CreateGate(ctx, repository.CreateGateParams{
		Name:             input.Name,
		GateTokenHash:    hash,
		StatusTtlSeconds: ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("creating gate: %w", err)
	}

	return &GateWithToken{Gate: gate, Token: plainToken}, nil
}

func (s *GateService) UpdateGate(ctx context.Context, id pgtype.UUID, input UpdateGateInput) (*repository.Gate, error) {
	ttl := input.StatusTTLSeconds
	if ttl <= 0 {
		ttl = 60
	}

	gate, err := s.queries.UpdateGate(ctx, repository.UpdateGateParams{
		ID:               id,
		Name:             input.Name,
		StatusTtlSeconds: ttl,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGateNotFound
		}
		return nil, fmt.Errorf("updating gate: %w", err)
	}

	return &gate, nil
}

func (s *GateService) DeleteGate(ctx context.Context, id pgtype.UUID) error {
	_, err := s.queries.GetGateByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGateNotFound
		}
		return fmt.Errorf("getting gate: %w", err)
	}

	return s.queries.DeleteGate(ctx, id)
}

func (s *GateService) RegenerateToken(ctx context.Context, id pgtype.UUID) (*GateWithToken, error) {
	_, err := s.queries.GetGateByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGateNotFound
		}
		return nil, fmt.Errorf("getting gate: %w", err)
	}

	plainToken, hash, err := generateToken()
	if err != nil {
		return nil, err
	}

	gate, err := s.queries.UpdateGateToken(ctx, repository.UpdateGateTokenParams{
		ID:            id,
		GateTokenHash: hash,
	})
	if err != nil {
		return nil, fmt.Errorf("updating gate token: %w", err)
	}

	return &GateWithToken{Gate: gate, Token: plainToken}, nil
}

func generateToken() (plainToken, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating token: %w", err)
	}
	plainToken = hex.EncodeToString(b)
	hash = auth.HashToken(plainToken)
	return plainToken, hash, nil
}
