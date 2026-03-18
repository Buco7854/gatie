package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatie-io/gatie-server/internal/repository"
	"github.com/gatie-io/gatie-server/internal/service"
)

type RoleRepository struct{ base }

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{base{db: pool, pool: pool}}
}

func (r *RoleRepository) BeginTx(ctx context.Context) (service.RoleRepository, error) {
	b, err := r.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	return &RoleRepository{b}, nil
}

// --- Roles ---

func (r *RoleRepository) ListRoles(ctx context.Context) ([]repository.Role, error) {
	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.description, COALESCE(array_agg(rp.permission_id ORDER BY rp.permission_id) FILTER (WHERE rp.permission_id IS NOT NULL), '{}')
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		GROUP BY r.id
		ORDER BY r.id`,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var out []repository.Role
	for rows.Next() {
		var role repository.Role
		if err := rows.Scan(&role.ID, &role.Description, &role.Permissions); err != nil {
			return nil, mapError(err)
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

func (r *RoleRepository) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	return queryRolePermissions(ctx, r.db, roleID)
}

func (r *RoleRepository) CreateRole(ctx context.Context, id, description string) (repository.Role, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO roles (id, description) VALUES ($1, $2) RETURNING id, description`,
		id, description,
	)
	var role repository.Role
	if err := row.Scan(&role.ID, &role.Description); err != nil {
		return repository.Role{}, mapError(err)
	}
	role.Permissions = []string{}
	return role, nil
}

func (r *RoleRepository) UpdateRole(ctx context.Context, id, description string) (repository.Role, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE roles SET description = $2 WHERE id = $1 RETURNING id, description`,
		id, description,
	)
	var role repository.Role
	if err := row.Scan(&role.ID, &role.Description); err != nil {
		return repository.Role{}, mapError(err)
	}
	return role, nil
}

func (r *RoleRepository) DeleteRole(ctx context.Context, id string) error {
	row := r.db.QueryRow(ctx, `DELETE FROM roles WHERE id = $1 RETURNING id`, id)
	var deleted string
	return mapError(row.Scan(&deleted))
}

func (r *RoleRepository) RoleInUse(ctx context.Context, roleID string) (bool, error) {
	row := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM members WHERE role_id = $1
			UNION ALL
			SELECT 1 FROM gate_memberships WHERE role_id = $1
		)`, roleID,
	)
	var inUse bool
	if err := row.Scan(&inUse); err != nil {
		return false, mapError(err)
	}
	return inUse, nil
}

func (r *RoleRepository) DeleteRolePermissions(ctx context.Context, roleID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	return mapError(err)
}

func (r *RoleRepository) AddRolePermission(ctx context.Context, roleID, permissionID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
		roleID, permissionID,
	)
	return mapError(err)
}

// --- Permissions ---

func (r *RoleRepository) ListPermissions(ctx context.Context) ([]repository.Permission, error) {
	rows, err := r.db.Query(ctx, `SELECT id, description FROM permissions ORDER BY id`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var out []repository.Permission
	for rows.Next() {
		var p repository.Permission
		if err := rows.Scan(&p.ID, &p.Description); err != nil {
			return nil, mapError(err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *RoleRepository) CreatePermission(ctx context.Context, id, description string) (repository.Permission, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO permissions (id, description) VALUES ($1, $2) RETURNING id, description`,
		id, description,
	)
	var p repository.Permission
	if err := row.Scan(&p.ID, &p.Description); err != nil {
		return repository.Permission{}, mapError(err)
	}
	return p, nil
}

func (r *RoleRepository) UpdatePermission(ctx context.Context, id, description string) (repository.Permission, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE permissions SET description = $2 WHERE id = $1 RETURNING id, description`,
		id, description,
	)
	var p repository.Permission
	if err := row.Scan(&p.ID, &p.Description); err != nil {
		return repository.Permission{}, mapError(err)
	}
	return p, nil
}

func (r *RoleRepository) DeletePermission(ctx context.Context, id string) error {
	row := r.db.QueryRow(ctx, `DELETE FROM permissions WHERE id = $1 RETURNING id`, id)
	var deleted string
	return mapError(row.Scan(&deleted))
}

func (r *RoleRepository) PermissionInUse(ctx context.Context, permissionID string) (bool, error) {
	row := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM role_permissions WHERE permission_id = $1)`,
		permissionID,
	)
	var inUse bool
	if err := row.Scan(&inUse); err != nil {
		return false, mapError(err)
	}
	return inUse, nil
}
