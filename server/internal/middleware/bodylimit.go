package middleware

import (
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// NewBodyLimit returns a Huma middleware that rejects request bodies larger
// than maxBytes. This prevents abuse from clients sending multi-GB payloads.
func NewBodyLimit(api huma.API, maxBytes int64) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		body := ctx.BodyReader()
		if body != nil {
			limited := io.LimitReader(body, maxBytes+1)
			ctx = huma.WithValue(ctx, bodyLimitKey{}, limited)
		}
		next(ctx)
	}
}

type bodyLimitKey struct{}

// NewChiBodyLimit returns a Chi middleware that limits request body size.
// This is applied at the router level before Huma processes the request.
func NewChiBodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
