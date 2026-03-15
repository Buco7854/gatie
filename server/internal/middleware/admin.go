package middleware

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RequireAdmin(ctx huma.Context, next func(huma.Context)) {
	claims := GetClaims(ctx)
	if claims == nil || claims.Role != "ADMIN" {
		writeError(ctx, http.StatusForbidden, "admin access required")
		return
	}
	next(ctx)
}
