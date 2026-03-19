package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/repository"
)

type AuthRepository struct{ base }

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{base{pool: pool}}
}

// --- Members (subset needed by auth) ---

func (r *AuthRepository) CountMembers(ctx context.Context) (int64, error) {
	row := r.conn(ctx).QueryRow(ctx, `SELECT count(*) FROM members`)
	var count int64
	err := row.Scan(&count)
	return count, mapError(err)
}

func (r *AuthRepository) CreateMember(ctx context.Context, arg repository.CreateMemberParams) (repository.Member, error) {
	displayName := pgtype.Text{}
	if arg.DisplayName != nil {
		displayName = pgtype.Text{String: *arg.DisplayName, Valid: true}
	}
	row := r.conn(ctx).QueryRow(ctx,
		`INSERT INTO members (username, display_name, password_hash, role_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, display_name, password_hash, role_id, created_at, updated_at`,
		arg.Username, displayName, arg.PasswordHash, arg.Role,
	)
	var m memberRow
	if err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return repository.Member{}, mapError(err)
	}
	return toRepoMember(m), nil
}

func (r *AuthRepository) GetMemberByUsername(ctx context.Context, username string) (repository.Member, error) {
	return queryMemberByUsername(ctx, r.conn(ctx), username)
}

func (r *AuthRepository) GetMemberByID(ctx context.Context, id string) (repository.Member, error) {
	return queryMemberByID(ctx, r.conn(ctx), id)
}

// --- Refresh tokens ---

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, arg repository.CreateRefreshTokenParams) (repository.RefreshToken, error) {
	memberUID, err := parseUUID(arg.MemberID)
	if err != nil {
		return repository.RefreshToken{}, err
	}
	row := r.conn(ctx).QueryRow(ctx,
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

func (r *AuthRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (repository.RefreshToken, error) {
	row := r.conn(ctx).QueryRow(ctx,
		`SELECT id, member_id, token_hash, expires_at, created_at
		FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	)
	var rt refreshTokenRow
	if err := row.Scan(&rt.ID, &rt.MemberID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt); err != nil {
		return repository.RefreshToken{}, mapError(err)
	}
	return toRepoRefreshToken(rt), nil
}

func (r *AuthRepository) DeleteRefreshToken(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	_, err = r.conn(ctx).Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, uid)
	return mapError(err)
}

func (r *AuthRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < now()`)
	return mapError(err)
}
