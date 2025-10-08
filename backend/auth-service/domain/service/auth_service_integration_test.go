// +build integration

package service

import (
	"context"
	"testing"
	"twigger-backend/backend/auth-service/domain/entity"
	"twigger-backend/backend/auth-service/infrastructure/persistence"
	testhelpers "twigger-backend/backend/auth-service/infrastructure/database/testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthServiceTest(t *testing.T) (*AuthService, context.Context, func()) {
	db := testhelpers.SetupTestDB(t)

	// Initialize repositories
	userRepo := persistence.NewPostgresUserRepository(db)
	workspaceRepo := persistence.NewPostgresWorkspaceRepository(db)
	sessionRepo := persistence.NewPostgresSessionRepository(db)
	auditRepo := persistence.NewPostgresAuditRepository(db)

	// Create service
	service := NewAuthService(userRepo, workspaceRepo, sessionRepo, auditRepo, db)
	ctx := context.Background()

	cleanup := func() {
		testhelpers.CleanupTestData(t, db)
		testhelpers.TeardownTestDB(t, db)
	}

	return service, ctx, cleanup
}

func TestAuthService_Integration_CompleteAuthentication_NewUser(t *testing.T) {
	service, ctx, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	t.Run("new user registration flow", func(t *testing.T) {
		firebaseUID := "firebase-new-user-integration"
		email := "newuser@example.com"
		provider := "google.com"
		photoURL := "https://example.com/photo.jpg"
		ipAddress := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		// Complete authentication (should create new user)
		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			provider,
			true,
			&photoURL,
			nil,
			&ipAddress,
			&userAgent,
		)

		require.NoError(t, err, "Should successfully register new user")
		assert.NotNil(t, response)
		assert.True(t, response.IsNewUser, "Should mark as new user")

		// Verify user details
		assert.NotNil(t, response.User)
		assert.Equal(t, email, response.User.Email)
		assert.Equal(t, firebaseUID, *response.User.FirebaseUID)
		assert.Equal(t, provider, response.User.Provider)
		assert.Equal(t, photoURL, *response.User.PhotoURL)
		assert.True(t, response.User.EmailVerified)

		// Verify username was generated
		assert.NotEmpty(t, response.User.Username)
		assert.Contains(t, response.User.Username, "newuser") // Should be based on email

		// Verify workspace was created
		assert.NotEmpty(t, response.Workspaces)
		assert.Len(t, response.Workspaces, 1, "Should create one default workspace")
		workspace := response.Workspaces[0]
		assert.Contains(t, workspace.Name, "Garden") // Default workspace name pattern
		assert.Equal(t, response.User.UserID, workspace.OwnerID)

		// Verify session was created
		assert.NotEqual(t, uuid.Nil, response.SessionID)
	})

	t.Run("new user creates workspace membership", func(t *testing.T) {
		firebaseUID := "firebase-workspace-member-test"
		email := "workspace@example.com"

		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			"google.com",
			true,
			nil,
			nil,
			nil,
			nil,
		)

		require.NoError(t, err)

		// Verify workspace membership exists
		// Query workspace_members table directly
		workspaceID := response.Workspaces[0].WorkspaceID
		userID := response.User.UserID

		var role string
		query := `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`
		err = service.db.QueryRowContext(ctx, query, workspaceID, userID).Scan(&role)
		require.NoError(t, err, "User should be a workspace member")
		assert.Equal(t, "admin", role, "User should be admin of their workspace")
	})
}

func TestAuthService_Integration_CompleteAuthentication_ExistingUser(t *testing.T) {
	service, ctx, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	// Create existing user
	firebaseUID := "firebase-existing-user"
	email := "existing@example.com"

	// First authentication (creates user)
	firstResponse, err := service.CompleteAuthentication(
		ctx,
		firebaseUID,
		email,
		"google.com",
		true,
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	require.True(t, firstResponse.IsNewUser)

	// Second authentication (existing user login)
	t.Run("existing user login flow", func(t *testing.T) {
		ipAddress := "10.0.0.1"
		userAgent := "Safari/17.0"

		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			"google.com",
			true,
			nil,
			nil,
			&ipAddress,
			&userAgent,
		)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.IsNewUser, "Should not mark as new user")

		// Verify same user
		assert.Equal(t, firstResponse.User.UserID, response.User.UserID)
		assert.Equal(t, email, response.User.Email)

		// Verify workspaces loaded
		assert.NotEmpty(t, response.Workspaces)
		assert.Equal(t, firstResponse.Workspaces[0].WorkspaceID, response.Workspaces[0].WorkspaceID)

		// Verify new session created
		assert.NotEqual(t, uuid.Nil, response.SessionID)
		assert.NotEqual(t, firstResponse.SessionID, response.SessionID, "Should create new session")

		// Verify last_login_at was updated
		assert.NotNil(t, response.User.LastLoginAt)
	})
}

