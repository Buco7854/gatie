package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/gatie-io/gatie-server/internal/service"
)

type GateMembershipServicer interface {
	ListGateMembers(ctx context.Context, gateID string) ([]service.GateMember, error)
	AddGateMember(ctx context.Context, input service.CreateGateMembershipInput) (*service.GateMember, error)
	UpdateGateMember(ctx context.Context, gateID, memberID string, input service.UpdateGateMembershipInput) (*service.GateMember, error)
	RemoveGateMember(ctx context.Context, gateID, memberID string) error
}

type GateMembershipHandler struct {
	svc        GateMembershipServicer
	middlewares huma.Middlewares
}

func NewGateMembershipHandler(svc GateMembershipServicer, authMW, adminMW func(huma.Context, func(huma.Context))) *GateMembershipHandler {
	return &GateMembershipHandler{
		svc:        svc,
		middlewares: huma.Middlewares{authMW, adminMW},
	}
}

// --- DTOs ---

type GateMemberBody struct {
	GateID      string  `json:"gate_id" doc:"Gate UUID"`
	MemberID    string  `json:"member_id" doc:"Member UUID"`
	Username    string  `json:"username" doc:"Member username"`
	DisplayName *string `json:"display_name,omitempty" doc:"Member display name"`
	RoleID      string  `json:"role_id" doc:"Role assigned on this gate"`
	CreatedAt   string  `json:"created_at" doc:"Assignment date (ISO 8601)"`
}

// --- Inputs ---

type ListGateMembersInput struct {
	GateID string `path:"gate-id" doc:"Gate UUID"`
}

type AddGateMemberInput struct {
	GateID string `path:"gate-id" doc:"Gate UUID"`
	Body   struct {
		MemberID string `json:"member_id" format:"uuid" doc:"Member UUID"`
		RoleID   string `json:"role_id" minLength:"1" maxLength:"20" doc:"Role ID to assign"`
	}
}

type UpdateGateMemberInput struct {
	GateID   string `path:"gate-id" doc:"Gate UUID"`
	MemberID string `path:"member-id" doc:"Member UUID"`
	Body     struct {
		RoleID string `json:"role_id" minLength:"1" maxLength:"20" doc:"New role ID"`
	}
}

type RemoveGateMemberInput struct {
	GateID   string `path:"gate-id" doc:"Gate UUID"`
	MemberID string `path:"member-id" doc:"Member UUID"`
}

// --- Outputs ---

type ListGateMembersOutput struct {
	Body []GateMemberBody
}

type GateMemberOutput struct {
	Body GateMemberBody
}

// --- Registration ---

func (h *GateMembershipHandler) Register(api huma.API) {
	sec := []map[string][]string{{"bearer": {}}}

	huma.Register(api, huma.Operation{
		OperationID: "list-gate-members",
		Method:      http.MethodGet,
		Path:        "/api/gates/{gate-id}/members",
		Summary:     "List gate members",
		Tags:        []string{"Gate Members"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.listGateMembers)

	huma.Register(api, huma.Operation{
		OperationID:   "add-gate-member",
		Method:        http.MethodPost,
		Path:          "/api/gates/{gate-id}/members",
		Summary:       "Add member to gate",
		Tags:          []string{"Gate Members"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusCreated,
	}, h.addGateMember)

	huma.Register(api, huma.Operation{
		OperationID: "update-gate-member",
		Method:      http.MethodPatch,
		Path:        "/api/gates/{gate-id}/members/{member-id}",
		Summary:     "Update gate member role",
		Tags:        []string{"Gate Members"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.updateGateMember)

	huma.Register(api, huma.Operation{
		OperationID:   "remove-gate-member",
		Method:        http.MethodDelete,
		Path:          "/api/gates/{gate-id}/members/{member-id}",
		Summary:       "Remove member from gate",
		Tags:          []string{"Gate Members"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusNoContent,
	}, h.removeGateMember)
}

// --- Handlers ---

func (h *GateMembershipHandler) listGateMembers(ctx context.Context, input *ListGateMembersInput) (*ListGateMembersOutput, error) {
	members, err := h.svc.ListGateMembers(ctx, input.GateID)
	if err != nil {
		return nil, mapGateMembershipError(err)
	}

	items := make([]GateMemberBody, len(members))
	for i, m := range members {
		items[i] = toGateMemberBody(m)
	}

	return &ListGateMembersOutput{Body: items}, nil
}

func (h *GateMembershipHandler) addGateMember(ctx context.Context, input *AddGateMemberInput) (*GateMemberOutput, error) {
	gm, err := h.svc.AddGateMember(ctx, service.CreateGateMembershipInput{
		GateID:   input.GateID,
		MemberID: input.Body.MemberID,
		RoleID:   input.Body.RoleID,
	})
	if err != nil {
		return nil, mapGateMembershipError(err)
	}

	return &GateMemberOutput{Body: toGateMemberBody(*gm)}, nil
}

func (h *GateMembershipHandler) updateGateMember(ctx context.Context, input *UpdateGateMemberInput) (*GateMemberOutput, error) {
	gm, err := h.svc.UpdateGateMember(ctx, input.GateID, input.MemberID, service.UpdateGateMembershipInput{
		RoleID: input.Body.RoleID,
	})
	if err != nil {
		return nil, mapGateMembershipError(err)
	}

	return &GateMemberOutput{Body: toGateMemberBody(*gm)}, nil
}

func (h *GateMembershipHandler) removeGateMember(ctx context.Context, input *RemoveGateMemberInput) (*struct{}, error) {
	if err := h.svc.RemoveGateMember(ctx, input.GateID, input.MemberID); err != nil {
		return nil, mapGateMembershipError(err)
	}

	return nil, nil
}

// --- Helpers ---

func mapGateMembershipError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidID):
		return huma.Error400BadRequest("invalid id format")
	case errors.Is(err, service.ErrGateMembershipNotFound):
		return huma.Error404NotFound("gate membership not found")
	case errors.Is(err, service.ErrGateNotFound):
		return huma.Error404NotFound("gate not found")
	case errors.Is(err, service.ErrMemberNotFound):
		return huma.Error404NotFound("member not found")
	case errors.Is(err, service.ErrGateMembershipExists):
		return huma.Error409Conflict("member already has access to this gate")
	default:
		return huma.Error500InternalServerError("gate membership error", err)
	}
}

func toGateMemberBody(m service.GateMember) GateMemberBody {
	var displayName *string
	if m.DisplayName != "" {
		displayName = &m.DisplayName
	}
	return GateMemberBody{
		GateID:      m.GateID,
		MemberID:    m.MemberID,
		Username:    m.Username,
		DisplayName: displayName,
		RoleID:      m.RoleID,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
}
