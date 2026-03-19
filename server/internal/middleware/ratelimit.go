package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valkey-io/valkey-go"
)

// TrustedProxies holds parsed CIDR networks used to identify trusted
// reverse proxies when extracting the real client IP from X-Forwarded-For.
type TrustedProxies struct {
	nets []*net.IPNet
}

// ParseTrustedProxies parses a comma-separated list of IPs and CIDR ranges.
// Single IPs are converted to /32 (IPv4) or /128 (IPv6).
// Returns an empty TrustedProxies (trust nothing) if input is empty.
func ParseTrustedProxies(raw string) (*TrustedProxies, error) {
	tp := &TrustedProxies{}
	if raw == "" {
		return tp, nil
	}

	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if !strings.Contains(entry, "/") {
			ip := net.ParseIP(entry)
			if ip == nil {
				return nil, fmt.Errorf("invalid trusted proxy IP: %q", entry)
			}
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			tp.nets = append(tp.nets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}

		_, cidr, err := net.ParseCIDR(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR: %w", err)
		}
		tp.nets = append(tp.nets, cidr)
	}

	return tp, nil
}

func (tp *TrustedProxies) isTrusted(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, n := range tp.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ExtractIP returns the real client IP from a request.
// It walks X-Forwarded-For from right to left, skipping IPs that belong
// to a trusted proxy network. The first non-trusted IP is the client.
// If all IPs are trusted or the header is absent, falls back to RemoteAddr.
func ExtractIP(ctx huma.Context, tp *TrustedProxies) string {
	xff := ctx.Header("X-Forwarded-For")
	if xff == "" || len(tp.nets) == 0 {
		return stripPort(ctx.RemoteAddr())
	}

	ips := strings.Split(xff, ",")
	for i := len(ips) - 1; i >= 0; i-- {
		ip := strings.TrimSpace(ips[i])
		if !tp.isTrusted(ip) {
			return ip
		}
	}

	return stripPort(ctx.RemoteAddr())
}

func stripPort(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// NewRateLimit returns a Huma middleware that rate-limits requests per IP
// using Valkey as backend. burst is the max number of requests allowed in
// the given window duration.
func NewRateLimit(api huma.API, vk valkey.Client, tp *TrustedProxies, burst int, window time.Duration) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		ip := ExtractIP(ctx, tp)
		key := fmt.Sprintf("ratelimit:%s:%s", ctx.URL().Path, ip)

		allowed, err := checkRateLimit(ctx.Context(), vk, key, burst, window)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusServiceUnavailable, "rate limiter unavailable")
			return
		}
		if !allowed {
			huma.WriteErr(api, ctx, http.StatusTooManyRequests, "too many requests, try again later")
			return
		}
		next(ctx)
	}
}

var rateLimitScript = valkey.NewLuaScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

// checkRateLimit implements an atomic fixed window counter using a Lua script.
func checkRateLimit(ctx context.Context, vk valkey.Client, key string, burst int, window time.Duration) (bool, error) {
	result := rateLimitScript.Exec(ctx, vk, []string{key}, []string{fmt.Sprintf("%d", int64(window.Seconds()))})

	count, err := result.AsInt64()
	if err != nil {
		return false, fmt.Errorf("rate limit script: %w", err)
	}

	return count <= int64(burst), nil
}
