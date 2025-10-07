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
	// Check if user exists
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err == nil {
		// Existing user - update last login and log event
		if err := s.userRepo.UpdateLastLogin(ctx, user.UserID); err != nil {
			return nil, fmt.Errorf("failed to update last login: %w", err)
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
		s.logAuditEvent(ctx, &user.UserID, entity.EventUserLogin, true, ipAddress, userAgent, nil)

		return &AuthResponse{
			User:       user,
			Workspaces: workspaces,
			SessionID:  session.SessionID,
			IsNewUser:  false,
		}, nil
	}

	// New user - create account + workspace in transaction
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
	defer tx.Rollback()

	// Create user
	user := &entity.User{
		UserID:        uuid.New(),
		FirebaseUID:   &firebaseUID,
		Email:         email,
		Username:      generateUsername(email),
		EmailVerified: emailVerified,
		PhotoURL:      photoURL,
		Provider:      provider,
		CreatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
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

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Add user as workspace admin (this happens automatically via trigger, but we can be explicit)
	// Note: The trigger will handle this, but we can add it explicitly for clarity
	// The trigger in migration 008 automatically adds the owner as admin

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

// generateUsername generates a unique username from an email
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

	// Add random suffix to ensure uniqueness
	// In production, this should check for uniqueness in database
	suffix := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s", username, suffix)
}
