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

// ExtractIP returns the client IP from a request, taking into account
// the number of trusted reverse proxies in front of the server.
//
// With trustedProxies=1 (e.g. Caddy), the rightmost IP in X-Forwarded-For
// is the one added by the trusted proxy = the real client IP.
// With trustedProxies=0 (direct access), X-Forwarded-For is ignored entirely.
func ExtractIP(ctx huma.Context, trustedProxies int) string {
	if trustedProxies <= 0 {
		return stripPort(ctx.RemoteAddr())
	}

	xff := ctx.Header("X-Forwarded-For")
	if xff == "" {
		return stripPort(ctx.RemoteAddr())
	}

	ips := strings.Split(xff, ",")
	// The trusted proxy at position N (from the right) added the client IP.
	// With 1 trusted proxy: we want ips[len-1] (the IP Caddy appended).
	idx := len(ips) - trustedProxies
	if idx < 0 {
		return stripPort(ctx.RemoteAddr())
	}

	return strings.TrimSpace(ips[idx])
}

func stripPort(addr string) string {
	if i := strings.LastIndexByte(addr, ':'); i != -1 {
		// Avoid stripping IPv6 colons — only strip if there's a bracket or no other colon.
		if strings.Contains(addr, "]") {
			// [::1]:8080 → [::1]
			if bracket := strings.LastIndexByte(addr, ']'); bracket < i {
				return addr[:i]
			}
			return addr
		}
		if strings.IndexByte(addr, ':') == i {
			// Only one colon → IPv4:port
			return addr[:i]
		}
	}
	return addr
}

// NewRateLimit returns a Huma middleware that rate-limits requests per IP
// using Valkey as backend. burst is the max number of requests allowed in
// the given window duration.
func NewRateLimit(api huma.API, vk valkey.Client, trustedProxies int, burst int, window time.Duration) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		ip := ExtractIP(ctx, trustedProxies)
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

// checkRateLimit implements a fixed window counter using Valkey INCR + EXPIRE.
func checkRateLimit(ctx context.Context, vk valkey.Client, key string, burst int, window time.Duration) (bool, error) {
	cmds := make(valkey.Commands, 0, 2)
	cmds = append(cmds, vk.B().Incr().Key(key).Build())
	cmds = append(cmds, vk.B().Expire().Key(key).Seconds(int64(window.Seconds())).Nx().Build())

	results := vk.DoMulti(ctx, cmds...)

	count, err := results[0].AsInt64()
	if err != nil {
		return false, fmt.Errorf("rate limit INCR: %w", err)
	}

	return count <= int64(burst), nil
}
