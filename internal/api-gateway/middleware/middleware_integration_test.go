// +build integration

package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ===== AUTHENTICATION MIDDLEWARE TESTS =====

func TestAuthMiddleware_DevMode(t *testing.T) {
	auth := NewAuthMiddleware(false) // Dev mode

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(userIDKey)
		assert.NotNil(t, userID, "User ID should be set in dev mode")
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := auth.RequireAuth(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Should allow access in dev mode")
}

func TestAuthMiddleware_ProductionMode_MissingToken(t *testing.T) {
	auth := NewAuthMiddleware(true) // Production mode

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := auth.RequireAuth(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "Should return 401 without token")
	assert.Contains(t, rec.Body.String(), "Missing authorization token")
}

func TestAuthMiddleware_ProductionMode_WithToken(t *testing.T) {
	auth := NewAuthMiddleware(true) // Production mode

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := auth.RequireAuth(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req.Header.Set("Authorization", "Bearer mock-firebase-token")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Note: This will pass because auth middleware has mock implementation
	// Real Firebase verification will be added in Phase 3
	assert.Equal(t, http.StatusOK, rec.Code, "Should allow access with token")
}

// ===== CORS MIDDLEWARE TESTS =====

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	cors := NewCORSMiddleware([]string{"http://localhost:3000", "https://app.twigger.com"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := cors.Handle(handler)

	req := httptest.NewRequest("OPTIONS", "/api/v1/plants", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	cors := NewCORSMiddleware([]string{"http://localhost:3000"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := cors.Handle(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	cors := NewCORSMiddleware([]string{"http://localhost:3000"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := cors.Handle(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"), "Should not set CORS header for disallowed origin")
}

func TestCORSMiddleware_Wildcard(t *testing.T) {
	cors := NewCORSMiddleware([]string{"*"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := cors.Handle(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req.Header.Set("Origin", "http://any-domain.com")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://any-domain.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

// ===== RATE LIMIT MIDDLEWARE TESTS =====

func TestRateLimitMiddleware_AllowWithinLimit(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(60) // 60 req/min

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := rateLimit.Limit(handler)

	// Make 5 requests (well within limit)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/v1/plants", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimitMiddleware_BlockOverLimit(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(5) // Low limit for testing

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := rateLimit.Limit(handler)

	// Make 7 requests rapidly (exceeds limit of 5)
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest("GET", "/api/v1/plants", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		if rec.Code == http.StatusOK {
			successCount++
		} else if rec.Code == http.StatusTooManyRequests {
			rateLimitedCount++
		}
	}

	assert.Equal(t, 5, successCount, "Should allow 5 requests")
	assert.Equal(t, 2, rateLimitedCount, "Should rate limit 2 requests")
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(60)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := rateLimit.Limit(handler)

	// Request from IP 1
	req1 := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	rec1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec1, req1)

	// Request from IP 2
	req2 := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req2.RemoteAddr = "192.168.1.200:12345"
	rec2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec1.Code)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRateLimitMiddleware_TokenRefill(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(60) // 60 req/min = 1 req/sec

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := rateLimit.Limit(handler)

	// Make request
	req1 := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	rec1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Wait for token refill (1 second = 1 token)
	time.Sleep(1100 * time.Millisecond)

	// Make another request (should succeed due to refill)
	req2 := httptest.NewRequest("GET", "/api/v1/plants", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	rec2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRateLimitMiddleware_Concurrent(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(100)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := rateLimit.Limit(handler)

	// Make 50 concurrent requests
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/api/v1/plants", nil)
			req.RemoteAddr = "192.168.1.100:12345"
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Greater(t, successCount, 40, "Most requests should succeed")
}

// ===== LOGGING MIDDLEWARE TESTS =====

func TestLoggingMiddleware_LogsRequest(t *testing.T) {
	logging := NewLoggingMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	wrappedHandler := logging.Log(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Note: Actual logging output would be tested with a mock logger
}

func TestLoggingMiddleware_LogsError(t *testing.T) {
	logging := NewLoggingMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})

	wrappedHandler := logging.Log(handler)

	req := httptest.NewRequest("GET", "/api/v1/plants", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ===== MIDDLEWARE CHAIN TESTS =====

func TestMiddlewareChain_OrderMatters(t *testing.T) {
	// Test that middleware executes in correct order
	executionOrder := []string{}
	var mu sync.Mutex

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			executionOrder = append(executionOrder, "middleware1-before")
			mu.Unlock()
			next.ServeHTTP(w, r)
			mu.Lock()
			executionOrder = append(executionOrder, "middleware1-after")
			mu.Unlock()
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			executionOrder = append(executionOrder, "middleware2-before")
			mu.Unlock()
			next.ServeHTTP(w, r)
			mu.Lock()
			executionOrder = append(executionOrder, "middleware2-after")
			mu.Unlock()
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		executionOrder = append(executionOrder, "handler")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	// Chain: middleware1 -> middleware2 -> handler
	chain := middleware1(middleware2(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	chain.ServeHTTP(rec, req)

	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	assert.Equal(t, expected, executionOrder, "Middleware should execute in correct order")
}

func TestMiddlewareChain_EarlyExit(t *testing.T) {
	// Test that middleware can stop request propagation
	handlerCalled := false

	blockingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			// Don't call next.ServeHTTP() - stop here
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	chain := blockingMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	chain.ServeHTTP(rec, req)

	assert.False(t, handlerCalled, "Handler should not be called when middleware blocks")
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
