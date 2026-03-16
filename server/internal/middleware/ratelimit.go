package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    float64 // max tokens
}

func newRateLimiter(rate float64, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    float64(burst),
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists {
		rl.visitors[key] = &visitor{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	elapsed := now.Sub(v.lastSeen).Seconds()
	v.lastSeen = now
	v.tokens += elapsed * rl.rate
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

func extractIP(ctx huma.Context) string {
	if forwarded := ctx.Header("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return ctx.Host()
}

// NewRateLimit returns a Huma middleware that rate-limits requests per IP.
// rate: requests per second, burst: max burst size.
func NewRateLimit(api huma.API, rate float64, burst int) func(huma.Context, func(huma.Context)) {
	rl := newRateLimiter(rate, burst)

	return func(ctx huma.Context, next func(huma.Context)) {
		ip := extractIP(ctx)
		if !rl.allow(ip) {
			huma.WriteErr(api, ctx, http.StatusTooManyRequests, "too many requests, try again later")
			return
		}
		next(ctx)
	}
}
