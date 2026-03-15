package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gatie-io/gatie-server/internal/middleware"
	"github.com/gatie-io/gatie-server/internal/repository"
	"github.com/gatie-io/gatie-server/internal/service"
)

type MemberHandler struct {
	memberService *service.MemberService
	middlewares   huma.Middlewares
}

func NewMemberHandler(memberService *service.MemberService, authMW, adminMW func(huma.Context, func(huma.Context))) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
		middlewares:   huma.Middlewares{authMW, adminMW},
	}
}

// --- DTOs ---

type MemberBody struct {
	ID          string  `json:"id" doc:"Member UUID"`
	Username    string  `json:"username" doc:"Username"`
	DisplayName *string `json:"display_name,omitempty" doc:"Display name"`
	Role        string  `json:"role" doc:"Role"`
	CreatedAt   string  `json:"created_at" doc:"Creation date (ISO 8601)"`
	UpdatedAt   string  `json:"updated_at" doc:"Last update date (ISO 8601)"`
}

type MembersPageBody struct {
	Items   []MemberBody `json:"items" doc:"Members"`
	Total   int64        `json:"total" doc:"Total number of members"`
	Page    int          `json:"page" doc:"Current page"`
	PerPage int          `json:"per_page" doc:"Items per page"`
}

// --- Inputs ---

type ListMembersInput struct {
	Page    int `query:"page" minimum:"1" default:"1" doc:"Page number"`
	PerPage int `query:"per_page" minimum:"1" maximum:"100" default:"20" doc:"Items per page"`
}

type GetMemberInput struct {
	MemberID string `path:"member-id" doc:"Member UUID"`
}

type CreateMemberBodyInput struct {
	Body struct {
		Username    string  `json:"username" minLength:"3" maxLength:"100" doc:"Username"`
		DisplayName *string `json:"display_name,omitempty" maxLength:"200" doc:"Display name"`
		Password    string  `json:"password" minLength:"8" maxLength:"128" doc:"Password"`
		Role        string  `json:"role" enum:"ADMIN,MEMBER" doc:"Role"`
	}
}

type UpdateMemberInput struct {
	MemberID string `path:"member-id" doc:"Member UUID"`
	Body     struct {
		Username    string  `json:"username" minLength:"3" maxLength:"100" doc:"Username"`
		DisplayName *string `json:"display_name,omitempty" maxLength:"200" doc:"Display name"`
		Role        string  `json:"role" enum:"ADMIN,MEMBER" doc:"Role"`
	}
}

type DeleteMemberInput struct {
	MemberID string `path:"member-id" doc:"Member UUID"`
}

// --- Outputs ---

type MemberOutput struct {
	Body MemberBody
}

type ListMembersOutput struct {
	Body MembersPageBody
}

// --- Registration ---

func (h *MemberHandler) Register(api huma.API) {
	sec := []map[string][]string{{"bearer": {}}}

	huma.Register(api, huma.Operation{
		OperationID: "list-members",
		Method:      http.MethodGet,
		Path:        "/api/members",
		Summary:     "List members",
		Tags:        []string{"Members"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.listMembers)

	huma.Register(api, huma.Operation{
		OperationID:   "create-member",
		Method:        http.MethodPost,
		Path:          "/api/members",
		Summary:       "Create member",
		Tags:          []string{"Members"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusCreated,
	}, h.createMember)

	huma.Register(api, huma.Operation{
		OperationID: "get-member",
		Method:      http.MethodGet,
		Path:        "/api/members/{member-id}",
		Summary:     "Get member",
		Tags:        []string{"Members"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.getMember)

	huma.Register(api, huma.Operation{
		OperationID: "update-member",
		Method:      http.MethodPatch,
		Path:        "/api/members/{member-id}",
		Summary:     "Update member",
		Tags:        []string{"Members"},
		Security:    sec,
		Middlewares: h.middlewares,
	}, h.updateMember)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-member",
		Method:        http.MethodDelete,
		Path:          "/api/members/{member-id}",
		Summary:       "Delete member",
		Tags:          []string{"Members"},
		Security:      sec,
		Middlewares:   h.middlewares,
		DefaultStatus: http.StatusNoContent,
	}, h.deleteMember)
}

// --- Handlers ---

func (h *MemberHandler) listMembers(ctx context.Context, input *ListMembersInput) (*ListMembersOutput, error) {
	page, err := h.memberService.ListMembers(ctx, input.Page, input.PerPage)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list members")
	}

	items := make([]MemberBody, len(page.Members))
	for i, m := range page.Members {
		items[i] = toMemberBody(m)
	}

	return &ListMembersOutput{Body: MembersPageBody{
		Items:   items,
		Total:   page.Total,
		Page:    input.Page,
		PerPage: input.PerPage,
	}}, nil
}

func (h *MemberHandler) createMember(ctx context.Context, input *CreateMemberBodyInput) (*MemberOutput, error) {
	displayName := ""
	if input.Body.DisplayName != nil {
		displayName = *input.Body.DisplayName
	}

	member, err := h.memberService.CreateMember(ctx, service.CreateMemberInput{
		Username:    input.Body.Username,
		DisplayName: displayName,
		Password:    input.Body.Password,
		Role:        input.Body.Role,
	})
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			return nil, huma.Error422UnprocessableEntity("username already taken", &huma.ErrorDetail{
				Location: "body.username",
				Message:  "username already taken",
				Value:    input.Body.Username,
			})
		}
		return nil, huma.Error500InternalServerError("failed to create member")
	}

	return &MemberOutput{Body: toMemberBody(*member)}, nil
}

