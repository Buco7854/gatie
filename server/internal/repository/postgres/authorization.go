package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/repository"
)

type AuthorizationRepository struct{ base }

func NewAuthorizationRepository(pool *pgxpool.Pool) *AuthorizationRepository {
	return &AuthorizationRepository{base{db: pool, pool: pool}}
}

func (r *AuthorizationRepository) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	return queryRolePermissions(ctx, r.db, roleID)
}

func (r *AuthorizationRepository) GetGateMembership(ctx context.Context, gateID, memberID string) (repository.GateMembership, error) {
	return queryGateMembership(ctx, r.db, gateID, memberID)
}
