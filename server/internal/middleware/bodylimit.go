package middleware

import (
	"net/http"
)

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
