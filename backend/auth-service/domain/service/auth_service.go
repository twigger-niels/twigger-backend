package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"
	"twigger-backend/backend/auth-service/domain/repository"

	"github.com/google/uuid"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo      repository.UserRepository
	workspaceRepo repository.WorkspaceRepository
	sessionRepo   repository.SessionRepository
	auditRepo     repository.AuditRepository
	db            *sql.DB
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo repository.UserRepository,
	workspaceRepo repository.WorkspaceRepository,
	sessionRepo repository.SessionRepository,
	auditRepo repository.AuditRepository,
	db *sql.DB,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		sessionRepo:   sessionRepo,
		auditRepo:     auditRepo,
		db:            db,
	}
}

// AuthResponse represents the response from authentication
type AuthResponse struct {
	User       *entity.User        `json:"user"`
	Workspaces []*entity.Workspace `json:"workspaces"`
	SessionID  uuid.UUID           `json:"session_id"`
	IsNewUser  bool                `json:"is_new_user"`
}

// CompleteAuthentication handles post-Firebase authentication logic
// This is called after Firebase has verified the JWT token
func (s *AuthService) CompleteAuthentication(
	ctx context.Context,
	firebaseUID string,
	email string,
	provider string,
	emailVerified bool,
	photoURL *string,
	deviceID *string,
	ipAddress *string,
	userAgent *string,
) (*AuthResponse, error) {
	// Check if user exists by Firebase UID
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err == nil && user != nil {
		// Existing user found by Firebase UID - update and track provider link
		return s.handleExistingUserLogin(ctx, user, firebaseUID, provider, photoURL, deviceID, ipAddress, userAgent)
	}

	// User not found by Firebase UID - check if account exists with this email
	// This handles account linking: user signed in with different provider (e.g., Google then Facebook)
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// Account exists with this email but different Firebase UID - link the new provider
		return s.handleAccountLinking(ctx, existingUser, firebaseUID, provider, photoURL, deviceID, ipAddress, userAgent)
	}

	// No existing account - create new user + workspace
	user, workspaces, err := s.createNewUser(ctx, firebaseUID, email, provider, emailVerified, photoURL)
	if err != nil {
		// Log failed registration
		s.logAuditEvent(ctx, nil, entity.EventUserRegistered, false, ipAddress, userAgent, map[string]interface{}{
			"error":    err.Error(),
			"email":    email,
			"provider": provider,
		})
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	// Track initial provider link for new user
	if err := s.userRepo.LinkProvider(ctx, user.UserID, provider, firebaseUID); err != nil {
		// Non-blocking: log error but don't fail authentication
		s.logAuditEvent(ctx, &user.UserID, "provider_link_failed", false, ipAddress, userAgent, map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
	}

	// Create session
	session, err := s.createSession(ctx, user.UserID, deviceID, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log successful registration
	s.logAuditEvent(ctx, &user.UserID, entity.EventUserRegistered, true, ipAddress, userAgent, map[string]interface{}{
		"email":    email,
		"provider": provider,
	})

	return &AuthResponse{
		User:       user,
		Workspaces: workspaces,
		SessionID:  session.SessionID,
		IsNewUser:  true,
	}, nil
}

// handleExistingUserLogin processes login for users found by Firebase UID
func (s *AuthService) handleExistingUserLogin(
	ctx context.Context,
	user *entity.User,
	firebaseUID string,
	provider string,
	photoURL *string,
	deviceID *string,
	ipAddress *string,
	userAgent *string,
) (*AuthResponse, error) {
	// Update last login timestamp
	now := time.Now()
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}
	user.LastLoginAt = &now

	// Update photo_url if provided and user doesn't have one
	if photoURL != nil && *photoURL != "" && (user.PhotoURL == nil || *user.PhotoURL == "") {
		user.PhotoURL = photoURL
		if err := s.userRepo.Update(ctx, user); err != nil {
			// Non-blocking: log error but don't fail authentication
			s.logAuditEvent(ctx, &user.UserID, "photo_update_failed", false, ipAddress, userAgent, map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Track provider link (idempotent - ON CONFLICT DO NOTHING)
	if err := s.userRepo.LinkProvider(ctx, user.UserID, provider, firebaseUID); err != nil {
		// Non-blocking: log error but don't fail authentication
		s.logAuditEvent(ctx, &user.UserID, "provider_link_failed", false, ipAddress, userAgent, map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
	}

	// Get user workspaces
	workspaces, err := s.userRepo.GetUserWorkspaces(ctx, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}

	// Create session
	session, err := s.createSession(ctx, user.UserID, deviceID, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log audit event
	s.logAuditEvent(ctx, &user.UserID, entity.EventUserLogin, true, ipAddress, userAgent, map[string]interface{}{
		"provider": provider,
	})

	return &AuthResponse{
		User:       user,
		Workspaces: workspaces,
		SessionID:  session.SessionID,
		IsNewUser:  false,
	}, nil
}

// handleAccountLinking links a new provider to an existing account (same email, different provider)
func (s *AuthService) handleAccountLinking(
	ctx context.Context,
	existingUser *entity.User,
	newFirebaseUID string,
	newProvider string,
	photoURL *string,
	deviceID *string,
	ipAddress *string,
	userAgent *string,
) (*AuthResponse, error) {
	// SECURITY: Prevent account takeover by not automatically linking accounts
	// If the user already has a Firebase UID (existing account), they must sign in with that provider
	if existingUser.FirebaseUID != nil && *existingUser.FirebaseUID != newFirebaseUID {
		// Log failed linking attempt for security monitoring
		s.logAuditEvent(ctx, &existingUser.UserID, "account_linking_blocked", false, ipAddress, userAgent, map[string]interface{}{
			"email":              existingUser.Email,
			"existing_provider":  existingUser.Provider,
			"attempted_provider": newProvider,
			"reason":             "automatic linking disabled for security",
		})

		return nil, fmt.Errorf("this email is already registered with %s. Please sign in using %s",
			existingUser.Provider, existingUser.Provider)
	}

	// If firebase_uid matches, this is the same account - just update provider info
	existingUser.FirebaseUID = &newFirebaseUID
	existingUser.Provider = newProvider

	// Update photo_url if provided and user doesn't have one
	if photoURL != nil && *photoURL != "" && (existingUser.PhotoURL == nil || *existingUser.PhotoURL == "") {
		existingUser.PhotoURL = photoURL
	}

	// Update user in database
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user for account linking: %w", err)
	}

	// Track the new provider link
	if err := s.userRepo.LinkProvider(ctx, existingUser.UserID, newProvider, newFirebaseUID); err != nil {
		// Non-blocking: log error but don't fail authentication
		s.logAuditEvent(ctx, &existingUser.UserID, "provider_link_failed", false, ipAddress, userAgent, map[string]interface{}{
			"error":    err.Error(),
			"provider": newProvider,
		})
	}

	// Update last login
	now := time.Now()
	if err := s.userRepo.UpdateLastLogin(ctx, existingUser.UserID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}
	existingUser.LastLoginAt = &now

	// Get user workspaces
	workspaces, err := s.userRepo.GetUserWorkspaces(ctx, existingUser.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}

	// Create session
	session, err := s.createSession(ctx, existingUser.UserID, deviceID, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log account linking event
	s.logAuditEvent(ctx, &existingUser.UserID, "account_linked", true, ipAddress, userAgent, map[string]interface{}{
		"email":        existingUser.Email,
		"new_provider": newProvider,
		"previous_provider": existingUser.Provider,
	})

	return &AuthResponse{
		User:       existingUser,
		Workspaces: workspaces,
		SessionID:  session.SessionID,
		IsNewUser:  false,
	}, nil
}

// createNewUser creates a new user and default workspace in a transaction
func (s *AuthService) createNewUser(
	ctx context.Context,
	firebaseUID string,
	email string,
	provider string,
	emailVerified bool,
	photoURL *string,
) (*entity.User, []*entity.Workspace, error) {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Add panic recovery to ensure transaction rollback
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-panic after rollback
		}
	}()
	defer tx.Rollback() // Rollback if not committed

	// Create user entity
	user := &entity.User{
		UserID:        uuid.New(),
		FirebaseUID:   &firebaseUID,
		Email:         email,
		Username:      generateUsernameWithRetry(ctx, tx, email),
		EmailVerified: emailVerified,
		PhotoURL:      photoURL,
		Provider:      provider,
		CreatedAt:     time.Now(),
	}

	// Insert user directly using transaction
	userQuery := `
		INSERT INTO users (
			user_id, firebase_uid, email, username, email_verified,
			phone_number, photo_url, provider, created_at, last_login_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`
	_, err = tx.ExecContext(ctx, userQuery,
		user.UserID,
		user.FirebaseUID,
		user.Email,
		user.Username,
		user.EmailVerified,
		nil, // phone_number
		user.PhotoURL,
		user.Provider,
		user.CreatedAt,
		nil, // last_login_at
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default workspace
	workspace := &entity.Workspace{
		WorkspaceID: uuid.New(),
		OwnerID:     user.UserID,
		Name:        fmt.Sprintf("%s's Garden", user.Username),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	workspaceQuery := `
		INSERT INTO workspaces (workspace_id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(ctx, workspaceQuery,
		workspace.WorkspaceID,
		workspace.OwnerID,
		workspace.Name,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// The trigger in migration 008 automatically adds the owner as admin
	// But we'll add it explicitly for clarity and to ensure it happens
	memberQuery := `
		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, 'admin', $3)
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, memberQuery,
		workspace.WorkspaceID,
		user.UserID,
		time.Now(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add workspace member: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return user and workspace list
	workspaces := []*entity.Workspace{workspace}
	return user, workspaces, nil
}

// createSession creates a new authentication session
func (s *AuthService) createSession(
	ctx context.Context,
	userID uuid.UUID,
	deviceID *string,
	ipAddress *string,
	userAgent *string,
) (*entity.Session, error) {
	session := &entity.Session{
		SessionID: uuid.New(),
		UserID:    userID,
		DeviceID:  deviceID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// logAuditEvent logs an audit event (non-blocking, best effort)
func (s *AuthService) logAuditEvent(
	ctx context.Context,
	userID *uuid.UUID,
	eventType entity.AuditEventType,
	success bool,
	ipAddress *string,
	userAgent *string,
	metadata map[string]interface{},
) {
	event := &entity.AuditEvent{
		UserID:    userID,
		EventType: eventType,
		Success:   success,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	// Log event (ignore errors to prevent blocking auth flow)
	_ = s.auditRepo.LogEvent(ctx, event)
}

// Logout revokes a user's session
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, deviceID *string, revokeAll bool) (int, error) {
	if revokeAll {
		if err := s.sessionRepo.RevokeAllForUser(ctx, userID); err != nil {
			return 0, fmt.Errorf("failed to revoke all sessions: %w", err)
		}

		// Get count of revoked sessions
		sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("failed to get sessions: %w", err)
		}

		// Log logout event
		s.logAuditEvent(ctx, &userID, entity.EventUserLogout, true, nil, nil, map[string]interface{}{
			"revoke_all": true,
			"count":      len(sessions),
		})

		return len(sessions), nil
	}

	if deviceID != nil {
		if err := s.sessionRepo.RevokeByDeviceID(ctx, userID, *deviceID); err != nil {
			return 0, fmt.Errorf("failed to revoke session by device: %w", err)
		}

		// Log logout event
		s.logAuditEvent(ctx, &userID, entity.EventUserLogout, true, nil, nil, map[string]interface{}{
			"device_id": *deviceID,
		})

		return 1, nil
	}

	return 0, fmt.Errorf("either device_id or revoke_all must be specified")
}

// GetUser retrieves a user by ID
func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserWorkspaces retrieves all workspaces for a user
func (s *AuthService) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error) {
	return s.userRepo.GetUserWorkspaces(ctx, userID)
}

// generateUsername generates a username from an email (without uniqueness check)
func generateUsername(email string) string {
	// Take part before @ and sanitize
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return "user"
	}

	username := parts[0]
	// Replace special characters with underscore
	username = strings.ReplaceAll(username, ".", "_")
	username = strings.ReplaceAll(username, "+", "_")
	username = strings.ReplaceAll(username, "-", "_")

	return username
}

// generateUsernameWithRetry generates a unique username with database uniqueness check
func generateUsernameWithRetry(ctx context.Context, tx *sql.Tx, email string) string {
	baseUsername := generateUsername(email)

	// Try base username first
	if isUsernameAvailable(ctx, tx, baseUsername) {
		return baseUsername
	}

	// Retry with random suffix up to 5 times
	for i := 0; i < 5; i++ {
		suffix := uuid.New().String()[:8]
		candidateUsername := fmt.Sprintf("%s_%s", baseUsername, suffix)
		if isUsernameAvailable(ctx, tx, candidateUsername) {
			return candidateUsername
		}
	}

	// Fallback: use UUID if all retries fail
	return fmt.Sprintf("%s_%s", baseUsername, uuid.New().String()[:12])
}

// isUsernameAvailable checks if a username is available in the database
func isUsernameAvailable(ctx context.Context, tx *sql.Tx, username string) bool {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1 AND deleted_at IS NULL`
	err := tx.QueryRowContext(ctx, query, username).Scan(&count)
	if err != nil {
		// If error occurs, assume username is not available (safe default)
		return false
	}
	return count == 0
}

// isUsernameConflictError checks if an error is due to username unique constraint violation
func isUsernameConflictError(err error) bool {
	if err == nil {
		return false
	}
	// Check for PostgreSQL unique constraint error on username
	// Error code 23505 is unique_violation
	errMsg := err.Error()
	return strings.Contains(errMsg, "unique constraint") &&
		strings.Contains(errMsg, "username") ||
		strings.Contains(errMsg, "users_username_key") ||
		strings.Contains(errMsg, "duplicate key value")
}

// generateUsernameForAttempt generates a username for a specific attempt number
// attempt 0: base username (e.g., "john_doe")
// attempt 1+: username with UUID suffix (e.g., "john_doe_a3f9c2b1")
func generateUsernameForAttempt(email string, attempt int) string {
	baseUsername := generateUsername(email)

	if attempt == 0 {
		return baseUsername
	}

	// Generate UUID suffix for retries
	suffix := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s", baseUsername, suffix)
}

// RegisterWithUsername creates a new user with optional custom username
// If username is not provided, it will be auto-generated from email
func (s *AuthService) RegisterWithUsername(
	ctx context.Context,
	firebaseUID string,
	email string,
	provider string,
	emailVerified bool,
	photoURL *string,
	username *string,
	deviceID *string,
	ipAddress *string,
	userAgent *string,
) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err == nil && existingUser != nil {
		// User already exists, return existing user with session
		return s.handleExistingUserLogin(ctx, existingUser, firebaseUID, provider, photoURL, deviceID, ipAddress, userAgent)
	}

	// Check if email is already registered (account linking scenario)
	existingUser, err = s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// Account exists with this email but different Firebase UID
		return s.handleAccountLinking(ctx, existingUser, firebaseUID, provider, photoURL, deviceID, ipAddress, userAgent)
	}

	// Create new user with provided or auto-generated username
	user, workspaces, err := s.createNewUserWithUsername(ctx, firebaseUID, email, provider, emailVerified, photoURL, username)
	if err != nil {
		// Log failed registration
		s.logAuditEvent(ctx, nil, entity.EventUserRegistered, false, ipAddress, userAgent, map[string]interface{}{
			"error":    err.Error(),
			"email":    email,
			"provider": provider,
		})
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	// Track initial provider link for new user
	if err := s.userRepo.LinkProvider(ctx, user.UserID, provider, firebaseUID); err != nil {
		// Non-blocking: log error but don't fail authentication
		s.logAuditEvent(ctx, &user.UserID, "provider_link_failed", false, ipAddress, userAgent, map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
	}

	// Create session
	session, err := s.createSession(ctx, user.UserID, deviceID, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log successful registration
	s.logAuditEvent(ctx, &user.UserID, entity.EventUserRegistered, true, ipAddress, userAgent, map[string]interface{}{
		"email":    email,
		"provider": provider,
		"username": user.Username,
	})

	return &AuthResponse{
		User:       user,
		Workspaces: workspaces,
		SessionID:  session.SessionID,
		IsNewUser:  true,
	}, nil
}

// createNewUserWithUsername creates a new user with optional custom username
func (s *AuthService) createNewUserWithUsername(
	ctx context.Context,
	firebaseUID string,
	email string,
	provider string,
	emailVerified bool,
	photoURL *string,
	username *string,
) (*entity.User, []*entity.Workspace, error) {
	// Retry loop to handle race conditions in username generation
	maxRetries := 10
	for attempt := 0; attempt < maxRetries; attempt++ {
		user, workspaces, err := s.attemptCreateUserWithUsername(
			ctx, firebaseUID, email, provider, emailVerified, photoURL, username, attempt,
		)

		if err == nil {
			return user, workspaces, nil
		}

		// Check if error is due to username conflict
		if isUsernameConflictError(err) && username == nil {
			// Auto-generated username conflict - retry with new suffix
			continue
		}

		// Other errors (including custom username conflicts) should be returned immediately
		return nil, nil, err
	}

	return nil, nil, fmt.Errorf("failed to create user after %d attempts: username generation exhausted", maxRetries)
}

// attemptCreateUserWithUsername makes a single attempt to create a user
func (s *AuthService) attemptCreateUserWithUsername(
	ctx context.Context,
	firebaseUID string,
	email string,
	provider string,
	emailVerified bool,
	photoURL *string,
	username *string,
	attempt int,
) (*entity.User, []*entity.Workspace, error) {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Add panic recovery to ensure transaction rollback
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-panic after rollback
		}
	}()
	defer tx.Rollback() // Rollback if not committed

	// Determine username
	var finalUsername string
	if username != nil && *username != "" {
		// Custom username provided - check availability once
		if !isUsernameAvailable(ctx, tx, *username) {
			return nil, nil, fmt.Errorf("username already taken")
		}
		finalUsername = *username
	} else {
		// Auto-generate username from email with attempt-based suffix
		finalUsername = generateUsernameForAttempt(email, attempt)
	}

	// Create user entity
	user := &entity.User{
		UserID:        uuid.New(),
		FirebaseUID:   &firebaseUID,
		Email:         email,
		Username:      finalUsername,
		EmailVerified: emailVerified,
		PhotoURL:      photoURL,
		Provider:      provider,
		CreatedAt:     time.Now(),
	}

	// Insert user directly using transaction
	userQuery := `
		INSERT INTO users (
			user_id, firebase_uid, email, username, email_verified,
			phone_number, photo_url, provider, created_at, last_login_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`
	_, err = tx.ExecContext(ctx, userQuery,
		user.UserID,
		user.FirebaseUID,
		user.Email,
		user.Username,
		user.EmailVerified,
		nil, // phone_number
		user.PhotoURL,
		user.Provider,
		user.CreatedAt,
		nil, // last_login_at
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default workspace
	workspace := &entity.Workspace{
		WorkspaceID: uuid.New(),
		OwnerID:     user.UserID,
		Name:        fmt.Sprintf("%s's Garden", user.Username),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	workspaceQuery := `
		INSERT INTO workspaces (workspace_id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(ctx, workspaceQuery,
		workspace.WorkspaceID,
		workspace.OwnerID,
		workspace.Name,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Add user as admin to workspace
	memberQuery := `
		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, 'admin', $3)
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, memberQuery,
		workspace.WorkspaceID,
		user.UserID,
		time.Now(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add workspace member: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return user and workspace list
	workspaces := []*entity.Workspace{workspace}
	return user, workspaces, nil
}
