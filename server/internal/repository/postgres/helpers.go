package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/repository"
)

func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, repository.ErrInvalidID
	}
	return id, nil
}

func uuidToString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func toRepoMember(m memberRow) repository.Member {
	var displayName *string
	if m.DisplayName.Valid {
		displayName = &m.DisplayName.String
	}
	return repository.Member{
		ID:           uuidToString(m.ID),
		Username:     m.Username,
		DisplayName:  displayName,
		PasswordHash: m.PasswordHash,
		Role:         m.RoleID,
		CreatedAt:    m.CreatedAt.Time,
		UpdatedAt:    m.UpdatedAt.Time,
	}
}

func toRepoGate(g gateRow) repository.Gate {
	return repository.Gate{
		ID:               uuidToString(g.ID),
		Name:             g.Name,
		GateTokenHash:    g.GateTokenHash,
		StatusTtlSeconds: g.StatusTtlSeconds,
		CreatedAt:        g.CreatedAt.Time,
		UpdatedAt:        g.UpdatedAt.Time,
	}
}

func queryMemberByID(ctx context.Context, db DBTX, id string) (repository.Member, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return repository.Member{}, err
	}
	row := db.QueryRow(ctx,
		`SELECT id, username, display_name, password_hash, role_id, created_at, updated_at
		FROM members WHERE id = $1`, uid,
	)
	var m memberRow
	if err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return repository.Member{}, mapError(err)
	}
	return toRepoMember(m), nil
}

func queryGateByID(ctx context.Context, db DBTX, id string) (repository.Gate, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return repository.Gate{}, err
	}
	row := db.QueryRow(ctx,
		`SELECT id, name, gate_token_hash, status_ttl_seconds, created_at, updated_at
		FROM gates WHERE id = $1`, uid,
	)
	var g gateRow
	if err := row.Scan(&g.ID, &g.Name, &g.GateTokenHash, &g.StatusTtlSeconds, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return repository.Gate{}, mapError(err)
	}
	return toRepoGate(g), nil
}

func queryMemberByUsername(ctx context.Context, db DBTX, username string) (repository.Member, error) {
	row := db.QueryRow(ctx,
		`SELECT id, username, display_name, password_hash, role_id, created_at, updated_at
		FROM members WHERE username = $1`, username,
	)
	var m memberRow
	if err := row.Scan(&m.ID, &m.Username, &m.DisplayName, &m.PasswordHash, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return repository.Member{}, mapError(err)
	}
	return toRepoMember(m), nil
}

func queryGateMembership(ctx context.Context, db DBTX, gateID, memberID string) (repository.GateMembership, error) {
	gateUUID, err := parseUUID(gateID)
	if err != nil {
		return repository.GateMembership{}, err
	}
	memberUUID, err := parseUUID(memberID)
	if err != nil {
		return repository.GateMembership{}, err
	}
	var gm gateMembershipRow
	row := db.QueryRow(ctx,
		`SELECT gate_id, member_id, role_id, created_at
		FROM gate_memberships WHERE gate_id = $1 AND member_id = $2`,
		gateUUID, memberUUID,
	)
	if err := row.Scan(&gm.GateID, &gm.MemberID, &gm.RoleID, &gm.CreatedAt); err != nil {
		return repository.GateMembership{}, mapError(err)
	}
	return repository.GateMembership{
		GateID:    uuidToString(gm.GateID),
		MemberID:  uuidToString(gm.MemberID),
		RoleID:    gm.RoleID,
		CreatedAt: gm.CreatedAt.Time,
	}, nil
}

func queryRolePermissions(ctx context.Context, db DBTX, roleID string) ([]string, error) {
	rows, err := db.Query(ctx,
		`SELECT permission_id FROM role_permissions WHERE role_id = $1`,
		roleID,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, mapError(err)
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func toRepoRefreshToken(rt refreshTokenRow) repository.RefreshToken {
	return repository.RefreshToken{
		ID:        uuidToString(rt.ID),
		MemberID:  uuidToString(rt.MemberID),
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt.Time,
		CreatedAt: rt.CreatedAt.Time,
	}
}
