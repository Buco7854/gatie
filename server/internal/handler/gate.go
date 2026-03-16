package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/repository"
	"github.com/gatie-io/gatie-server/internal/service"
)

type GateHandler struct {
	gateService *service.GateService
	middlewares huma.Middlewares
}

func NewGateHandler(gateService *service.GateService, authMW, adminMW func(huma.Context, func(huma.Context))) *GateHandler {
	return &GateHandler{
		gateService: gateService,
		middlewares: huma.Middlewares{authMW, adminMW},
	}
}

// --- DTOs ---

type GateBody struct {
	ID               string `json:"id" doc:"Gate UUID"`
	Name             string `json:"name" doc:"Gate name"`
	StatusTTLSeconds int32  `json:"status_ttl_seconds" doc:"Status TTL in seconds"`
	CreatedAt        string `json:"created_at" doc:"Creation date (ISO 8601)"`
	UpdatedAt        string `json:"updated_at" doc:"Last update date (ISO 8601)"`
}

type GateWithTokenBody struct {
	GateBody
	Token string `json:"token" doc:"Plain gate token — save it now, it will never be shown again"`
}

type GatesPageBody struct {
	Items   []GateBody `json:"items" doc:"Gates"`
	Total   int64      `json:"total" doc:"Total number of gates"`
	Page    int        `json:"page" doc:"Current page"`
	PerPage int        `json:"per_page" doc:"Items per page"`
}

// --- Inputs ---

type ListGatesInput struct {
	Page    int `query:"page" minimum:"1" default:"1" doc:"Page number"`
	PerPage int `query:"per_page" minimum:"1" maximum:"100" default:"20" doc:"Items per page"`
}

type GetGateInput struct {
	GateID string `path:"gate-id" doc:"Gate UUID"`
}

type CreateGateBodyInput struct {
	Body struct {
		Name             string `json:"name" minLength:"1" maxLength:"100" doc:"Gate name"`
		StatusTTLSeconds int32  `json:"status_ttl_seconds,omitempty" minimum:"1" maximum:"86400" default:"60" doc:"Status TTL in seconds"`
	}
}

type UpdateGateInput struct {
	GateID string `path:"gate-id" doc:"Gate UUID"`
	Body   struct {
		Name             string `json:"name" minLength:"1" maxLength:"100" doc:"Gate name"`
		StatusTTLSeconds int32  `json:"status_ttl_seconds,omitempty" minimum:"1" maximum:"86400" default:"60" doc:"Status TTL in seconds"`
	}
}

type DeleteGateInput struct {
	GateID string `path:"gate-id" doc:"Gate UUID"`
}

type RegenerateTokenInput struct {
	GateID string `path:"gate-id" doc:"Gate UUID"`
}

// --- Outputs ---

type GateOutput struct {
	Body GateBody
}

type GateWithTokenOutput struct {
	Body GateWithTokenBody
}

type ListGatesOutput struct {
	Body GatesPageBody
}

// --- Registration ---

