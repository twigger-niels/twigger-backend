// +build integration

package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"
	testhelpers "twigger-backend/backend/auth-service/infrastructure/database/testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuditTest(t *testing.T) (*PostgresAuditRepository, *sql.DB, context.Context, func()) {
	db := testhelpers.SetupTestDB(t)
	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	cleanup := func() {
		testhelpers.CleanupTestData(t, db)
		testhelpers.TeardownTestDB(t, db)
	}

	return repo.(*PostgresAuditRepository), db, ctx, cleanup
}

// createTestUser creates a test user for audit events
func createTestUser(t *testing.T, db *sql.DB, ctx context.Context, userID uuid.UUID) {
	t.Helper()
	query := `
		INSERT INTO users (user_id, email, username, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO NOTHING
	`
	_, err := db.ExecContext(ctx, query,
		userID,
		fmt.Sprintf("test-%s@example.com", userID.String()[:8]),
		fmt.Sprintf("testuser-%s", userID.String()[:8]))
	require.NoError(t, err, "Should create test user")
}

func TestAuditRepository_Integration_LogEvent(t *testing.T) {
	repo, db, ctx, cleanup := setupAuditTest(t)
	defer cleanup()

	t.Run("log event to partitioned table", func(t *testing.T) {
		userID := uuid.New()
		createTestUser(t, db, ctx, userID) // Create user first

		ipAddress := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		event := &entity.AuditEvent{
			UserID:    &userID,
			EventType: entity.EventUserLogin,
			Success:   true,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
			Metadata: map[string]interface{}{
				"provider": "google.com",
				"device":   "mobile",
			},
			CreatedAt: time.Now(),
		}

		err := repo.LogEvent(ctx, event)
		require.NoError(t, err, "Should successfully log event")
		assert.NotZero(t, event.ID, "Event ID should be set")
	})

	t.Run("log event without user ID (anonymous)", func(t *testing.T) {
		event := &entity.AuditEvent{
			UserID:    nil,
			EventType: entity.EventUserRegistered,
			Success:   false,
			Metadata: map[string]interface{}{
				"error": "invalid email",
				"email": "bad@example.com",
			},
		}

		err := repo.LogEvent(ctx, event)
		require.NoError(t, err, "Should allow nil user_id")
		assert.NotZero(t, event.ID)
	})

	t.Run("log multiple events and query", func(t *testing.T) {
		userID := uuid.New()
		createTestUser(t, db, ctx, userID) // Create user first

		// Log multiple events
		events := []*entity.AuditEvent{
			{
				UserID:    &userID,
				EventType: entity.EventUserLogin,
				Success:   true,
			},
			{
				UserID:    &userID,
				EventType: entity.EventUserLogin,
				Success:   false,
			},
			{
				UserID:    &userID,
				EventType: entity.EventUserLogout,
				Success:   true,
			},
		}

		for _, evt := range events {
			err := repo.LogEvent(ctx, evt)
			require.NoError(t, err)
		}

		// Query user events
		results, err := repo.GetUserEvents(ctx, userID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 3, "Should retrieve all 3 events")

		// Verify order (most recent first)
		assert.Equal(t, entity.EventUserLogout, results[0].EventType)
	})
}

func TestAuditRepository_Integration_GetUserEventsByType(t *testing.T) {
	repo, db, ctx, cleanup := setupAuditTest(t)
	defer cleanup()

	userID := uuid.New()
	createTestUser(t, db, ctx, userID) // Create user first

	// Log different event types
	events := []*entity.AuditEvent{
		{UserID: &userID, EventType: entity.EventUserLogin, Success: true},
		{UserID: &userID, EventType: entity.EventUserLogin, Success: false},
		{UserID: &userID, EventType: entity.EventUserLogout, Success: true},
		{UserID: &userID, EventType: entity.EventUserRegistered, Success: true},
	}

	for _, evt := range events {
		err := repo.LogEvent(ctx, evt)
		require.NoError(t, err)
	}

	// Query only login events
	loginEvents, err := repo.GetUserEventsByType(ctx, userID, entity.EventUserLogin, 10)
	require.NoError(t, err)
	assert.Len(t, loginEvents, 2, "Should retrieve 2 login events")

	for _, evt := range loginEvents {
		assert.Equal(t, entity.EventUserLogin, evt.EventType)
	}
}

func TestAuditRepository_Integration_GetFailedLoginAttempts(t *testing.T) {
	repo, db, ctx, cleanup := setupAuditTest(t)
	defer cleanup()

	userID := uuid.New()
	createTestUser(t, db, ctx, userID) // Create user first
	since := time.Now().Add(-1 * time.Hour)

	// Log successful and failed logins
	events := []*entity.AuditEvent{
		{UserID: &userID, EventType: entity.EventUserLogin, Success: false, CreatedAt: time.Now()},
		{UserID: &userID, EventType: entity.EventUserLogin, Success: false, CreatedAt: time.Now()},
		{UserID: &userID, EventType: entity.EventUserLogin, Success: true, CreatedAt: time.Now()},
		{UserID: &userID, EventType: entity.EventUserLogin, Success: false, CreatedAt: time.Now()},
	}

	for _, evt := range events {
		err := repo.LogEvent(ctx, evt)
		require.NoError(t, err)
	}

	// Count failed attempts
	count, err := repo.GetFailedLoginAttempts(ctx, userID, since)
	require.NoError(t, err)
	assert.Equal(t, 3, count, "Should count 3 failed login attempts")
}

