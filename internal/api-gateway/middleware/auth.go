package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"firebase.google.com/go/v4/auth"
	"twigger-backend/internal/api-gateway/firebase"
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
		// Skip auth for OPTIONS (preflight) requests - handled by CORS middleware
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

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
		firebaseToken, err := m.verifyFirebaseTokenFull(r.Context(), token)
		if err != nil {
			utils.RespondUnauthorized(w, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// Extract user ID from Firebase token (UID field)
		userID := firebaseToken.UID
		if userID == "" {
			log.Printf("DEBUG: Firebase token UID is empty. Token: %+v", firebaseToken)
			utils.RespondUnauthorized(w, "Invalid user ID in token")
			return
		}

		log.Printf("DEBUG: Authenticated user: %s", userID)

		// Store user ID and claims in context
		ctx := utils.SetUserID(r.Context(), userID)
		ctx = context.WithValue(ctx, "firebase_claims", firebaseToken.Claims)

		// Extract language preferences from user claims if available
		if lang, ok := firebaseToken.Claims["preferred_language"].(string); ok {
			var countryID *string
			if country, ok := firebaseToken.Claims["country"].(string); ok {
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

// verifyFirebaseTokenFull verifies a Firebase JWT token and returns the full token
func (m *AuthMiddleware) verifyFirebaseTokenFull(ctx context.Context, token string) (*auth.Token, error) {
	// Check if Firebase is initialized
	if !firebase.IsInitialized() {
		// Try to initialize Firebase
		if err := firebase.InitializeFirebase(ctx); err != nil {
			return nil, fmt.Errorf("Firebase not initialized: %w", err)
		}
	}

	// Verify the token using Firebase
	decodedToken, err := firebase.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return decodedToken, nil
}

// verifyFirebaseToken verifies a Firebase JWT token using Firebase Admin SDK (returns claims only)
func (m *AuthMiddleware) verifyFirebaseToken(ctx context.Context, token string) (map[string]interface{}, error) {
	// Check if Firebase is initialized
	if !firebase.IsInitialized() {
		// Try to initialize Firebase
		if err := firebase.InitializeFirebase(ctx); err != nil {
			// If initialization fails and we're in development mode, return mock
			if !m.enabled {
				return map[string]interface{}{
					"sub":            "dev-firebase-user",
					"email":          "dev@example.com",
					"email_verified": true,
				}, nil
			}
			return nil, fmt.Errorf("Firebase not initialized: %w", err)
		}
	}

	// Verify the token using Firebase
	decodedToken, err := firebase.VerifyIDToken(ctx, token)
	if err != nil {
		// Log the error for debugging
		log.Printf("Firebase token verification failed: %v", err)

		// If verification fails and we're in development mode with no credentials, return mock
		credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
		if credPath == "" && !m.enabled {
			return map[string]interface{}{
				"sub":            "dev-firebase-user",
				"email":          "dev@example.com",
				"email_verified": true,
			}, nil
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("Firebase token verified successfully for user: %v", decodedToken.UID)

	// Return the claims
	return decodedToken.Claims, nil
}
