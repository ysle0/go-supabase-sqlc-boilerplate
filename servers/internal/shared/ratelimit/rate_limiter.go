package ratelimit

import (
	"net/http"
	"sync"

	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	readLimiter  *rate.Limiter
	writeLimiter *rate.Limiter
	mu           sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with configurable limits
func NewRateLimiter() *RateLimiter {
	readLimit := rate.Limit(shared.EnvFloat64("RATE_LIMIT_READ", 100.0))  // requests per second for read operations
	writeLimit := rate.Limit(shared.EnvFloat64("RATE_LIMIT_WRITE", 10.0)) // requests per second for write operations

	return &RateLimiter{
		readLimiter:  rate.NewLimiter(readLimit, int(readLimit)*2),   // burst size = limit * 2
		writeLimiter: rate.NewLimiter(writeLimit, int(writeLimit)*2), // burst size = limit * 2
	}
}

// LimitByRequest applies rate limiting based on HTTP method
func (rl *RateLimiter) LimitByRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var limiter *rate.Limiter

		rl.mu.RLock()
		// Choose limiter based on HTTP method
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			limiter = rl.readLimiter
		default: // POST, PUT, DELETE, etc.
			limiter = rl.writeLimiter
		}
		rl.mu.RUnlock()

		if !limiter.Allow() {
			// Rate limit exceeded
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too many requests", "retry_after": 60}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetReadLimiter returns the read operations limiter
func (rl *RateLimiter) GetReadLimiter() *rate.Limiter {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.readLimiter
}

// GetWriteLimiter returns the write operations limiter
func (rl *RateLimiter) GetWriteLimiter() *rate.Limiter {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.writeLimiter
}

// NewWebSocketLimiter creates a limiter for WebSocket connections
func NewWebSocketLimiter() *rate.Limiter {
	wsLimit := rate.Limit(shared.EnvFloat64("WS_RATE_LIMIT", 50.0)) // messages per second per connection
	return rate.NewLimiter(wsLimit, int(wsLimit)*2)
}
