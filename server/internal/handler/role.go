package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/gatie-io/gatie-server/internal/service"
)

type RoleServicer interface {
	ListRoles(ctx context.Context) ([]service.Role, error)
	CreateRole(ctx context.Context, id, description string) (*service.Role, error)
	UpdateRole(ctx context.Context, id, description string) (*service.Role, error)
	DeleteRole(ctx context.Context, id string) error
	SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) (*service.Role, error)
	ListPermissions(ctx context.Context) ([]service.Permission, error)
	CreatePermission(ctx context.Context, id, description string) (*service.Permission, error)
	UpdatePermission(ctx context.Context, id, description string) (*service.Permission, error)
	DeletePermission(ctx context.Context, id string) error
}

type RoleHandler struct {
	roleService RoleServicer
	middlewares huma.Middlewares
}

func NewRoleHandler(roleService RoleServicer, authMW, adminMW func(huma.Context, func(huma.Context))) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		middlewares: huma.Middlewares{authMW, adminMW},
	}
}

// --- DTOs ---

type RoleBody struct {
	ID          string   `json:"id" doc:"Role ID"`
	Description string   `json:"description" doc:"Role description"`
	Permissions []string `json:"permissions" doc:"Permissions granted by this role"`
}

type PermissionBody struct {
	ID          string `json:"id" doc:"Permission ID"`
	Description string `json:"description" doc:"Permission description"`
}

// --- Inputs ---

type CreateRoleInput struct {
	Body struct {
		ID          string `json:"id" minLength:"1" maxLength:"20" pattern:"^[A-Z][A-Z0-9_]*$" doc:"Role ID (uppercase, e.g. OPERATOR)"`
		Description string `json:"description" minLength:"1" maxLength:"500" doc:"Role description"`
	}
}

type UpdateRoleInput struct {
	RoleID string `path:"role-id" doc:"Role ID"`
	Body   struct {
		Description string `json:"description" minLength:"1" maxLength:"500" doc:"New description"`
	}
}

type DeleteRoleInput struct {
	RoleID string `path:"role-id" doc:"Role ID"`
}

type SetRolePermissionsInput struct {
	RoleID string `path:"role-id" doc:"Role ID"`
	Body   struct {
		Permissions []string `json:"permissions" doc:"Permission IDs to assign"`
	}
}

type CreatePermissionInput struct {
	Body struct {
		ID          string `json:"id" minLength:"1" maxLength:"50" pattern:"^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$" doc:"Permission ID (format namespace:action, e.g. gate:view_logs)"`
		Description string `json:"description" minLength:"1" maxLength:"500" doc:"Permission description"`
	}
}

type UpdatePermissionInput struct {
	PermissionID string `path:"permission-id" doc:"Permission ID"`
	Body         struct {
		Description string `json:"description" minLength:"1" maxLength:"500" doc:"New description"`
	}
}

type DeletePermissionInput struct {
	PermissionID string `path:"permission-id" doc:"Permission ID"`
}

// --- Outputs ---

type ListRolesOutput struct {
	Body []RoleBody
}

type RoleOutput struct {
	Body RoleBody
}

type ListPermissionsOutput struct {
	Body []PermissionBody
}

type PermissionOutput struct {
	Body PermissionBody
}

// --- Registration ---

