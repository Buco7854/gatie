package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/gatie-io/gatie-server/internal/auth"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func NewAuthMiddleware(api huma.API, jwt *auth.JWTManager) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		header := ctx.Header("Authorization")
		if header == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing authorization header")
			return
		}

		token, found := strings.CutPrefix(header, "Bearer ")
		if !found {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		claims, err := jwt.ValidateAccessToken(token)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx = huma.WithValue(ctx, ClaimsKey, claims)
		next(ctx)
	}
}

func GetClaims(ctx huma.Context) *auth.Claims {
	claims, _ := ctx.Context().Value(ClaimsKey).(*auth.Claims)
	return claims
}

// GetClaimsFromContext retrieves claims from a standard context.Context.
// Used in handler functions which receive context.Context, not huma.Context.
func GetClaimsFromContext(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(ClaimsKey).(*auth.Claims)
	return claims
}