func (h *MemberHandler) getMember(ctx context.Context, input *GetMemberInput) (*MemberOutput, error) {
	id, err := parseUUID(input.MemberID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid member ID")
	}

	member, err := h.memberService.GetMember(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrMemberNotFound) {
			return nil, huma.Error404NotFound("member not found")
		}
		return nil, huma.Error500InternalServerError("failed to get member")
	}

	return &MemberOutput{Body: toMemberBody(*member)}, nil
}

func (h *MemberHandler) updateMember(ctx context.Context, input *UpdateMemberInput) (*MemberOutput, error) {
	id, err := parseUUID(input.MemberID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid member ID")
	}

	claims := middleware.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	displayName := ""
	if input.Body.DisplayName != nil {
		displayName = *input.Body.DisplayName
	}

	member, err := h.memberService.UpdateMember(ctx, id, service.UpdateMemberInput{
		Username:    input.Body.Username,
		DisplayName: displayName,
		Role:        input.Body.Role,
		CallerID:    claims.MemberID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMemberNotFound):
			return nil, huma.Error404NotFound("member not found")
		case errors.Is(err, service.ErrSelfRoleChange):
			return nil, huma.Error422UnprocessableEntity("cannot change your own role")
		case errors.Is(err, service.ErrUsernameExists):
			return nil, huma.Error422UnprocessableEntity("username already taken", &huma.ErrorDetail{
				Location: "body.username",
				Message:  "username already taken",
				Value:    input.Body.Username,
			})
		}
		return nil, huma.Error500InternalServerError("failed to update member")
	}

	return &MemberOutput{Body: toMemberBody(*member)}, nil
}

func (h *MemberHandler) deleteMember(ctx context.Context, input *DeleteMemberInput) (*struct{}, error) {
	id, err := parseUUID(input.MemberID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid member ID")
	}

	claims := middleware.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	err = h.memberService.DeleteMember(ctx, id, claims.MemberID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMemberNotFound):
			return nil, huma.Error404NotFound("member not found")
		case errors.Is(err, service.ErrSelfDelete):
			return nil, huma.Error422UnprocessableEntity("cannot delete your own account")
		case errors.Is(err, service.ErrLastAdmin):
			return nil, huma.Error422UnprocessableEntity("cannot delete the last admin")
		default:
			return nil, huma.Error500InternalServerError("failed to delete member")
		}
	}

	return nil, nil
}

// --- Helpers ---

func toMemberBody(m repository.Member) MemberBody {
	var displayName *string
	if m.DisplayName.Valid {
		displayName = &m.DisplayName.String
	}
	return MemberBody{
		ID:          uuidBytesToString(m.ID.Bytes),
		Username:    m.Username,
		DisplayName: displayName,
		Role:        m.Role,
		CreatedAt:   m.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}
