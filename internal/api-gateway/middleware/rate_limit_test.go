package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	tests := []struct {
		name           string
		capacity       float64
		refillRate     float64
		requestCount   int
		sleepBetween   time.Duration
		expectedAllows int
	}{
		{
			name:           "allows_requests_within_capacity",
			capacity:       5,
			refillRate:     1.0, // 1 token per second
			requestCount:   5,
			sleepBetween:   0,
			expectedAllows: 5,
		},
		{
			name:           "blocks_requests_exceeding_capacity",
			capacity:       5,
			refillRate:     1.0,
			requestCount:   10,
			sleepBetween:   0,
			expectedAllows: 5,
		},
		{
			name:           "refills_tokens_over_time",
			capacity:       2,
			refillRate:     1.0, // 1 token per second
			requestCount:   4,
			sleepBetween:   1 * time.Second,
			expectedAllows: 4, // 2 initial + 2 refilled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := NewTokenBucket(tt.capacity, tt.refillRate)
			allowed := 0

			for i := 0; i < tt.requestCount; i++ {
				if i > 0 && tt.sleepBetween > 0 {
					time.Sleep(tt.sleepBetween)
				}
				if bucket.Allow() {
					allowed++
				}
			}

			if allowed != tt.expectedAllows {
				t.Errorf("Expected %d allowed requests, got %d", tt.expectedAllows, allowed)
			}
		})
	}
}

func TestRateLimiter_EndpointSpecificLimits(t *testing.T) {
	limiter := NewRateLimiter()

	tests := []struct {
		name         string
		endpoint     string
		expectedRate float64
	}{
		{
			name:         "verify_endpoint_strict_limit",
			endpoint:     "/api/v1/auth/verify",
			expectedRate: 5.0 / 60.0, // 5 per minute
		},
		{
			name:         "logout_endpoint_moderate_limit",
			endpoint:     "/api/v1/auth/logout",
			expectedRate: 10.0 / 60.0, // 10 per minute
		},
		{
			name:         "me_endpoint_higher_limit",
			endpoint:     "/api/v1/auth/me",
			expectedRate: 60.0 / 60.0, // 60 per minute = 1/sec
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, exists := limiter.endpointLimits[tt.endpoint]
			if !exists {
				t.Errorf("Endpoint %s not found in limits", tt.endpoint)
				return
			}

			if limit.RefillRate != tt.expectedRate {
				t.Errorf("Expected refill rate %f for %s, got %f",
					tt.expectedRate, tt.endpoint, limit.RefillRate)
			}
		})
	}
}

func TestRateLimitMiddleware_BlocksExcessRequests(t *testing.T) {
	limiter := NewRateLimiter()

	// Override with test limits for faster testing
	limiter.endpointLimits["/test"] = EndpointLimit{
		RequestsPerMinute: 3,
		Capacity:          3,
		RefillRate:        3.0 / 60.0,
	}

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make 5 requests, expect first 3 to succeed, last 2 to be rate limited
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if i < 3 {
			// First 3 requests should succeed
			if w.Code != http.StatusOK {
				t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
			}
		} else {
			// Last 2 requests should be rate limited
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Request %d: expected status 429, got %d", i+1, w.Code)
			}

			// Check for Retry-After header
			if w.Header().Get("Retry-After") == "" {
				t.Error("Expected Retry-After header in rate limited response")
			}

			// Check for rate limit headers
			if w.Header().Get("X-RateLimit-Limit") == "" {
				t.Error("Expected X-RateLimit-Limit header")
			}
		}
	}
}

func TestRateLimitMiddleware_DifferentClientsIndependent(t *testing.T) {
	limiter := NewRateLimiter()

	limiter.endpointLimits["/test"] = EndpointLimit{
		RequestsPerMinute: 2,
		Capacity:          2,
		RefillRate:        2.0 / 60.0,
	}

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Client 1 makes 2 requests (exhausts limit)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Client 1 request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Client 2 should still be able to make requests
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345" // Different IP
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Client 2: expected 200 (independent limit), got %d", w.Code)
	}
}

func TestRateLimitMiddleware_HeadersSet(t *testing.T) {
	limiter := NewRateLimiter()

	limiter.endpointLimits["/test"] = EndpointLimit{
		RequestsPerMinute: 10,
		Capacity:          10,
		RefillRate:        10.0 / 60.0,
	}

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check rate limit headers are present
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Expected X-RateLimit-Limit header")
	}

	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("Expected X-RateLimit-Remaining header")
	}

	t.Logf("Rate limit headers: Limit=%s, Remaining=%s",
		w.Header().Get("X-RateLimit-Limit"),
		w.Header().Get("X-RateLimit-Remaining"))
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		xri        string
		remoteAddr string
		expected   string
	}{
		{
			name:       "prefers_x_forwarded_for",
			xff:        "203.0.113.1",
			xri:        "198.51.100.1",
			remoteAddr: "192.0.2.1:12345",
			expected:   "203.0.113.1",
		},
		{
			name:       "uses_x_real_ip_if_no_xff",
			xff:        "",
			xri:        "198.51.100.1",
			remoteAddr: "192.0.2.1:12345",
			expected:   "198.51.100.1",
		},
		{
			name:       "uses_remote_addr_if_no_headers",
			xff:        "",
			xri:        "",
			remoteAddr: "192.0.2.1:12345",
			expected:   "192.0.2.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("Expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}
