package service

import (
	"context"
	"testing"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccountLinking_SameEmailDifferentProviders tests that users can link multiple providers to one account
func TestAccountLinking_SameEmailDifferentProviders(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockWorkspaceRepo := NewMockWorkspaceRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockAuditRepo := NewMockAuditRepository()

	service := NewAuthService(mockUserRepo, mockWorkspaceRepo, mockSessionRepo, mockAuditRepo, nil)

	ctx := context.Background()
	email := "test@example.com"
	ipAddress := stringPtr("192.168.1.1")
	userAgent := stringPtr("TestAgent")

	t.Run("setup: create initial Google user", func(t *testing.T) {
		// Manually create a user as if they registered with Google
		// (bypassing transaction-based createNewUser which needs real DB)
		googleUID := "google-uid-123"
		googleProvider := "google.com"
		photoURL := stringPtr("https://google.com/photo.jpg")

		existingUser := &entity.User{
			UserID:        uuid.New(),
			FirebaseUID:   &googleUID,
			Email:         email,
			Username:      "testuser",
			Provider:      googleProvider,
			EmailVerified: true,
			PhotoURL:      photoURL,
			CreatedAt:     time.Now(),
		}

		// Add to mock repository
		err := mockUserRepo.Create(ctx, existingUser)
		require.NoError(t, err)

		// Track initial provider link
		err = mockUserRepo.LinkProvider(ctx, existingUser.UserID, googleProvider, googleUID)
		require.NoError(t, err)

		// Verify setup
		users := mockUserRepo.GetAllUsers()
		assert.Len(t, users, 1)
		assert.Equal(t, googleUID, *users[0].FirebaseUID)

		links := mockUserRepo.GetAllLinkedAccounts()
		assert.Len(t, links, 1)
		assert.Equal(t, googleProvider, links[0].Provider)
	})

	t.Run("second login with Facebook links to existing account", func(t *testing.T) {
		facebookUID := "facebook-uid-456"
		facebookProvider := "facebook.com"

		response, err := service.CompleteAuthentication(
			ctx,
			facebookUID,
			email, // Same email as Google account
			facebookProvider,
			true,
			nil,
			nil,
			ipAddress,
			userAgent,
		)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.IsNewUser, "Should link to existing account, not create new user")
		assert.Equal(t, email, response.User.Email)
		assert.Equal(t, facebookProvider, response.User.Provider, "Provider should be updated to most recent")

		// Verify no duplicate user was created
		users := mockUserRepo.GetAllUsers()
		assert.Len(t, users, 1, "Should still have only one user")

		// Verify both providers are linked
		links := mockUserRepo.GetAllLinkedAccounts()
		assert.Len(t, links, 2, "Should have both Google and Facebook linked")

		providers := make(map[string]bool)
		for _, link := range links {
			providers[link.Provider] = true
		}
		assert.True(t, providers["google.com"], "Google provider should be linked")
		assert.True(t, providers["facebook.com"], "Facebook provider should be linked")

		// Verify account_linked audit event was logged
		auditEvents := mockAuditRepo.GetAllEvents()
		var foundLinkingEvent bool
		for _, event := range auditEvents {
			if event.EventType == "account_linked" {
				foundLinkingEvent = true
				assert.True(t, event.Success)
				assert.Equal(t, email, event.Metadata["email"])
				assert.Equal(t, facebookProvider, event.Metadata["new_provider"])
				break
			}
		}
		assert.True(t, foundLinkingEvent, "Should have logged account_linked event")
	})

	t.Run("subsequent login with original provider still works", func(t *testing.T) {
		googleUID := "google-uid-123"
		googleProvider := "google.com"

		response, err := service.CompleteAuthentication(
			ctx,
			googleUID,
			email,
			googleProvider,
			true,
			nil,
			nil,
			ipAddress,
			userAgent,
		)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.IsNewUser)
		assert.Equal(t, email, response.User.Email)

		// Should still have only one user
		users := mockUserRepo.GetAllUsers()
		assert.Len(t, users, 1)

		// Should still have both providers linked
		links := mockUserRepo.GetAllLinkedAccounts()
		assert.Len(t, links, 2)
	})
}

// TestPhotoURLUpdate tests that photo URLs are updated when provided
func TestPhotoURLUpdate(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockWorkspaceRepo := NewMockWorkspaceRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockAuditRepo := NewMockAuditRepository()

	service := NewAuthService(mockUserRepo, mockWorkspaceRepo, mockSessionRepo, mockAuditRepo, nil)

	ctx := context.Background()
	email := "photo-test@example.com"
	firebaseUID := "firebase-uid-789"
	provider := "google.com"

	t.Run("existing user without photo gets updated", func(t *testing.T) {
		// Setup: create user without photo
		user := &entity.User{
			UserID:        uuid.New(),
			FirebaseUID:   &firebaseUID,
			Email:         email,
			Username:      "photouser",
			Provider:      provider,
			EmailVerified: true,
			PhotoURL:      nil, // No photo initially
			CreatedAt:     time.Now(),
		}
		err := mockUserRepo.Create(ctx, user)
		require.NoError(t, err)

		newPhotoURL := stringPtr("https://google.com/new-photo.jpg")

		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			provider,
			true,
			newPhotoURL,
			nil,
			nil,
			nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, response.User.PhotoURL)
		assert.Equal(t, "https://google.com/new-photo.jpg", *response.User.PhotoURL)
	})

	t.Run("existing user with photo keeps original", func(t *testing.T) {
		// User already has photo from previous test
		existingPhotoURL := stringPtr("https://google.com/new-photo.jpg")

		attemptedPhotoURL := stringPtr("https://facebook.com/different-photo.jpg")

		response, err := service.CompleteAuthentication(
			ctx,
			firebaseUID,
			email,
			provider,
			true,
			attemptedPhotoURL,
			nil,
			nil,
			nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, response.User.PhotoURL)
		assert.Equal(t, *existingPhotoURL, *response.User.PhotoURL, "Should keep original photo, not overwrite")
	})
}

