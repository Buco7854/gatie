package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/repository"
)

type GateRepository interface {
	BeginTx(ctx context.Context) (GateRepository, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CountGates(ctx context.Context) (int64, error)
	ListGates(ctx context.Context, arg repository.ListParams) ([]repository.Gate, error)
	GetGateByID(ctx context.Context, id string) (repository.Gate, error)
	CreateGate(ctx context.Context, arg repository.CreateGateParams) (repository.Gate, error)
	PatchGate(ctx context.Context, arg repository.PatchGateParams) (repository.Gate, error)
	DeleteGate(ctx context.Context, id string) error
	UpdateGateToken(ctx context.Context, arg repository.UpdateGateTokenParams) (repository.Gate, error)
}

var (
	ErrGateNotFound = errors.New("gate not found")
	ErrInvalidTTL   = errors.New("status_ttl_seconds must be between 1 and 86400")
)

type GateService struct {
	repo GateRepository
}

func NewGateService(repo GateRepository) *GateService {
	return &GateService{repo: repo}
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
	Name             *string
	StatusTTLSeconds *int32
}

type GateWithToken struct {
	Gate  Gate
	Token string
}

func (s *GateService) ListGates(ctx context.Context, page, perPage int) (*GatePage, error) {
	total, err := s.repo.CountGates(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting gates: %w", err)
	}

	rows, err := s.repo.ListGates(ctx, repository.ListParams{
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
	row, err := s.repo.GetGateByID(ctx, id)
	if err != nil {
		return nil, mapGateServiceError(err, "getting gate")
	}

	g := toGate(row)
	return &g, nil
}

func (s *GateService) CreateGate(ctx context.Context, input CreateGateInput) (*GateWithToken, error) {
	plainToken, hash, err := generateToken()
	if err != nil {
		return nil, err
	}

	if input.StatusTTLSeconds < 1 || input.StatusTTLSeconds > 86400 {
		return nil, ErrInvalidTTL
	}

	row, err := s.repo.CreateGate(ctx, repository.CreateGateParams{
		Name:             input.Name,
		GateTokenHash:    hash,
		StatusTtlSeconds: input.StatusTTLSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("creating gate: %w", err)
	}

	return &GateWithToken{Gate: toGate(row), Token: plainToken}, nil
}

func (s *GateService) UpdateGate(ctx context.Context, id string, input UpdateGateInput) (*Gate, error) {
	if input.Name == nil && input.StatusTTLSeconds == nil {
		return nil, ErrNothingToUpdate
	}

	params := repository.PatchGateParams{ID: id}

	if input.Name != nil {
		params.Name = input.Name
	}
	if input.StatusTTLSeconds != nil {
		if *input.StatusTTLSeconds < 1 || *input.StatusTTLSeconds > 86400 {
			return nil, ErrInvalidTTL
		}
		params.StatusTtlSeconds = input.StatusTTLSeconds
	}

	row, err := s.repo.PatchGate(ctx, params)
	if err != nil {
		return nil, mapGateServiceError(err, "updating gate")
	}

	g := toGate(row)
	return &g, nil
}

func (s *GateService) DeleteGate(ctx context.Context, id string) error {
	if err := s.repo.DeleteGate(ctx, id); err != nil {
		return mapGateServiceError(err, "deleting gate")
	}
	return nil
}

func (s *GateService) RegenerateToken(ctx context.Context, id string) (*GateWithToken, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.GetGateByID(ctx, id)
	if err != nil {
		return nil, mapGateServiceError(err, "getting gate")
	}

	plainToken, hash, err := generateToken()
	if err != nil {
		return nil, err
	}

	row, err := tx.UpdateGateToken(ctx, repository.UpdateGateTokenParams{
		ID:            id,
		GateTokenHash: hash,
	})
	if err != nil {
		return nil, fmt.Errorf("updating gate token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
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

func mapGateServiceError(err error, fallback string) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return ErrInvalidID
	case errors.Is(err, repository.ErrNotFound):
		return ErrGateNotFound
	default:
		return fmt.Errorf("%s: %w", fallback, err)
	}
}

func toGate(r repository.Gate) Gate {
	return Gate{
		ID:               r.ID,
		Name:             r.Name,
		StatusTTLSeconds: r.StatusTtlSeconds,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}
