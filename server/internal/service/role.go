package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gatie-io/gatie-server/internal/repository"
)

type RoleRepository interface {
	BeginTx(ctx context.Context) (RoleRepository, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	ListRoles(ctx context.Context) ([]repository.Role, error)
	CreateRole(ctx context.Context, id, description string) (repository.Role, error)
	UpdateRole(ctx context.Context, id, description string) (repository.Role, error)
	DeleteRole(ctx context.Context, id string) error
	RoleInUse(ctx context.Context, roleID string) (bool, error)
	GetRolePermissions(ctx context.Context, roleID string) ([]string, error)
	DeleteRolePermissions(ctx context.Context, roleID string) error
	AddRolePermission(ctx context.Context, roleID, permissionID string) error
	ListPermissions(ctx context.Context) ([]repository.Permission, error)
	CreatePermission(ctx context.Context, id, description string) (repository.Permission, error)
	UpdatePermission(ctx context.Context, id, description string) (repository.Permission, error)
	DeletePermission(ctx context.Context, id string) error
	PermissionInUse(ctx context.Context, permissionID string) (bool, error)
}

var (
	ErrRoleProtected       = errors.New("this role is protected and cannot be modified")
	ErrPermissionProtected = errors.New("this permission is protected and cannot be modified")
	ErrRoleInUse           = errors.New("role is assigned to members or gate memberships")
	ErrPermissionInUse     = errors.New("permission is assigned to roles")
	ErrRoleNotFound        = errors.New("role not found")
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrRoleExists          = errors.New("role already exists")
	ErrPermissionExists    = errors.New("permission already exists")
)

type RoleService struct {
	repo RoleRepository
}

func NewRoleService(repo RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

type Role struct {
	ID          string
	Description string
	Permissions []string
}

type Permission struct {
	ID          string
	Description string
}

// --- Roles ---

func (s *RoleService) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing roles: %w", err)
	}

	roles := make([]Role, len(rows))
	for i, r := range rows {
		roles[i] = Role{
			ID:          r.ID,
			Description: r.Description,
			Permissions: r.Permissions,
		}
	}
	return roles, nil
}

func (s *RoleService) CreateRole(ctx context.Context, id, description string) (*Role, error) {
	if id == RoleAdmin {
		return nil, ErrRoleProtected
	}

	row, err := s.repo.CreateRole(ctx, id, description)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrRoleExists
		}
		return nil, fmt.Errorf("creating role: %w", err)
	}

	return &Role{ID: row.ID, Description: row.Description, Permissions: row.Permissions}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id, description string) (*Role, error) {
	if id == RoleAdmin {
		return nil, ErrRoleProtected
	}

	row, err := s.repo.UpdateRole(ctx, id, description)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("updating role: %w", err)
	}

	perms, err := s.repo.GetRolePermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching role permissions: %w", err)
	}

	return &Role{ID: row.ID, Description: row.Description, Permissions: perms}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, id string) error {
	if id == RoleAdmin {
		return ErrRoleProtected
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	inUse, err := tx.RoleInUse(ctx, id)
	if err != nil {
		return fmt.Errorf("checking role usage: %w", err)
	}
	if inUse {
		return ErrRoleInUse
	}

	if err := tx.DeleteRole(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("deleting role: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *RoleService) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) (*Role, error) {
	if roleID == RoleAdmin {
		return nil, ErrRoleProtected
	}

	for _, pid := range permissionIDs {
		if pid == PermAll {
			return nil, ErrPermissionProtected
		}
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := tx.DeleteRolePermissions(ctx, roleID); err != nil {
		return nil, fmt.Errorf("clearing role permissions: %w", err)
	}

	for _, pid := range permissionIDs {
		if err := tx.AddRolePermission(ctx, roleID, pid); err != nil {
			return nil, fmt.Errorf("adding permission %s: %w", pid, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("re-fetching roles: %w", err)
	}
	for _, r := range roles {
		if r.ID == roleID {
			return &Role{ID: r.ID, Description: r.Description, Permissions: r.Permissions}, nil
		}
	}
	return nil, ErrRoleNotFound
}

// --- Permissions ---

func (s *RoleService) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing permissions: %w", err)
	}

	perms := make([]Permission, len(rows))
	for i, p := range rows {
		perms[i] = Permission{ID: p.ID, Description: p.Description}
	}
	return perms, nil
}

func (s *RoleService) CreatePermission(ctx context.Context, id, description string) (*Permission, error) {
	if id == PermAll {
		return nil, ErrPermissionProtected
	}

	row, err := s.repo.CreatePermission(ctx, id, description)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrPermissionExists
		}
		return nil, fmt.Errorf("creating permission: %w", err)
	}

	return &Permission{ID: row.ID, Description: row.Description}, nil
}

func (s *RoleService) UpdatePermission(ctx context.Context, id, description string) (*Permission, error) {
	if id == PermAll {
		return nil, ErrPermissionProtected
	}

	row, err := s.repo.UpdatePermission(ctx, id, description)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("updating permission: %w", err)
	}

	return &Permission{ID: row.ID, Description: row.Description}, nil
}

func (s *RoleService) DeletePermission(ctx context.Context, id string) error {
	if id == PermAll {
		return ErrPermissionProtected
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	inUse, err := tx.PermissionInUse(ctx, id)
	if err != nil {
		return fmt.Errorf("checking permission usage: %w", err)
	}
	if inUse {
		return ErrPermissionInUse
	}

	if err := tx.DeletePermission(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("deleting permission: %w", err)
	}

	return tx.Commit(ctx)
}