func (h *RoleHandler) Register(api huma.API) {
	sec := []map[string][]string{{"bearer": {}}}

	huma.Register(api, huma.Operation{
		OperationID: "list-roles",
		Method:      http.MethodGet,
		Path:        "/api/roles",
		Summary:     "List roles",
		Tags:        []string{"Roles"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.listRoles)

	huma.Register(api, huma.Operation{
		OperationID:   "create-role",
		Method:        http.MethodPost,
		Path:          "/api/roles",
		Summary:       "Create role",
		Tags:          []string{"Roles"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusCreated,
	}, h.createRole)

	huma.Register(api, huma.Operation{
		OperationID: "update-role",
		Method:      http.MethodPatch,
		Path:        "/api/roles/{role-id}",
		Summary:     "Update role",
		Tags:        []string{"Roles"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.updateRole)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-role",
		Method:        http.MethodDelete,
		Path:          "/api/roles/{role-id}",
		Summary:       "Delete role",
		Tags:          []string{"Roles"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusNoContent,
	}, h.deleteRole)

	huma.Register(api, huma.Operation{
		OperationID: "set-role-permissions",
		Method:      http.MethodPut,
		Path:        "/api/roles/{role-id}/permissions",
		Summary:     "Set role permissions",
		Tags:        []string{"Roles"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.setRolePermissions)

	huma.Register(api, huma.Operation{
		OperationID: "list-permissions",
		Method:      http.MethodGet,
		Path:        "/api/permissions",
		Summary:     "List permissions",
		Tags:        []string{"Permissions"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.listPermissions)

	huma.Register(api, huma.Operation{
		OperationID:   "create-permission",
		Method:        http.MethodPost,
		Path:          "/api/permissions",
		Summary:       "Create permission",
		Tags:          []string{"Permissions"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusCreated,
	}, h.createPermission)

	huma.Register(api, huma.Operation{
		OperationID: "update-permission",
		Method:      http.MethodPatch,
		Path:        "/api/permissions/{permission-id}",
		Summary:     "Update permission",
		Tags:        []string{"Permissions"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.updatePermission)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-permission",
		Method:        http.MethodDelete,
		Path:          "/api/permissions/{permission-id}",
		Summary:       "Delete permission",
		Tags:          []string{"Permissions"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusNoContent,
	}, h.deletePermission)
}

// --- Handlers ---

func (h *RoleHandler) listRoles(ctx context.Context, _ *struct{}) (*ListRolesOutput, error) {
	roles, err := h.roleService.ListRoles(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list roles", err)
	}
	items := make([]RoleBody, len(roles))
	for i, r := range roles {
		items[i] = RoleBody{ID: r.ID, Description: r.Description, Permissions: r.Permissions}
	}
	return &ListRolesOutput{Body: items}, nil
}

func (h *RoleHandler) createRole(ctx context.Context, input *CreateRoleInput) (*RoleOutput, error) {
	role, err := h.roleService.CreateRole(ctx, input.Body.ID, input.Body.Description)
	if err != nil {
		return nil, mapRoleError(err)
	}
	return &RoleOutput{Body: RoleBody{ID: role.ID, Description: role.Description, Permissions: role.Permissions}}, nil
}

func (h *RoleHandler) updateRole(ctx context.Context, input *UpdateRoleInput) (*RoleOutput, error) {
	role, err := h.roleService.UpdateRole(ctx, input.RoleID, input.Body.Description)
	if err != nil {
		return nil, mapRoleError(err)
	}
	return &RoleOutput{Body: RoleBody{ID: role.ID, Description: role.Description, Permissions: role.Permissions}}, nil
}

func (h *RoleHandler) deleteRole(ctx context.Context, input *DeleteRoleInput) (*struct{}, error) {
	if err := h.roleService.DeleteRole(ctx, input.RoleID); err != nil {
		return nil, mapRoleError(err)
	}
	return nil, nil
}

func (h *RoleHandler) setRolePermissions(ctx context.Context, input *SetRolePermissionsInput) (*RoleOutput, error) {
	role, err := h.roleService.SetRolePermissions(ctx, input.RoleID, input.Body.Permissions)
	if err != nil {
		return nil, mapRoleError(err)
	}
	return &RoleOutput{Body: RoleBody{ID: role.ID, Description: role.Description, Permissions: role.Permissions}}, nil
}

func (h *RoleHandler) listPermissions(ctx context.Context, _ *struct{}) (*ListPermissionsOutput, error) {
	perms, err := h.roleService.ListPermissions(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list permissions", err)
	}
	items := make([]PermissionBody, len(perms))
	for i, p := range perms {
		items[i] = PermissionBody{ID: p.ID, Description: p.Description}
	}
	return &ListPermissionsOutput{Body: items}, nil
}

func (h *RoleHandler) createPermission(ctx context.Context, input *CreatePermissionInput) (*PermissionOutput, error) {
	perm, err := h.roleService.CreatePermission(ctx, input.Body.ID, input.Body.Description)
	if err != nil {
		return nil, mapPermissionError(err)
	}
	return &PermissionOutput{Body: PermissionBody{ID: perm.ID, Description: perm.Description}}, nil
}

func (h *RoleHandler) updatePermission(ctx context.Context, input *UpdatePermissionInput) (*PermissionOutput, error) {
	perm, err := h.roleService.UpdatePermission(ctx, input.PermissionID, input.Body.Description)
	if err != nil {
		return nil, mapPermissionError(err)
	}
	return &PermissionOutput{Body: PermissionBody{ID: perm.ID, Description: perm.Description}}, nil
}

func (h *RoleHandler) deletePermission(ctx context.Context, input *DeletePermissionInput) (*struct{}, error) {
	if err := h.roleService.DeletePermission(ctx, input.PermissionID); err != nil {
		return nil, mapPermissionError(err)
	}
	return nil, nil
}

// --- Error mapping ---

func mapRoleError(err error) error {
	switch {
	case errors.Is(err, service.ErrRoleProtected):
		return huma.Error422UnprocessableEntity("this role is protected")
	case errors.Is(err, service.ErrRoleInUse):
		return huma.Error422UnprocessableEntity("role is in use and cannot be deleted")
	case errors.Is(err, service.ErrRoleNotFound):
		return huma.Error404NotFound("role not found")
	case errors.Is(err, service.ErrRoleExists):
		return huma.Error409Conflict("role already exists")
	case errors.Is(err, service.ErrPermissionProtected):
		return huma.Error422UnprocessableEntity("the wildcard permission cannot be assigned to custom roles")
	default:
		return huma.Error500InternalServerError("role error", err)
	}
}

func mapPermissionError(err error) error {
	switch {
	case errors.Is(err, service.ErrPermissionProtected):
		return huma.Error422UnprocessableEntity("this permission is protected")
	case errors.Is(err, service.ErrPermissionInUse):
		return huma.Error422UnprocessableEntity("permission is in use and cannot be deleted")
	case errors.Is(err, service.ErrPermissionNotFound):
		return huma.Error404NotFound("permission not found")
	case errors.Is(err, service.ErrPermissionExists):
		return huma.Error409Conflict("permission already exists")
	default:
		return huma.Error500InternalServerError("permission error", err)
	}
}