func TestAuthService_Integration_AuditLogging(t *testing.T) {
	service, ctx, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	t.Run("registration audit event persisted", func(t *testing.T) {
		firebaseUID := "firebase-audit-registration"
		email := "audit@example.com"
		ipAddress := "192.168.1.100"
		userAgent := "Chrome/120.0"

		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			"apple.com",
			true,
			nil,
			nil,
			&ipAddress,
			&userAgent,
		)
		require.NoError(t, err)

		// Query audit log to verify event was persisted
		var eventCount int
		query := `
			SELECT COUNT(*)
			FROM auth_audit_log
			WHERE user_id = $1
			  AND event_type = $2
			  AND success = true
		`
		err = service.db.QueryRowContext(ctx, query, response.User.UserID, entity.EventUserRegistered).Scan(&eventCount)
		require.NoError(t, err)
		assert.Equal(t, 1, eventCount, "Should log registration event")

		// Verify event details
		var storedIP, storedUA string
		var metadata string
		detailQuery := `
			SELECT ip_address, user_agent, metadata::text
			FROM auth_audit_log
			WHERE user_id = $1 AND event_type = $2
		`
		err = service.db.QueryRowContext(ctx, detailQuery, response.User.UserID, entity.EventUserRegistered).
			Scan(&storedIP, &storedUA, &metadata)
		require.NoError(t, err)
		assert.Equal(t, ipAddress, storedIP)
		assert.Equal(t, userAgent, storedUA)
		assert.Contains(t, metadata, email)
		assert.Contains(t, metadata, "apple.com")
	})

	t.Run("login audit event persisted", func(t *testing.T) {
		firebaseUID := "firebase-audit-login"
		email := "login-audit@example.com"

		// Create user first
		_, err := service.CompleteAuthentication(ctx, firebaseUID, email, "google.com", true, nil, nil, nil, nil)
		require.NoError(t, err)

		// Login again
		ipAddress := "10.0.0.2"
		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			"google.com",
			true,
			nil,
			nil,
			&ipAddress,
			nil,
		)
		require.NoError(t, err)

		// Verify login event logged
		var loginCount int
		query := `
			SELECT COUNT(*)
			FROM auth_audit_log
			WHERE user_id = $1 AND event_type = $2
		`
		err = service.db.QueryRowContext(ctx, query, response.User.UserID, entity.EventUserLogin).Scan(&loginCount)
		require.NoError(t, err)
		assert.Equal(t, 1, loginCount, "Should log login event")
	})
}

func TestAuthService_Integration_Logout(t *testing.T) {
	service, ctx, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	// Create user and session
	firebaseUID := "firebase-logout-test"
	response, err := service.CompleteAuthentication(
		ctx,
		firebaseUID,
		"logout@example.com",
		"google.com",
		true,
		nil,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	userID := response.User.UserID

	t.Run("logout revokes session", func(t *testing.T) {
		deviceID := "device-123"

		// Create another session with device ID
		_, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			"logout@example.com",
			"google.com",
			true,
			nil,
			&deviceID,
			nil,
			nil,
		)
		require.NoError(t, err)

		// Logout with device ID
		count, err := service.Logout(ctx, userID, &deviceID, false)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should revoke 1 session")

		// Verify logout event logged
		var logoutCount int
		query := `
			SELECT COUNT(*)
			FROM auth_audit_log
			WHERE user_id = $1 AND event_type = $2
		`
		err = service.db.QueryRowContext(ctx, query, userID, entity.EventUserLogout).Scan(&logoutCount)
		require.NoError(t, err)
		assert.Equal(t, 1, logoutCount, "Should log logout event")
	})

	t.Run("revoke all sessions", func(t *testing.T) {
		// Create multiple sessions
		for i := 0; i < 3; i++ {
			_, err := service.CompleteAuthentication(
				ctx,
				firebaseUID,
				"logout@example.com",
				"google.com",
				true,
				nil,
				nil,
				nil,
				nil,
			)
			require.NoError(t, err)
		}

		// Revoke all
		count, err := service.Logout(ctx, userID, nil, true)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3, "Should revoke at least 3 sessions")
	})
}

func TestAuthService_Integration_TransactionIntegrity(t *testing.T) {
	service, ctx, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	t.Run("user creation is atomic", func(t *testing.T) {
		firebaseUID := "firebase-atomic-test"
		email := "atomic@example.com"

		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			"google.com",
			true,
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)

		// Verify all three entities exist (user, workspace, workspace_member)
		// 1. User exists
		var userExists bool
		err = service.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)",
			response.User.UserID).Scan(&userExists)
		require.NoError(t, err)
		assert.True(t, userExists)

		// 2. Workspace exists
		var workspaceExists bool
		err = service.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM workspaces WHERE workspace_id = $1)",
			response.Workspaces[0].WorkspaceID).Scan(&workspaceExists)
		require.NoError(t, err)
		assert.True(t, workspaceExists)

		// 3. Workspace member exists
		var memberExists bool
		err = service.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2)",
			response.Workspaces[0].WorkspaceID, response.User.UserID).Scan(&memberExists)
		require.NoError(t, err)
		assert.True(t, memberExists)
	})
}

func TestAuthService_Integration_EmailPasswordAuth(t *testing.T) {
	service, ctx, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	t.Run("email/password user with verified email", func(t *testing.T) {
		firebaseUID := "firebase-email-password-verified"
		email := "emailauth@example.com"

		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			"password", // Email/password provider
			true,       // Email verified
			nil,
			nil,
			nil,
			nil,
		)

		require.NoError(t, err, "Should allow email/password with verified email")
		assert.NotNil(t, response)
		assert.Equal(t, "password", response.User.Provider)
	})
}

func TestAuthService_Integration_UsernameGeneration(t *testing.T) {
	service, ctx, cleanup := setupAuthServiceTest(t)
	defer cleanup()

	t.Run("unique username generation", func(t *testing.T) {
		baseEmail := "testuser@example.com"

		// Create first user
		response1, err := service.CompleteAuthentication(
			ctx,
			"firebase-username-1",
			baseEmail,
			"google.com",
			true,
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		username1 := response1.User.Username

		// Try to create second user with similar email (should generate different username)
		response2, err := service.CompleteAuthentication(
			ctx,
			"firebase-username-2",
			"testuser+alias@example.com",
			"google.com",
			true,
			nil,
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
		username2 := response2.User.Username

		// Usernames should be different
		assert.NotEqual(t, username1, username2, "Usernames should be unique")
	})
}
