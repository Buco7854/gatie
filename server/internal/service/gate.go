package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/convert"
	"github.com/gatie-io/gatie-server/internal/repository"
	"github.com/gatie-io/gatie-server/internal/repository/postgres"
)

var ErrGateNotFound = errors.New("gate not found")

type GateService struct {
	queries *postgres.Queries
}

func NewGateService(queries *postgres.Queries) *GateService {
	return &GateService{queries: queries}
}

type GatePage struct {
	Gates []Gate
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
	Gate  Gate
	Token string
}

func (s *GateService) ListGates(ctx context.Context, page, perPage int) (*GatePage, error) {
	total, err := s.queries.CountGates(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting gates: %w", err)
	}

	rows, err := s.queries.ListGates(ctx, postgres.ListGatesParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("listing gates: %w", err)
	}

	gates := make([]Gate, len(rows))
	for i, r := range rows {
		gates[i] = toGate(r)
	}

	return &GatePage{Gates: gates, Total: total}, nil
}

func (s *GateService) GetGate(ctx context.Context, id string) (*Gate, error) {
	uid, err := convert.ParseUUID(id)
	if err != nil {
		return nil, ErrGateNotFound
	}

	row, err := s.queries.GetGateByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrGateNotFound
		}
		return nil, fmt.Errorf("getting gate: %w", err)
	}

	g := toGate(row)
	return &g, nil
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

	row, err := s.queries.CreateGate(ctx, postgres.CreateGateParams{
		Name:             input.Name,
		GateTokenHash:    hash,
		StatusTtlSeconds: ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("creating gate: %w", err)
	}

	return &GateWithToken{Gate: toGate(row), Token: plainToken}, nil
}

func (s *GateService) UpdateGate(ctx context.Context, id string, input UpdateGateInput) (*Gate, error) {
	uid, err := convert.ParseUUID(id)
	if err != nil {
		return nil, ErrGateNotFound
	}

	ttl := input.StatusTTLSeconds
	if ttl <= 0 {
		ttl = 60
	}

	row, err := s.queries.UpdateGate(ctx, postgres.UpdateGateParams{
		ID:               uid,
		Name:             input.Name,
		StatusTtlSeconds: ttl,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrGateNotFound
		}
		return nil, fmt.Errorf("updating gate: %w", err)
	}

	g := toGate(row)
	return &g, nil
}

func (s *GateService) DeleteGate(ctx context.Context, id string) error {
	uid, err := convert.ParseUUID(id)
	if err != nil {
		return ErrGateNotFound
	}

	_, err = s.queries.GetGateByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrGateNotFound
		}
		return fmt.Errorf("getting gate: %w", err)
	}

	return s.queries.DeleteGate(ctx, uid)
}

func (s *GateService) RegenerateToken(ctx context.Context, id string) (*GateWithToken, error) {
	uid, err := convert.ParseUUID(id)
	if err != nil {
		return nil, ErrGateNotFound
	}

	_, err = s.queries.GetGateByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrGateNotFound
		}
		return nil, fmt.Errorf("getting gate: %w", err)
	}

	plainToken, hash, err := generateToken()
	if err != nil {
		return nil, err
	}

	row, err := s.queries.UpdateGateToken(ctx, postgres.UpdateGateTokenParams{
		ID:            uid,
		GateTokenHash: hash,
	})
	if err != nil {
		return nil, fmt.Errorf("updating gate token: %w", err)
	}

	return &GateWithToken{Gate: toGate(row), Token: plainToken}, nil
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

func toGate(r postgres.Gate) Gate {
	return Gate{
		ID:               convert.UUIDToString(r.ID),
		Name:             r.Name,
		StatusTTLSeconds: r.StatusTtlSeconds,
		CreatedAt:        r.CreatedAt.Time,
		UpdatedAt:        r.UpdatedAt.Time,
	}
}
