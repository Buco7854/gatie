package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/gatie-io/gatie-server/internal/service"
)

type AuthServicer interface {
	NeedsSetup(ctx context.Context) (bool, error)
	Setup(ctx context.Context, input service.SetupInput) (*service.AuthResult, error)
	Login(ctx context.Context, input service.LoginInput) (*service.AuthResult, error)
	Refresh(ctx context.Context, rawToken string) (*service.AuthResult, error)
	Logout(ctx context.Context, rawToken string) error
}

type AuthHandler struct {
	authService AuthServicer
	rateLimitMW func(huma.Context, func(huma.Context))
}

func NewAuthHandler(authService AuthServicer, rateLimitMW func(huma.Context, func(huma.Context))) *AuthHandler {
	return &AuthHandler{authService: authService, rateLimitMW: rateLimitMW}
}

// --- Setup ---

type SetupInput struct {
	Body struct {
		Username string `json:"username" minLength:"3" maxLength:"100" doc:"Admin username"`
		Password string `json:"password" minLength:"8" maxLength:"128" doc:"Admin password"`
	}
}

type AuthTokenBody struct {
	AccessToken  string `json:"access_token" doc:"JWT access token"`
	RefreshToken string `json:"refresh_token" doc:"Opaque refresh token"`
	MemberID     string `json:"member_id" doc:"Member UUID"`
	Role         string `json:"role" doc:"Member role"`
	Username     string `json:"username" doc:"Member username"`
}

type AuthTokenOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      AuthTokenBody
}

type SetupStatusBody struct {
	NeedsSetup bool `json:"needs_setup" doc:"Whether initial setup is required"`
}

type SetupStatusOutput struct {
	Body SetupStatusBody
}

func (h *AuthHandler) Register(api huma.API) {
	rl := huma.Middlewares{h.rateLimitMW}

	huma.Register(api, huma.Operation{
		OperationID: "setup-status",
		Method:      http.MethodGet,
		Path:        "/api/setup/status",
		Summary:     "Setup status",
		Description: "Check whether initial setup is needed.",
		Tags:        []string{"Auth"},
	}, h.setupStatus)

	huma.Register(api, huma.Operation{
		OperationID:   "setup",
		Method:        http.MethodPost,
		Path:          "/api/setup",
		Summary:       "Initial setup",
		Description:   "Create the first admin account. Only available when no members exist.",
		Tags:          []string{"Auth"},
		DefaultStatus: http.StatusCreated,
		Middlewares:   rl,
	}, h.setup)

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "Login",
		Description: "Authenticate with username and password.",
		Tags:        []string{"Auth"},
		Middlewares: rl,
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "refresh",
		Method:      http.MethodPost,
		Path:        "/api/auth/refresh",
		Summary:     "Refresh tokens",
		Description: "Get a new access token using a refresh token.",
		Tags:        []string{"Auth"},
		Middlewares: rl,
	}, h.refresh)

	huma.Register(api, huma.Operation{
		OperationID:   "logout",
		Method:        http.MethodPost,
		Path:          "/api/auth/logout",
		Summary:       "Logout",
		Description:   "Revoke the refresh token.",
		Tags:          []string{"Auth"},
		DefaultStatus: http.StatusNoContent,
	}, h.logout)
}

func (h *AuthHandler) setupStatus(ctx context.Context, input *struct{}) (*SetupStatusOutput, error) {
	needed, err := h.authService.NeedsSetup(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to check setup status", err)
	}
	return &SetupStatusOutput{Body: SetupStatusBody{NeedsSetup: needed}}, nil
}

func (h *AuthHandler) setup(ctx context.Context, input *SetupInput) (*AuthTokenOutput, error) {
	result, err := h.authService.Setup(ctx, service.SetupInput{
		Username: input.Body.Username,
		Password: input.Body.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrSetupAlreadyCompleted) {
			return nil, huma.Error409Conflict("setup already completed")
		}
		return nil, huma.Error500InternalServerError("failed to create admin account", err)
	}

	return buildAuthOutput(result), nil
}

// --- Login ---

type LoginInput struct {
	Body struct {
		Username string `json:"username" minLength:"1" doc:"Username"`
		Password string `json:"password" minLength:"1" doc:"Password"`
	}
}

func (h *AuthHandler) login(ctx context.Context, input *LoginInput) (*AuthTokenOutput, error) {
	result, err := h.authService.Login(ctx, service.LoginInput{
		Username: input.Body.Username,
		Password: input.Body.Password,
	})
	if err != nil {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}

	return buildAuthOutput(result), nil
}

// --- Refresh ---

type RefreshInput struct {
	RefreshToken string `cookie:"refresh_token" required:"true"`
}

func (h *AuthHandler) refresh(ctx context.Context, input *RefreshInput) (*AuthTokenOutput, error) {
	result, err := h.authService.Refresh(ctx, input.RefreshToken)
	if err != nil {
		return nil, huma.Error401Unauthorized("invalid or expired refresh token")
	}

	return buildAuthOutput(result), nil
}

// --- Logout ---

type LogoutInput struct {
	RefreshToken string `cookie:"refresh_token"`
}

func (h *AuthHandler) logout(ctx context.Context, input *LogoutInput) (*struct{}, error) {
	if input.RefreshToken != "" {
		h.authService.Logout(ctx, input.RefreshToken)
	}
	return nil, nil
}

// --- Helpers ---

func buildAuthOutput(result *service.AuthResult) *AuthTokenOutput {
	cookie := buildRefreshCookie(result.RefreshToken, 7*24*time.Hour)

	return &AuthTokenOutput{
		SetCookie: cookie,
		Body: AuthTokenBody{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			MemberID:     result.Member.ID,
			Role:         result.Member.Role,
			Username:     result.Member.Username,
		},
	}
}

func buildRefreshCookie(token string, maxAge time.Duration) string {
	return "refresh_token=" + token +
		"; HttpOnly; Secure; SameSite=Strict; Path=/api/auth" +
		"; Max-Age=" + formatSeconds(maxAge)
}

func formatSeconds(d time.Duration) string {
	s := int(d.Seconds())
	result := ""
	if s < 0 {
		result = "-"
		s = -s
	}
	digits := []byte{}
	if s == 0 {
		digits = append(digits, '0')
	}
	for s > 0 {
		digits = append(digits, byte('0'+s%10))
		s /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return result + string(digits)
}
