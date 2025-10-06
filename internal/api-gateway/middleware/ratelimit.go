package middleware

import (
	"net/http"
	"sync"
	"time"

	"twigger-backend/internal/api-gateway/utils"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window (e.g., 1 minute)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup old visitors periodically
	go rl.cleanup()

	return rl
}

// Limit is middleware that enforces rate limiting
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as visitor identifier
		// In production, might want to use authenticated user ID
		ip := getIP(r)

		if !rl.allow(ip) {
			utils.RespondJSON(w, http.StatusTooManyRequests, utils.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Code:    "RATE_LIMIT_EXCEEDED",
				Message: "Too many requests. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks if a request from the visitor is allowed
func (rl *RateLimiter) allow(visitorID string) bool {
	rl.mu.Lock()
	v, exists := rl.visitors[visitorID]
	if !exists {
		v = &visitor{
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}
		rl.visitors[visitorID] = v
	}
	rl.mu.Unlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	// Refill tokens if window has passed
	now := time.Now()
	if now.Sub(v.lastRefill) >= rl.window {
		v.tokens = rl.rate
		v.lastRefill = now
	}

	// Check if tokens available
	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// cleanup removes old visitors periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for id, v := range rl.visitors {
			v.mu.Lock()
			if now.Sub(v.lastRefill) > rl.window*2 {
				delete(rl.visitors, id)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getIP extracts the IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (from proxy/load balancer)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