func TestAuditRepository_Integration_CountEventsByType(t *testing.T) {
	repo, db, ctx, cleanup := setupAuditTest(t)
	defer cleanup()

	startDate := time.Now().Add(-1 * time.Hour)
	endDate := time.Now().Add(1 * time.Hour)

	// Log various event types
	for i := 0; i < 5; i++ {
		userID := uuid.New()
		createTestUser(t, db, ctx, userID) // Create user first
		err := repo.LogEvent(ctx, &entity.AuditEvent{
			UserID:    &userID,
			EventType: entity.EventUserRegistered,
			Success:   true,
		})
		require.NoError(t, err)
	}

	for i := 0; i < 3; i++ {
		userID := uuid.New()
		createTestUser(t, db, ctx, userID) // Create user first
		err := repo.LogEvent(ctx, &entity.AuditEvent{
			UserID:    &userID,
			EventType: entity.EventUserLogin,
			Success:   true,
		})
		require.NoError(t, err)
	}

	// Count registrations
	count, err := repo.CountEventsByType(ctx, entity.EventUserRegistered, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count, "Should count 5 registration events")

	// Count logins
	loginCount, err := repo.CountEventsByType(ctx, entity.EventUserLogin, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, int64(3), loginCount, "Should count 3 login events")
}

func TestAuditRepository_Integration_MetadataJSONB(t *testing.T) {
	repo, db, ctx, cleanup := setupAuditTest(t)
	defer cleanup()

	userID := uuid.New()
	createTestUser(t, db, ctx, userID) // Create user first

	// Log event with complex metadata
	event := &entity.AuditEvent{
		UserID:    &userID,
		EventType: entity.EventUserLogin,
		Success:   true,
		Metadata: map[string]interface{}{
			"provider": "google.com",
			"device": map[string]interface{}{
				"type":    "mobile",
				"os":      "iOS",
				"version": "17.0",
			},
			"location": map[string]interface{}{
				"country": "US",
				"city":    "San Francisco",
			},
			"count": 42,
			"flags": []string{"new_device", "suspicious_location"},
		},
	}

	err := repo.LogEvent(ctx, event)
	require.NoError(t, err)

	// Retrieve and verify metadata
	events, err := repo.GetUserEvents(ctx, userID, 1, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)

	retrieved := events[0]
	assert.NotNil(t, retrieved.Metadata)
	assert.Equal(t, "google.com", retrieved.Metadata["provider"])

	// Verify nested objects
	device, ok := retrieved.Metadata["device"].(map[string]interface{})
	require.True(t, ok, "Device should be a map")
	assert.Equal(t, "mobile", device["type"])
	assert.Equal(t, "iOS", device["os"])

	// Verify array
	flags, ok := retrieved.Metadata["flags"].([]interface{})
	require.True(t, ok, "Flags should be an array")
	assert.Len(t, flags, 2)
}

func TestAuditRepository_Integration_GetEventsByDateRange(t *testing.T) {
	repo, db, ctx, cleanup := setupAuditTest(t)
	defer cleanup()

	// Log events at different times
	now := time.Now()
	events := []*entity.AuditEvent{
		{
			UserID:    &uuid.UUID{},
			EventType: entity.EventUserLogin,
			Success:   true,
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			UserID:    &uuid.UUID{},
			EventType: entity.EventUserLogin,
			Success:   true,
			CreatedAt: now.Add(-1 * time.Hour),
		},
		{
			UserID:    &uuid.UUID{},
			EventType: entity.EventUserLogin,
			Success:   true,
			CreatedAt: now,
		},
	}

	for _, evt := range events {
		*evt.UserID = uuid.New() // Generate unique ID for each
		createTestUser(t, db, ctx, *evt.UserID) // Create user first
		err := repo.LogEvent(ctx, evt)
		require.NoError(t, err)
	}

	// Query events in specific date range
	startDate := now.Add(-90 * time.Minute)
	endDate := now.Add(10 * time.Minute)

	results, err := repo.GetEventsByDateRange(ctx, startDate, endDate)
	require.NoError(t, err)
	assert.Len(t, results, 2, "Should retrieve 2 events within date range")
}

func TestAuditRepository_Integration_PartitionedTableInsert(t *testing.T) {
	repo, db, ctx, cleanup := setupAuditTest(t)
	defer cleanup()

	t.Run("verify partition exists for current month", func(t *testing.T) {
		// The migration 008 should create the default partition
		// Let's verify we can insert events

		userID := uuid.New()
		createTestUser(t, db, ctx, userID) // Create user first
		event := &entity.AuditEvent{
			UserID:    &userID,
			EventType: entity.EventUserLogin,
			Success:   true,
			CreatedAt: time.Now(),
		}

		err := repo.LogEvent(ctx, event)
		require.NoError(t, err, "Should insert into partitioned table")

		// Verify event was inserted
		events, err := repo.GetUserEvents(ctx, userID, 1, 0)
		require.NoError(t, err)
		require.Len(t, events, 1)
		assert.Equal(t, entity.EventUserLogin, events[0].EventType)
	})
}
