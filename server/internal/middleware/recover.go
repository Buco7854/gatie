package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/danielgtaylor/huma/v2"
)

func NewRecover(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "error", r, "stack", string(debug.Stack()))
				huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error")
			}
		}()
		next(ctx)
	}
}
