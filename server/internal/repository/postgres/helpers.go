package postgres

import (
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
		Role:         m.Role,
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

func toRepoRefreshToken(rt refreshTokenRow) repository.RefreshToken {
	return repository.RefreshToken{
		ID:        uuidToString(rt.ID),
		MemberID:  uuidToString(rt.MemberID),
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt.Time,
		CreatedAt: rt.CreatedAt.Time,
	}
}
