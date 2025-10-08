// +build integration

package persistence

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"
	testhelpers "twigger-backend/backend/auth-service/infrastructure/database/testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTest(t *testing.T) (*PostgresUserRepository, context.Context, func()) {
	db := testhelpers.SetupTestDB(t)
	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	cleanup := func() {
		testhelpers.CleanupTestData(t, db)
		testhelpers.TeardownTestDB(t, db)
	}

	return repo.(*PostgresUserRepository), ctx, cleanup
}

func TestUserRepository_Integration_Create(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	t.Run("create new user with all fields", func(t *testing.T) {
		firebaseUID := "firebase-integration-test-001"
		email := "integration@example.com"
		username := "integrationuser"
		photoURL := "https://example.com/photo.jpg"

		user := &entity.User{
			UserID:        uuid.New(),
			FirebaseUID:   &firebaseUID,
			Email:         email,
			Username:      username,
			EmailVerified: true,
			PhotoURL:      &photoURL,
			Provider:      "google.com",
			CreatedAt:     time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err, "Should create user successfully")

		// Verify user was created
		retrieved, err := repo.GetByID(ctx, user.UserID)
		require.NoError(t, err)
		assert.Equal(t, email, retrieved.Email)
		assert.Equal(t, username, retrieved.Username)
		assert.Equal(t, firebaseUID, *retrieved.FirebaseUID)
		assert.True(t, retrieved.EmailVerified)
		assert.Equal(t, photoURL, *retrieved.PhotoURL)
		assert.Equal(t, "google.com", retrieved.Provider)
	})

	t.Run("create user with minimal fields", func(t *testing.T) {
		firebaseUID := "firebase-minimal-test"
		user := &entity.User{
			UserID:      uuid.New(),
			FirebaseUID: &firebaseUID,
			Email:       "minimal@example.com",
			Username:    "minimaluser",
			Provider:    "password",
			CreatedAt:   time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Verify nullable fields are nil
		retrieved, err := repo.GetByID(ctx, user.UserID)
		require.NoError(t, err)
		assert.Nil(t, retrieved.PhotoURL)
		assert.Nil(t, retrieved.PhoneNumber)
		assert.Nil(t, retrieved.LastLoginAt)
	})

	t.Run("duplicate firebase_uid rejected", func(t *testing.T) {
		firebaseUID := "firebase-duplicate-test"
		user1 := &entity.User{
			UserID:      uuid.New(),
			FirebaseUID: &firebaseUID,
			Email:       "user1@example.com",
			Username:    "user1",
			CreatedAt:   time.Now(),
		}

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		// Try to create another user with same firebase_uid
		user2 := &entity.User{
			UserID:      uuid.New(),
			FirebaseUID: &firebaseUID,
			Email:       "user2@example.com",
			Username:    "user2",
			CreatedAt:   time.Now(),
		}

		err = repo.Create(ctx, user2)
		assert.Error(t, err, "Should reject duplicate firebase_uid")
	})

	t.Run("duplicate email rejected", func(t *testing.T) {
		email := "duplicate@example.com"
		user1 := &entity.User{
			UserID:      uuid.New(),
			FirebaseUID: stringPtr("firebase-email-dup-1"),
			Email:       email,
			Username:    "emaildup1",
			CreatedAt:   time.Now(),
		}

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		// Try to create another user with same email
		user2 := &entity.User{
			UserID:      uuid.New(),
			FirebaseUID: stringPtr("firebase-email-dup-2"),
			Email:       email,
			Username:    "emaildup2",
			CreatedAt:   time.Now(),
		}

		err = repo.Create(ctx, user2)
		assert.Error(t, err, "Should reject duplicate email")
	})
}

func TestUserRepository_Integration_GetByFirebaseUID(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	firebaseUID := "firebase-getby-test"
	user := &entity.User{
		UserID:      uuid.New(),
		FirebaseUID: &firebaseUID,
		Email:       "getby@example.com",
		Username:    "getbyuser",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("retrieve existing user by firebase_uid", func(t *testing.T) {
		retrieved, err := repo.GetByFirebaseUID(ctx, firebaseUID)
		require.NoError(t, err)
		assert.Equal(t, user.UserID, retrieved.UserID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("non-existent firebase_uid returns error", func(t *testing.T) {
		_, err := repo.GetByFirebaseUID(ctx, "non-existent-uid")
		assert.Error(t, err, "Should return error for non-existent user")
	})
}

func TestUserRepository_Integration_GetByEmail(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	email := "email-lookup@example.com"
	user := &entity.User{
		UserID:      uuid.New(),
		FirebaseUID: stringPtr("firebase-email-lookup"),
		Email:       email,
		Username:    "emaillookup",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("retrieve user by email", func(t *testing.T) {
		retrieved, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, user.UserID, retrieved.UserID)
		assert.Equal(t, email, retrieved.Email)
	})

	t.Run("case-sensitive email lookup", func(t *testing.T) {
		// Note: Email lookup is case-sensitive by default in PostgreSQL
		_, err := repo.GetByEmail(ctx, "EMAIL-LOOKUP@EXAMPLE.COM")
		assert.Error(t, err, "Should not find user with different case email")
	})
}

func TestUserRepository_Integration_UpdateLastLogin(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	user := &entity.User{
		UserID:      uuid.New(),
		FirebaseUID: stringPtr("firebase-lastlogin-test"),
		Email:       "lastlogin@example.com",
		Username:    "lastloginuser",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify LastLoginAt is nil initially
	retrieved, err := repo.GetByID(ctx, user.UserID)
	require.NoError(t, err)
	assert.Nil(t, retrieved.LastLoginAt)

	// Update last login
	time.Sleep(100 * time.Millisecond) // Ensure timestamp difference
	err = repo.UpdateLastLogin(ctx, user.UserID)
	require.NoError(t, err)

	// Verify LastLoginAt is set
	updated, err := repo.GetByID(ctx, user.UserID)
	require.NoError(t, err)
	assert.NotNil(t, updated.LastLoginAt)
	assert.True(t, updated.LastLoginAt.After(user.CreatedAt))

	// Update again
	firstLogin := *updated.LastLoginAt
	time.Sleep(100 * time.Millisecond)
	err = repo.UpdateLastLogin(ctx, user.UserID)
	require.NoError(t, err)

	// Verify timestamp updated
	secondUpdate, err := repo.GetByID(ctx, user.UserID)
	require.NoError(t, err)
	assert.True(t, secondUpdate.LastLoginAt.After(firstLogin))
}

func TestUserRepository_Integration_Update(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	user := &entity.User{
		UserID:      uuid.New(),
		FirebaseUID: stringPtr("firebase-update-test"),
		Email:       "update@example.com",
		Username:    "updateuser",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("update user fields", func(t *testing.T) {
		// Modify user
		newPhotoURL := "https://example.com/new-photo.jpg"
		user.PhotoURL = &newPhotoURL
		user.EmailVerified = true

		err := repo.Update(ctx, user)
		require.NoError(t, err)

		// Verify updates
		retrieved, err := repo.GetByID(ctx, user.UserID)
		require.NoError(t, err)
		assert.Equal(t, newPhotoURL, *retrieved.PhotoURL)
		assert.True(t, retrieved.EmailVerified)
	})
}

func TestUserRepository_Integration_SoftDelete(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	user := &entity.User{
		UserID:      uuid.New(),
		FirebaseUID: stringPtr("firebase-delete-test"),
		Email:       "delete@example.com",
		Username:    "deleteuser",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("soft delete sets deleted_at", func(t *testing.T) {
		err := repo.SoftDelete(ctx, user.UserID)
		require.NoError(t, err)

		// Verify user is no longer retrievable (soft deleted users are filtered)
		_, err = repo.GetByID(ctx, user.UserID)
		assert.Error(t, err, "Should not retrieve soft-deleted user")
		assert.Equal(t, sql.ErrNoRows, err)
	})
}

func TestUserRepository_Integration_GetUserWorkspaces(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	// Create test user
	user := &entity.User{
		UserID:      uuid.New(),
		FirebaseUID: stringPtr("firebase-workspace-test"),
		Email:       "workspace@example.com",
		Username:    "workspaceuser",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create workspaces directly in database
	// Note: trigger automatically adds owner as admin member
	workspace1ID := uuid.New()
	workspace2ID := uuid.New()

	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, 'Workspace 1', NOW(), NOW())
	`, workspace1ID, user.UserID)
	require.NoError(t, err)

	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, 'Workspace 2', NOW(), NOW())
	`, workspace2ID, user.UserID)
	require.NoError(t, err)

	// Workspace members are automatically created by trigger, no manual insert needed

	t.Run("retrieve user workspaces", func(t *testing.T) {
		workspaces, err := repo.GetUserWorkspaces(ctx, user.UserID)
		require.NoError(t, err)
		assert.Len(t, workspaces, 2, "Should retrieve 2 workspaces")

		// Verify workspace details
		names := []string{workspaces[0].Name, workspaces[1].Name}
		assert.Contains(t, names, "Workspace 1")
		assert.Contains(t, names, "Workspace 2")
	})
}

func TestUserRepository_Integration_GetLinkedAccounts(t *testing.T) {
	repo, ctx, cleanup := setupUserTest(t)
	defer cleanup()

	user := &entity.User{
		UserID:      uuid.New(),
		FirebaseUID: stringPtr("firebase-linked-test"),
		Email:       "linked@example.com",
		Username:    "linkeduser",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("link provider accounts", func(t *testing.T) {
		// Link Google account
		err := repo.LinkProvider(ctx, user.UserID, "google.com", "google-provider-id-123")
		require.NoError(t, err)

		// Link Facebook account
		err = repo.LinkProvider(ctx, user.UserID, "facebook.com", "facebook-provider-id-456")
		require.NoError(t, err)

		// Get linked accounts
		accounts, err := repo.GetLinkedAccounts(ctx, user.UserID)
		require.NoError(t, err)
		assert.Len(t, accounts, 2, "Should have 2 linked accounts")

		// Verify providers
		providers := []string{accounts[0].Provider, accounts[1].Provider}
		assert.Contains(t, providers, "google.com")
		assert.Contains(t, providers, "facebook.com")
	})

	t.Run("duplicate provider link is idempotent", func(t *testing.T) {
		// LinkProvider uses ON CONFLICT DO NOTHING, so duplicates are silently ignored
		err := repo.LinkProvider(ctx, user.UserID, "google.com", "google-provider-id-123")
		assert.NoError(t, err, "Duplicate provider link should be idempotent")

		// Verify still only 2 accounts
		accounts, err := repo.GetLinkedAccounts(ctx, user.UserID)
		require.NoError(t, err)
		assert.Len(t, accounts, 2, "Should still have only 2 linked accounts")
	})
}
