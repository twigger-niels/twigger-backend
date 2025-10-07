package middleware

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"twigger-backend/internal/api-gateway/utils"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	tokens         float64
	capacity       float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
	mu             sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:         capacity,
		capacity:       capacity,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request can proceed and consumes a token if allowed
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+(elapsed*tb.refillRate))
	tb.lastRefillTime = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// RateLimiter manages rate limits for different endpoints and clients
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex

	// Default limits
	defaultCapacity   float64
	defaultRefillRate float64

	// Endpoint-specific limits
	endpointLimits map[string]EndpointLimit
}

// EndpointLimit defines rate limit for a specific endpoint
type EndpointLimit struct {
	RequestsPerMinute int
	Capacity          float64
	RefillRate        float64 // tokens per second
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets:           make(map[string]*TokenBucket),
		defaultCapacity:   100,                // 100 requests
		defaultRefillRate: 100.0 / 60.0,       // 100 per minute = 1.667/sec
		endpointLimits: map[string]EndpointLimit{
			"/api/v1/auth/verify": {
				RequestsPerMinute: 5,        // Very strict - 5 per minute
				Capacity:          5,
				RefillRate:        5.0 / 60.0, // 0.083 per second
			},
			"/api/v1/auth/logout": {
				RequestsPerMinute: 10,       // 10 per minute
				Capacity:          10,
				RefillRate:        10.0 / 60.0,
			},
			"/api/v1/auth/me": {
				RequestsPerMinute: 60,       // 60 per minute
				Capacity:          60,
				RefillRate:        60.0 / 60.0, // 1 per second
			},
		},
	}
}

// getBucket gets or creates a token bucket for a client
func (rl *RateLimiter) getBucket(key string, endpoint string) *TokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	// Create new bucket with endpoint-specific limits
	capacity := rl.defaultCapacity
	refillRate := rl.defaultRefillRate

	if limit, ok := rl.endpointLimits[endpoint]; ok {
		capacity = limit.Capacity
		refillRate = limit.RefillRate
	}

	rl.mu.Lock()
	// Double-check after acquiring write lock
	if bucket, exists := rl.buckets[key]; exists {
		rl.mu.Unlock()
		return bucket
	}

	bucket = NewTokenBucket(capacity, refillRate)
	rl.buckets[key] = bucket
	rl.mu.Unlock()

	return bucket
}

// cleanup removes old buckets periodically to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			// Remove buckets that are full (haven't been used recently)
			for key, bucket := range rl.buckets {
				bucket.mu.Lock()
				if bucket.tokens >= bucket.capacity {
					delete(rl.buckets, key)
				}
				bucket.mu.Unlock()
			}
			rl.mu.Unlock()
		}
	}()
}

// RateLimitMiddleware returns a middleware that enforces rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	// Start cleanup goroutine
	limiter.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client identifier (IP address)
			clientIP := getClientIP(r)
			endpoint := r.URL.Path

			// Create unique key: IP + endpoint
			key := fmt.Sprintf("%s:%s", clientIP, endpoint)

			// Get or create bucket for this client+endpoint
			bucket := limiter.getBucket(key, endpoint)

			// Check if request is allowed
			if !bucket.Allow() {
				// Log rate limit exceeded
				log.Printf("WARN: Rate limit exceeded for %s on %s", clientIP, endpoint)

				// Calculate retry-after based on refill rate
				limit := limiter.endpointLimits[endpoint]
				retryAfter := 60 / limit.RequestsPerMinute // seconds

				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")

				utils.RespondJSON(w, http.StatusTooManyRequests, map[string]interface{}{
					"error":   "too_many_requests",
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Rate limit exceeded. Please try again later.",
				})
				return
			}

			// Add rate limit headers
			if limit, ok := limiter.endpointLimits[endpoint]; ok {
				bucket.mu.Lock()
				remaining := int(bucket.tokens)
				bucket.mu.Unlock()

				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			}

			// Request allowed, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to get client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	return r.RemoteAddr
}

// min returns the smaller of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