// TestProviderTracking tests that all providers are tracked in linked_accounts
func TestProviderTracking(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockWorkspaceRepo := NewMockWorkspaceRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockAuditRepo := NewMockAuditRepository()

	service := NewAuthService(mockUserRepo, mockWorkspaceRepo, mockSessionRepo, mockAuditRepo, nil)

	ctx := context.Background()
	email := "multiauth@example.com"
	ipAddress := stringPtr("192.168.1.1")
	userAgent := stringPtr("TestAgent")

	t.Run("all provider authentications are tracked", func(t *testing.T) {
		providers := []struct {
			uid      string
			provider string
		}{
			{"google-uid-abc", "google.com"},
			{"facebook-uid-def", "facebook.com"},
			{"password-uid-ghi", "password"},
		}

		// Setup: manually create initial user (bypassing transaction-based createNewUser)
		firstProvider := providers[0]
		user := &entity.User{
			UserID:        uuid.New(),
			FirebaseUID:   &firstProvider.uid,
			Email:         email,
			Username:      "multiauthuser",
			Provider:      firstProvider.provider,
			EmailVerified: true,
			CreatedAt:     time.Now(),
		}
		err := mockUserRepo.Create(ctx, user)
		require.NoError(t, err)

		// Authenticate with each provider
		for _, p := range providers {
			response, err := service.CompleteAuthentication(
				ctx,
				p.uid,
				email,
				p.provider,
				true,
				nil,
				nil,
				ipAddress,
				userAgent,
			)

			require.NoError(t, err, "Provider %s authentication failed", p.provider)
			assert.NotNil(t, response)
			assert.False(t, response.IsNewUser, "Should link to existing account")
		}

		// Verify only one user exists
		users := mockUserRepo.GetAllUsers()
		assert.Len(t, users, 1)

		// Verify all providers are linked
		links := mockUserRepo.GetAllLinkedAccounts()
		assert.Len(t, links, len(providers), "All providers should be tracked")

		trackedProviders := make(map[string]string)
		for _, link := range links {
			trackedProviders[link.Provider] = link.ProviderUserID
		}

		for _, p := range providers {
			uid, found := trackedProviders[p.provider]
			assert.True(t, found, "Provider %s should be tracked", p.provider)
			assert.Equal(t, p.uid, uid, "Provider UID should match for %s", p.provider)
		}
	})
}

// TestAccountLinking_IdempotentProviderLinks tests that duplicate provider links don't cause errors
func TestAccountLinking_IdempotentProviderLinks(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockWorkspaceRepo := NewMockWorkspaceRepository()
	mockSessionRepo := NewMockSessionRepository()
	mockAuditRepo := NewMockAuditRepository()

	service := NewAuthService(mockUserRepo, mockWorkspaceRepo, mockSessionRepo, mockAuditRepo, nil)

	ctx := context.Background()
	email := "idempotent@example.com"
	firebaseUID := "google-uid-idempotent"
	provider := "google.com"

	t.Run("multiple logins with same provider don't create duplicate links", func(t *testing.T) {
		// Setup: manually create user (bypassing transaction-based createNewUser)
		user := &entity.User{
			UserID:        uuid.New(),
			FirebaseUID:   &firebaseUID,
			Email:         email,
			Username:      "idempotentuser",
			Provider:      provider,
			EmailVerified: true,
			CreatedAt:     time.Now(),
		}
		err := mockUserRepo.Create(ctx, user)
		require.NoError(t, err)

		// First login
		_, err = service.CompleteAuthentication(ctx, firebaseUID, email, provider, true, nil, nil, nil, nil)
		require.NoError(t, err)

		// Second login with same provider
		_, err = service.CompleteAuthentication(ctx, firebaseUID, email, provider, true, nil, nil, nil, nil)
		require.NoError(t, err)

		// Third login with same provider
		_, err = service.CompleteAuthentication(ctx, firebaseUID, email, provider, true, nil, nil, nil, nil)
		require.NoError(t, err)

		// Should have only one provider link
		links := mockUserRepo.GetAllLinkedAccounts()
		assert.Len(t, links, 1, "Should have only one link despite multiple logins")
		assert.Equal(t, provider, links[0].Provider)
		assert.Equal(t, firebaseUID, links[0].ProviderUserID)
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
