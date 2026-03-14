package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/gatie-io/gatie-server/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// --- Setup ---

type SetupInput struct {
	Body struct {
		Username string `json:"username" minLength:"3" maxLength:"100" doc:"Admin username"`
		Password string `json:"password" minLength:"8" maxLength:"128" doc:"Admin password"`
	}
}

type AuthTokenBody struct {
	AccessToken  string     `json:"access_token" doc:"JWT access token"`
	RefreshToken string     `json:"refresh_token" doc:"Opaque refresh token"`
	MemberID     string     `json:"member_id" doc:"Member UUID"`
	Role         string     `json:"role" doc:"Member role"`
	Username     string     `json:"username" doc:"Member username"`
}

type AuthTokenOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      AuthTokenBody
}

func (h *AuthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "setup",
		Method:        http.MethodPost,
		Path:          "/setup",
		Summary:       "Initial setup",
		Description:   "Create the first admin account. Only available when no members exist.",
		Tags:          []string{"Auth"},
		DefaultStatus: http.StatusCreated,
	}, h.setup)

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login",
		Description: "Authenticate with username and password.",
		Tags:        []string{"Auth"},
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "refresh",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Refresh tokens",
		Description: "Get a new access token using a refresh token.",
		Tags:        []string{"Auth"},
	}, h.refresh)

	huma.Register(api, huma.Operation{
		OperationID:   "logout",
		Method:        http.MethodPost,
		Path:          "/auth/logout",
		Summary:       "Logout",
		Description:   "Revoke the refresh token.",
		Tags:          []string{"Auth"},
		DefaultStatus: http.StatusNoContent,
	}, h.logout)
}

func (h *AuthHandler) setup(ctx context.Context, input *SetupInput) (*AuthTokenOutput, error) {
	result, err := h.authService.Setup(ctx, service.SetupInput{
		Username: input.Body.Username,
		Password: input.Body.Password,
	})
	if err != nil {
		if err.Error() == "setup already completed" {
			return nil, huma.Error409Conflict("setup already completed")
		}
		return nil, huma.Error500InternalServerError("failed to create admin account")
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
	Body struct {
		RefreshToken string `json:"refresh_token" minLength:"1" doc:"Refresh token"`
	}
}

func (h *AuthHandler) refresh(ctx context.Context, input *RefreshInput) (*AuthTokenOutput, error) {
	result, err := h.authService.Refresh(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, huma.Error401Unauthorized("invalid or expired refresh token")
	}

	return buildAuthOutput(result), nil
}

// --- Logout ---

type LogoutInput struct {
	Body struct {
		RefreshToken string `json:"refresh_token" minLength:"1" doc:"Refresh token to revoke"`
	}
}

func (h *AuthHandler) logout(ctx context.Context, input *LogoutInput) (*struct{}, error) {
	h.authService.Logout(ctx, input.Body.RefreshToken)
	return nil, nil
}

// --- Helpers ---

func buildAuthOutput(result *service.AuthResult) *AuthTokenOutput {
	memberID := uuidBytesToString(result.Member.ID.Bytes)

	cookie := buildRefreshCookie(result.RefreshToken, 7*24*time.Hour)

	return &AuthTokenOutput{
		SetCookie: cookie,
		Body: AuthTokenBody{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			MemberID:     memberID,
			Role:         result.Member.Role,
			Username:     result.Member.Username,
		},
	}
}

func buildRefreshCookie(token string, maxAge time.Duration) string {
	return "refresh_token=" + token +
		"; HttpOnly; Secure; SameSite=Strict; Path=/auth" +
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

func uuidBytesToString(u [16]byte) string {
	return formatHex(u[0:4]) + "-" + formatHex(u[4:6]) + "-" + formatHex(u[6:8]) + "-" + formatHex(u[8:10]) + "-" + formatHex(u[10:16])
}

func formatHex(b []byte) string {
	const hexDigits = "0123456789abcdef"
	result := make([]byte, len(b)*2)
	for i, v := range b {
		result[i*2] = hexDigits[v>>4]
		result[i*2+1] = hexDigits[v&0x0f]
	}
	return string(result)
}
