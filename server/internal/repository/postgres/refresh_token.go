package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/repository"
)

func (r *Repository) CreateRefreshToken(ctx context.Context, arg repository.CreateRefreshTokenParams) (repository.RefreshToken, error) {
	memberUID, err := parseUUID(arg.MemberID)
	if err != nil {
		return repository.RefreshToken{}, err
	}
	row := r.db.QueryRow(ctx,
		`INSERT INTO refresh_tokens (member_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, member_id, token_hash, expires_at, created_at`,
		memberUID, arg.TokenHash, pgtype.Timestamptz{Time: arg.ExpiresAt, Valid: true},
	)
	var rt refreshTokenRow
	if err := row.Scan(&rt.ID, &rt.MemberID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt); err != nil {
		return repository.RefreshToken{}, mapError(err)
	}
	return toRepoRefreshToken(rt), nil
}

func (r *Repository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (repository.RefreshToken, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, member_id, token_hash, expires_at, created_at
		FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	)
	var rt refreshTokenRow
	if err := row.Scan(&rt.ID, &rt.MemberID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt); err != nil {
		return repository.RefreshToken{}, mapError(err)
	}
	return toRepoRefreshToken(rt), nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, uid)
	return mapError(err)
}

func (r *Repository) DeleteRefreshTokensByMember(ctx context.Context, memberID string) error {
	uid, err := parseUUID(memberID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE member_id = $1`, uid)
	return mapError(err)
}

func (r *Repository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < now()`)
	return mapError(err)
}
