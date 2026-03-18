package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/repository"
)

type GateMembershipRepository struct{ base }

func NewGateMembershipRepository(pool *pgxpool.Pool) *GateMembershipRepository {
	return &GateMembershipRepository{base{db: pool, pool: pool}}
}

func (r *GateMembershipRepository) ListGateMemberships(ctx context.Context, gateID string) ([]repository.GateMembershipWithMember, error) {
	uid, err := parseUUID(gateID)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT gm.gate_id, gm.member_id, m.username, m.display_name, gm.role_id, gm.created_at
		FROM gate_memberships gm
		JOIN members m ON m.id = gm.member_id
		WHERE gm.gate_id = $1
		ORDER BY gm.created_at ASC`,
		uid,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var out []repository.GateMembershipWithMember
	for rows.Next() {
		var (
			gateUUID   pgtype.UUID
			memberUUID pgtype.UUID
			username   string
			dn         pgtype.Text
			roleID     string
			createdAt  pgtype.Timestamptz
		)
		if err := rows.Scan(&gateUUID, &memberUUID, &username, &dn, &roleID, &createdAt); err != nil {
			return nil, mapError(err)
		}
		var displayName *string
		if dn.Valid {
			displayName = &dn.String
		}
		out = append(out, repository.GateMembershipWithMember{
			GateID:      uuidToString(gateUUID),
			MemberID:    uuidToString(memberUUID),
			Username:    username,
			DisplayName: displayName,
			RoleID:      roleID,
			CreatedAt:   createdAt.Time,
		})
	}
	return out, rows.Err()
}

func (r *GateMembershipRepository) GetGateMembership(ctx context.Context, gateID, memberID string) (repository.GateMembership, error) {
	return queryGateMembership(ctx, r.db, gateID, memberID)
}

func (r *GateMembershipRepository) CreateGateMembership(ctx context.Context, arg repository.CreateGateMembershipParams) (repository.GateMembership, error) {
	gateUUID, err := parseUUID(arg.GateID)
	if err != nil {
		return repository.GateMembership{}, err
	}
	memberUUID, err := parseUUID(arg.MemberID)
	if err != nil {
		return repository.GateMembership{}, err
	}

	var gm gateMembershipRow
	row := r.db.QueryRow(ctx,
		`INSERT INTO gate_memberships (gate_id, member_id, role_id)
		VALUES ($1, $2, $3)
		RETURNING gate_id, member_id, role_id, created_at`,
		gateUUID, memberUUID, arg.RoleID,
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

func (r *GateMembershipRepository) UpdateGateMembership(ctx context.Context, arg repository.UpdateGateMembershipParams) (repository.GateMembership, error) {
	gateUUID, err := parseUUID(arg.GateID)
	if err != nil {
		return repository.GateMembership{}, err
	}
	memberUUID, err := parseUUID(arg.MemberID)
	if err != nil {
		return repository.GateMembership{}, err
	}

	var gm gateMembershipRow
	row := r.db.QueryRow(ctx,
		`UPDATE gate_memberships SET role_id = $3
		WHERE gate_id = $1 AND member_id = $2
		RETURNING gate_id, member_id, role_id, created_at`,
		gateUUID, memberUUID, arg.RoleID,
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

func (r *GateMembershipRepository) DeleteGateMembership(ctx context.Context, gateID, memberID string) error {
	gateUUID, err := parseUUID(gateID)
	if err != nil {
		return err
	}
	memberUUID, err := parseUUID(memberID)
	if err != nil {
		return err
	}

	row := r.db.QueryRow(ctx,
		`DELETE FROM gate_memberships WHERE gate_id = $1 AND member_id = $2 RETURNING gate_id`,
		gateUUID, memberUUID,
	)
	var deleted pgtype.UUID
	return mapError(row.Scan(&deleted))
}

func (r *GateMembershipRepository) GetGateByID(ctx context.Context, id string) (repository.Gate, error) {
	return queryGateByID(ctx, r.db, id)
}

func (r *GateMembershipRepository) GetMemberByID(ctx context.Context, id string) (repository.Member, error) {
	return queryMemberByID(ctx, r.db, id)
}
