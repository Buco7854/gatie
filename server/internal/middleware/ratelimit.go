package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valkey-io/valkey-go"
)

func extractIP(ctx huma.Context) string {
	if forwarded := ctx.Header("X-Forwarded-For"); forwarded != "" {
		if i := strings.IndexByte(forwarded, ','); i != -1 {
			return strings.TrimSpace(forwarded[:i])
		}
		return strings.TrimSpace(forwarded)
	}
	return ctx.RemoteAddr()
}

// NewRateLimit returns a Huma middleware that rate-limits requests per IP
// using Valkey as backend. burst is the max number of requests allowed in
// the given window duration.
func NewRateLimit(api huma.API, vk valkey.Client, burst int, window time.Duration) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		ip := extractIP(ctx)
		key := fmt.Sprintf("ratelimit:%s", ip)

		allowed, err := checkRateLimit(ctx.Context(), vk, key, burst, window)
		if err != nil {
			next(ctx)
			return
		}
		if !allowed {
			huma.WriteErr(api, ctx, http.StatusTooManyRequests, "too many requests, try again later")
			return
		}
		next(ctx)
	}
}

// checkRateLimit implements a sliding window counter using Valkey INCR + EXPIRE.
func checkRateLimit(ctx context.Context, vk valkey.Client, key string, burst int, window time.Duration) (bool, error) {
	cmds := make(valkey.Commands, 0, 3)
	cmds = append(cmds, vk.B().Incr().Key(key).Build())
	cmds = append(cmds, vk.B().Expire().Key(key).Seconds(int64(window.Seconds())).Nx().Build())

	results := vk.DoMulti(ctx, cmds...)

	count, err := results[0].AsInt64()
	if err != nil {
		return false, fmt.Errorf("rate limit INCR: %w", err)
	}

	return count <= int64(burst), nil
}
