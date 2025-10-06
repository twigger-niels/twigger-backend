package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"twigger-backend/internal/api-gateway/utils"
)

// AuthMiddleware handles authentication using Firebase JWT tokens
type AuthMiddleware struct {
	projectID string
	enabled   bool
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(projectID string) *AuthMiddleware {
	// Check if auth is enabled (for development, can be disabled)
	enabled := os.Getenv("AUTH_ENABLED") != "false"

	return &AuthMiddleware{
		projectID: projectID,
		enabled:   enabled,
	}
}

// RequireAuth is middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if disabled (development mode)
		if !m.enabled {
			// Set a default user ID for development
			ctx := utils.SetUserID(r.Context(), "dev-user-123")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondUnauthorized(w, "Missing authorization header")
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondUnauthorized(w, "Invalid authorization header format")
			return
		}

		token := parts[1]

		// Verify Firebase JWT token
		claims, err := m.verifyFirebaseToken(r.Context(), token)
		if err != nil {
			utils.RespondUnauthorized(w, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// Extract user ID from claims
		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			utils.RespondUnauthorized(w, "Invalid user ID in token")
			return
		}

		// Store user ID in context
		ctx := utils.SetUserID(r.Context(), userID)

		// Extract language preferences from user claims if available
		if lang, ok := claims["preferred_language"].(string); ok {
			var countryID *string
			if country, ok := claims["country"].(string); ok {
				countryID = &country
			}

			langCtx := &utils.LanguageContext{
				LanguageID: lang,
				CountryID:  countryID,
			}
			ctx = utils.SetLanguageContext(ctx, langCtx)
		}

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is middleware that optionally handles authentication
// If a valid token is provided, user context is set. Otherwise, request proceeds without auth.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if disabled
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Try to extract and verify token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			next.ServeHTTP(w, r)
			return
		}

		token := parts[1]
		claims, err := m.verifyFirebaseToken(r.Context(), token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Set user ID if valid
		if userID, ok := claims["sub"].(string); ok && userID != "" {
			ctx := utils.SetUserID(r.Context(), userID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// verifyFirebaseToken verifies a Firebase JWT token
// TODO: Implement actual Firebase Admin SDK verification
func (m *AuthMiddleware) verifyFirebaseToken(ctx context.Context, token string) (map[string]interface{}, error) {
	// For now, return a mock implementation
	// In production, this should use Firebase Admin SDK:
	//
	// import firebase "firebase.google.com/go/v4"
	// import "firebase.google.com/go/v4/auth"
	//
	// app, err := firebase.NewApp(ctx, nil)
	// client, err := app.Auth(ctx)
	// token, err := client.VerifyIDToken(ctx, idToken)
	// return token.Claims, nil

	return map[string]interface{}{
		"sub": "mock-user-id",
	}, nil
}
