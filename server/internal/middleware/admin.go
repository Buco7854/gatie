package middleware

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/gatie-io/gatie-server/internal/service"
)

func NewRequireAdmin(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims := GetClaims(ctx)
		if claims == nil {
			slog.Warn("admin middleware: no claims in context, auth middleware may not have run")
			huma.WriteErr(api, ctx, http.StatusForbidden, "admin access required")
			return
		}
		if claims.Role != service.RoleAdmin {
			huma.WriteErr(api, ctx, http.StatusForbidden, "admin access required")
			return
		}
		next(ctx)
	}
}
