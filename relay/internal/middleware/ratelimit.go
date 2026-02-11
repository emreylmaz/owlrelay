package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/emreylmaz/owlrelay/relay/internal/store"
)

// RateLimiter implements in-memory rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	limits   map[string]*tokenLimit
	cleanup  time.Duration
}

type tokenLimit struct {
	count    int
	resetAt  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits:  make(map[string]*tokenLimit),
		cleanup: time.Minute * 5,
	}
	go rl.cleanupLoop()
	return rl
}

// RateLimit creates a rate limiting middleware
func (rl *RateLimiter) RateLimit(tokenStore *store.TokenStore) func(http.Handler) http.Handler {
	_ = tokenStore // Reserved for future use
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := TokenFromContext(r.Context())
			if token == nil {
				// No token, skip rate limiting (auth middleware should handle this)
				next.ServeHTTP(w, r)
				return
			}

			// Use token ID as rate limit key
			key := strconv.FormatInt(token.ID, 10)

			limit := token.RateLimit
			if limit <= 0 {
				limit = 100 // Default
			}

			if !rl.allow(key, limit) {
				retryAfter := rl.getRetryAfter(key)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":{"code":"RATE_LIMITED","message":"Too many requests","retryAfter":` + strconv.Itoa(retryAfter) + `}}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allow(key string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowDuration := time.Minute

	tl, exists := rl.limits[key]
	if !exists || tl.resetAt.Before(now) {
		rl.limits[key] = &tokenLimit{
			count:   1,
			resetAt: now.Add(windowDuration),
		}
		return true
	}

	if tl.count >= limit {
		return false
	}

	tl.count++
	return true
}

func (rl *RateLimiter) getRetryAfter(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if tl, exists := rl.limits[key]; exists {
		remaining := time.Until(tl.resetAt)
		if remaining > 0 {
			return int(remaining.Seconds()) + 1
		}
	}
	return 1
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, tl := range rl.limits {
			if tl.resetAt.Before(now) {
				delete(rl.limits, key)
			}
		}
		rl.mu.Unlock()
	}
}