func (h *GateHandler) Register(api huma.API) {
	sec := []map[string][]string{{"bearer": {}}}

	huma.Register(api, huma.Operation{
		OperationID: "list-gates",
		Method:      http.MethodGet,
		Path:        "/api/gates",
		Summary:     "List gates",
		Tags:        []string{"Gates"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.listGates)

	huma.Register(api, huma.Operation{
		OperationID:   "create-gate",
		Method:        http.MethodPost,
		Path:          "/api/gates",
		Summary:       "Create gate",
		Tags:          []string{"Gates"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusCreated,
	}, h.createGate)

	huma.Register(api, huma.Operation{
		OperationID: "get-gate",
		Method:      http.MethodGet,
		Path:        "/api/gates/{gate-id}",
		Summary:     "Get gate",
		Tags:        []string{"Gates"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.getGate)

	huma.Register(api, huma.Operation{
		OperationID: "update-gate",
		Method:      http.MethodPatch,
		Path:        "/api/gates/{gate-id}",
		Summary:     "Update gate",
		Tags:        []string{"Gates"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.updateGate)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-gate",
		Method:        http.MethodDelete,
		Path:          "/api/gates/{gate-id}",
		Summary:       "Delete gate",
		Tags:          []string{"Gates"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusNoContent,
	}, h.deleteGate)

	huma.Register(api, huma.Operation{
		OperationID: "regenerate-gate-token",
		Method:      http.MethodPost,
		Path:        "/api/gates/{gate-id}/token",
		Summary:     "Regenerate gate token",
		Tags:        []string{"Gates"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.regenerateToken)
}

// --- Handlers ---

func (h *GateHandler) listGates(ctx context.Context, input *ListGatesInput) (*ListGatesOutput, error) {
	page, err := h.gateService.ListGates(ctx, input.Page, input.PerPage)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list gates")
	}

	items := make([]GateBody, len(page.Gates))
	for i, g := range page.Gates {
		items[i] = toGateBody(g)
	}

	return &ListGatesOutput{Body: GatesPageBody{
		Items:   items,
		Total:   page.Total,
		Page:    input.Page,
		PerPage: input.PerPage,
	}}, nil
}

func (h *GateHandler) createGate(ctx context.Context, input *CreateGateBodyInput) (*GateWithTokenOutput, error) {
	result, err := h.gateService.CreateGate(ctx, service.CreateGateInput{
		Name:             input.Body.Name,
		StatusTTLSeconds: input.Body.StatusTTLSeconds,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create gate")
	}

	return &GateWithTokenOutput{Body: GateWithTokenBody{
		GateBody: toGateBody(result.Gate),
		Token:    result.Token,
	}}, nil
}

func (h *GateHandler) getGate(ctx context.Context, input *GetGateInput) (*GateOutput, error) {
	id, err := parseGateUUID(input.GateID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid gate ID")
	}

	gate, err := h.gateService.GetGate(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrGateNotFound) {
			return nil, huma.Error404NotFound("gate not found")
		}
		return nil, huma.Error500InternalServerError("failed to get gate")
	}

	return &GateOutput{Body: toGateBody(*gate)}, nil
}

func (h *GateHandler) updateGate(ctx context.Context, input *UpdateGateInput) (*GateOutput, error) {
	id, err := parseGateUUID(input.GateID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid gate ID")
	}

	gate, err := h.gateService.UpdateGate(ctx, id, service.UpdateGateInput{
		Name:             input.Body.Name,
		StatusTTLSeconds: input.Body.StatusTTLSeconds,
	})
	if err != nil {
		if errors.Is(err, service.ErrGateNotFound) {
			return nil, huma.Error404NotFound("gate not found")
		}
		return nil, huma.Error500InternalServerError("failed to update gate")
	}

	return &GateOutput{Body: toGateBody(*gate)}, nil
}

func (h *GateHandler) deleteGate(ctx context.Context, input *DeleteGateInput) (*struct{}, error) {
	id, err := parseGateUUID(input.GateID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid gate ID")
	}

	if err := h.gateService.DeleteGate(ctx, id); err != nil {
		if errors.Is(err, service.ErrGateNotFound) {
			return nil, huma.Error404NotFound("gate not found")
		}
		return nil, huma.Error500InternalServerError("failed to delete gate")
	}

	return nil, nil
}

func (h *GateHandler) regenerateToken(ctx context.Context, input *RegenerateTokenInput) (*GateWithTokenOutput, error) {
	id, err := parseGateUUID(input.GateID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid gate ID")
	}

	result, err := h.gateService.RegenerateToken(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrGateNotFound) {
			return nil, huma.Error404NotFound("gate not found")
		}
		return nil, huma.Error500InternalServerError("failed to regenerate token")
	}

	return &GateWithTokenOutput{Body: GateWithTokenBody{
		GateBody: toGateBody(result.Gate),
		Token:    result.Token,
	}}, nil
}

// --- Helpers ---

func toGateBody(g repository.Gate) GateBody {
	return GateBody{
		ID:               uuidBytesToString(g.ID.Bytes),
		Name:             g.Name,
		StatusTTLSeconds: g.StatusTtlSeconds,
		CreatedAt:        g.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        g.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func parseGateUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}
