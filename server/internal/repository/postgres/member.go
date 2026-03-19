package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/repository"
	"github.com/gatie-io/gatie-server/internal/service"
)

type MemberRepository struct{ base }

func NewMemberRepository(pool *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{base{db: pool, pool: pool}}
}

func (r *MemberRepository) BeginTx(ctx context.Context) (service.MemberRepository, error) {
	b, err := r.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	return &MemberRepository{b}, nil
}

func (r *MemberRepository) CountMembers(ctx context.Context) (int64, error) {
	row := r.db.QueryRow(ctx, `SELECT count(*) FROM members`)
	var count int64
	err := row.Scan(&count)
	return count, mapError(err)
}

func (r *MemberRepository) CountMembersByRole(ctx context.Context, role string) (int64, error) {
	row := r.db.QueryRow(ctx, `SELECT count(*) FROM members WHERE role_id = $1`, role)
	var count int64
	err := row.Scan(&count)
	return count, mapError(err)
}

func (r *MemberRepository) CountMembersByRoleForUpdate(ctx context.Context, role string) (int64, error) {
	row := r.db.QueryRow(ctx, `SELECT count(*) FROM members WHERE role_id = $1 FOR UPDATE`, role)
	var count int64
	err := row.Scan(&count)
	return count, mapError(err)
}

func (r *MemberRepository) ListMembers(ctx context.Context, arg repository.ListParams) ([]repository.Member, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, username, display_name, password_hash, role_id, created_at, updated_at
		FROM members ORDER BY created_at ASC LIMIT $1 OFFSET $2`,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var out []repository.Member
	for rows.Next() {
		var m memberRow
		if err := rows.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, mapError(err)
		}
		out = append(out, toRepoMember(m))
	}
	return out, rows.Err()
}

func (r *MemberRepository) GetMemberByID(ctx context.Context, id string) (repository.Member, error) {
	return queryMemberByID(ctx, r.db, id)
}

func (r *MemberRepository) GetMemberByUsername(ctx context.Context, username string) (repository.Member, error) {
	return queryMemberByUsername(ctx, r.db, username)
}

func (r *MemberRepository) CreateMember(ctx context.Context, arg repository.CreateMemberParams) (repository.Member, error) {
	displayName := pgtype.Text{}
	if arg.DisplayName != nil {
		displayName = pgtype.Text{String: *arg.DisplayName, Valid: true}
	}
	row := r.db.QueryRow(ctx,
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

func (r *MemberRepository) PatchMember(ctx context.Context, arg repository.PatchMemberParams) (repository.Member, error) {
	uid, err := parseUUID(arg.ID)
	if err != nil {
		return repository.Member{}, err
	}

	setClauses := []string{}
	args := []any{uid}
	i := 2

	if arg.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", i))
		args = append(args, *arg.Username)
		i++
	}
	if arg.SetDisplayNameNull {
		setClauses = append(setClauses, "display_name = NULL")
	} else if arg.DisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", i))
		args = append(args, *arg.DisplayName)
		i++
	}
	if arg.Role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role_id = $%d", i))
		args = append(args, *arg.Role)
		i++
	}

	if len(setClauses) == 0 {
		return r.GetMemberByID(ctx, arg.ID)
	}

	query := fmt.Sprintf(
		`UPDATE members SET %s, updated_at = now() WHERE id = $1
		RETURNING id, username, display_name, password_hash, role_id, created_at, updated_at`,
		strings.Join(setClauses, ", "),
	)

	row := r.db.QueryRow(ctx, query, args...)
	var m memberRow
	if err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return repository.Member{}, mapError(err)
	}
	return toRepoMember(m), nil
}

func (r *MemberRepository) DeleteMember(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	row := r.db.QueryRow(ctx, `DELETE FROM members WHERE id = $1 RETURNING id`, uid)
	var deleted pgtype.UUID
	return mapError(row.Scan(&deleted))
}
