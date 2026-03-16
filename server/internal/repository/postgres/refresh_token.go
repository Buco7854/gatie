package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateRefreshTokenParams struct {
	MemberID  pgtype.UUID
	TokenHash string
	ExpiresAt pgtype.Timestamptz
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRow(ctx,
		`INSERT INTO refresh_tokens (member_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, member_id, token_hash, expires_at, created_at`,
		arg.MemberID, arg.TokenHash, arg.ExpiresAt,
	)
	var rt RefreshToken
	err := row.Scan(&rt.ID, &rt.MemberID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt)
	return rt, MapError(err)
}

func (q *Queries) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	row := q.db.QueryRow(ctx,
		`SELECT id, member_id, token_hash, expires_at, created_at
		FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	)
	var rt RefreshToken
	err := row.Scan(&rt.ID, &rt.MemberID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt)
	return rt, MapError(err)
}

func (q *Queries) DeleteRefreshToken(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, id)
	return MapError(err)
}

func (q *Queries) DeleteRefreshTokensByMember(ctx context.Context, memberID pgtype.UUID) error {
	_, err := q.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE member_id = $1`, memberID)
	return MapError(err)
}

func (q *Queries) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := q.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < now()`)
	return MapError(err)
}
