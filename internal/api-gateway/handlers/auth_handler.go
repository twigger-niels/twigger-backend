package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"twigger-backend/backend/auth-service/domain/service"
	"twigger-backend/backend/auth-service/infrastructure/persistence"
	"twigger-backend/internal/api-gateway/utils"

	"github.com/google/uuid"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
	userRepo    *persistence.PostgresUserRepository
	db          *sql.DB
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *sql.DB) *AuthHandler {
	// Initialize repositories
	userRepo := persistence.NewPostgresUserRepository(db)
	workspaceRepo := persistence.NewPostgresWorkspaceRepository(db)
	sessionRepo := persistence.NewPostgresSessionRepository(db)
	auditRepo := persistence.NewPostgresAuditRepository(db)

	// Initialize auth service
	authService := service.NewAuthService(userRepo, workspaceRepo, sessionRepo, auditRepo, db)

	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo.(*persistence.PostgresUserRepository),
		db:          db,
	}
}

// VerifyRequest represents the request body for POST /api/v1/auth/verify
type VerifyRequest struct {
	DeviceID *string `json:"device_id,omitempty"`
}

// LogoutRequest represents the request body for POST /api/v1/auth/logout
type LogoutRequest struct {
	DeviceID   *string `json:"device_id,omitempty"`
	RevokeAll  bool    `json:"revoke_all"`
}

// HandleVerify handles POST /api/v1/auth/verify
// This endpoint is called after Firebase has verified the JWT token
// The middleware has already extracted the user info from the token
func (h *AuthHandler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract Firebase UID from context (set by middleware)
	userIDStr := utils.GetUserID(ctx)
	if userIDStr == "" {
		utils.RespondUnauthorized(w, "Missing user ID")
		return
	}

	// Extract Firebase claims from context (set by middleware)
	claims, ok := ctx.Value("firebase_claims").(map[string]interface{})
	if !ok {
		utils.RespondUnauthorized(w, "Missing Firebase claims")
		return
	}

	// Use the user ID from context (which is the Firebase UID)
	firebaseUID := userIDStr

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		utils.RespondUnauthorized(w, "Invalid authentication token")
		return
	}

	// Extract optional fields
	provider := "email" // default
	if p, ok := claims["firebase"].(map[string]interface{}); ok {
		if signInProvider, ok := p["sign_in_provider"].(string); ok {
			provider = signInProvider
		}
	}

	emailVerified, _ := claims["email_verified"].(bool)

	var photoURL *string
	if picture, ok := claims["picture"].(string); ok && picture != "" {
		photoURL = &picture
	}

	// Parse request body for device_id
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		utils.RespondError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Extract client info
	ipAddress := getClientIP(r)
	userAgent := getUserAgent(r)

	// Complete authentication
	response, err := h.authService.CompleteAuthentication(
		ctx,
		firebaseUID,
		email,
		provider,
		emailVerified,
		photoURL,
		req.DeviceID,
		ipAddress,
		userAgent,
	)

	if err != nil {
		// Log detailed error server-side
		logError(r.Context(), "authentication failed", err, firebaseUID)
		// Return generic error to client
		utils.RespondError(w, errors.New("authentication failed"))
		return
	}

	// Return response
	utils.RespondJSON(w, http.StatusOK, response)
}

// HandleLogout handles POST /api/v1/auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get Firebase UID from context (set by auth middleware)
	firebaseUID := utils.GetUserID(ctx)
	if firebaseUID == "" {
		utils.RespondUnauthorized(w, "User not authenticated")
		return
	}

	// Get user UUID by Firebase UID
	userID, err := h.getUserIDByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		// Log detailed error server-side
		logError(ctx, "user lookup failed in logout", err, firebaseUID)
		// Return generic error to client
		utils.RespondUnauthorized(w, "User not authenticated")
		return
	}

	// Parse request body
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		utils.RespondError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Logout
	sessionsRevoked, err := h.authService.Logout(ctx, userID, req.DeviceID, req.RevokeAll)
	if err != nil {
		// Log detailed error server-side
		logError(ctx, "logout failed", err, firebaseUID)
		// Return generic error to client
		utils.RespondError(w, errors.New("logout failed"))
		return
	}

	// Return response
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message":          "Logged out successfully",
		"sessions_revoked": sessionsRevoked,
	})
}

// HandleMe handles GET /api/v1/auth/me
func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get Firebase UID from context (set by auth middleware)
	firebaseUID := utils.GetUserID(ctx)
	if firebaseUID == "" {
		utils.RespondUnauthorized(w, "User not authenticated")
		return
	}

	// Get user UUID by Firebase UID
	userID, err := h.getUserIDByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		// Log detailed error server-side
		logError(ctx, "user lookup failed in /me", err, firebaseUID)
		// Return generic error to client
		utils.RespondUnauthorized(w, "User not authenticated")
		return
	}

	// Get user
	user, err := h.authService.GetUser(ctx, userID)
	if err != nil {
		// Log detailed error server-side
		logError(ctx, "failed to get user details", err, firebaseUID)
		// Return generic error to client
		utils.RespondError(w, errors.New("failed to retrieve user information"))
		return
	}

	// Get user workspaces
	workspaces, err := h.authService.GetUserWorkspaces(ctx, userID)
	if err != nil {
		// Log detailed error server-side
		logError(ctx, "failed to get user workspaces", err, firebaseUID)
		// Return generic error to client
		utils.RespondError(w, errors.New("failed to retrieve workspaces"))
		return
	}

	// Return response
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"user":       user,
		"workspaces": workspaces,
	})
}

// Helper functions

func getClientIP(r *http.Request) *string {
	var ip string

	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ip = xff
	} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
		// Check X-Real-IP header
		ip = xri
	} else if r.RemoteAddr != "" {
		// Use RemoteAddr, but strip the port
		ip = stripPort(r.RemoteAddr)
	}

	if ip != "" {
		return &ip
	}

	return nil
}

// stripPort removes the port from an IP address string
// Handles both IPv4 (127.0.0.1:8080) and IPv6 ([::1]:8080) formats
func stripPort(addr string) string {
	// For IPv6 addresses like [::1]:8080
	if len(addr) > 0 && addr[0] == '[' {
		// Find the closing bracket
		if idx := strings.Index(addr, "]"); idx != -1 {
			return addr[1:idx] // Return just the IP without brackets
		}
	}

	// For IPv4 addresses like 127.0.0.1:8080
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}

	return addr
}

func getUserAgent(r *http.Request) *string {
	ua := r.Header.Get("User-Agent")
	if ua != "" {
		return &ua
	}
	return nil
}

// getUserIDByFirebaseUID looks up the user's UUID by their Firebase UID
func (h *AuthHandler) getUserIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error) {
	user, err := h.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return uuid.Nil, err
	}
	return user.UserID, nil
}

// logError logs detailed error information server-side (not exposed to client)
func logError(ctx context.Context, message string, err error, firebaseUID string) {
	// Use structured logging in production
	// For now, use standard log package
	log.Printf("ERROR: %s | Firebase UID: %s | Error: %v", message, firebaseUID, err)
}
