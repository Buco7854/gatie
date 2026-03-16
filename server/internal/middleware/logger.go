package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type handlerErrorKey struct{}

type handlerErrorState struct{ err error }

// ErrorCaptureTransformer captures errors returned by handlers into the request
// context so that the request logger middleware can log them after next() returns.
// Must be registered in huma.Config.Transformers before creating the API instance.
func ErrorCaptureTransformer(ctx huma.Context, status string, v any) (any, error) {
	if err, ok := v.(error); ok {
		if state, ok := ctx.Context().Value(handlerErrorKey{}).(*handlerErrorState); ok {
			state.err = err
		}
	}
	return v, nil
}

// NewRequestLogger returns a middleware that logs every request with method,
// path, status, duration, and the handler error if any.
func NewRequestLogger() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		start := time.Now()
		state := &handlerErrorState{}
		ctx = huma.WithValue(ctx, handlerErrorKey{}, state)

		next(ctx)

		status := ctx.Status()
		if status == 0 {
			status = http.StatusOK
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		attrs := []slog.Attr{
			slog.String("method", ctx.Method()),
			slog.String("path", ctx.URL().Path),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
		}

		if state.err != nil {
			attrs = append(attrs, slog.String("error", state.err.Error()))
			if em, ok := state.err.(*huma.ErrorModel); ok && len(em.Errors) > 0 {
				msgs := make([]string, 0, len(em.Errors))
				for _, d := range em.Errors {
					if d != nil && d.Message != "" {
						msgs = append(msgs, d.Message)
					}
				}
				if len(msgs) > 0 {
					attrs = append(attrs, slog.String("cause", strings.Join(msgs, "; ")))
				}
			}
		}

		slog.LogAttrs(ctx.Context(), level, "request", attrs...)
	}
}
