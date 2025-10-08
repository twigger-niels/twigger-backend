package persistence

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditRepository_LogEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	t.Run("successful event logging", func(t *testing.T) {
		userID := uuid.New()
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
			},
		}

		mock.ExpectQuery(`INSERT INTO auth_audit_log`).
			WithArgs(
				event.UserID,
				event.EventType,
				event.Success,
				event.IPAddress,
				event.UserAgent,
				sqlmock.AnyArg(), // metadata JSON
				sqlmock.AnyArg(), // created_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := repo.LogEvent(ctx, event)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), event.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("event with nil user ID (anonymous event)", func(t *testing.T) {
		event := &entity.AuditEvent{
			UserID:    nil,
			EventType: entity.EventUserRegistered,
			Success:   false,
			Metadata: map[string]interface{}{
				"error": "invalid email",
			},
		}

		mock.ExpectQuery(`INSERT INTO auth_audit_log`).
			WithArgs(
				nil, // nil user_id
				event.EventType,
				event.Success,
				nil, // nil ip_address
				nil, // nil user_agent
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

		err := repo.LogEvent(ctx, event)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), event.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("event with empty metadata", func(t *testing.T) {
		userID := uuid.New()
		event := &entity.AuditEvent{
			UserID:    &userID,
			EventType: entity.EventUserLogout,
			Success:   true,
			Metadata:  nil, // no metadata
		}

		mock.ExpectQuery(`INSERT INTO auth_audit_log`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

		err := repo.LogEvent(ctx, event)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sets created_at if not provided", func(t *testing.T) {
		userID := uuid.New()
		event := &entity.AuditEvent{
			UserID:    &userID,
			EventType: entity.EventUserLogin,
			Success:   true,
			CreatedAt: time.Time{}, // zero time
		}

		mock.ExpectQuery(`INSERT INTO auth_audit_log`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))

		beforeLog := time.Now()
		err := repo.LogEvent(ctx, event)
		afterLog := time.Now()

		assert.NoError(t, err)
		assert.False(t, event.CreatedAt.IsZero(), "CreatedAt should be set")
		assert.True(t, event.CreatedAt.After(beforeLog) || event.CreatedAt.Equal(beforeLog))
		assert.True(t, event.CreatedAt.Before(afterLog) || event.CreatedAt.Equal(afterLog))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		userID := uuid.New()
		event := &entity.AuditEvent{
			UserID:    &userID,
			EventType: entity.EventUserLogin,
			Success:   true,
		}

		mock.ExpectQuery(`INSERT INTO auth_audit_log`).
			WillReturnError(sql.ErrConnDone)

		err := repo.LogEvent(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to log audit event")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuditRepository_GetUserEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	t.Run("retrieve user events with pagination", func(t *testing.T) {
		userID := uuid.New()
		ipAddress := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "event_type", "success",
			"ip_address", "user_agent", "metadata", "created_at",
		}).
			AddRow(1, userID, "user_login", true, ipAddress, userAgent, []byte(`{"provider":"google.com"}`), time.Now()).
			AddRow(2, userID, "user_registered", true, ipAddress, userAgent, []byte(`{}`), time.Now().Add(-1*time.Hour))

		mock.ExpectQuery(`SELECT .+ FROM auth_audit_log WHERE user_id = \$1`).
			WithArgs(userID, 10, 0).
			WillReturnRows(rows)

		events, err := repo.GetUserEvents(ctx, userID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.Equal(t, entity.EventUserLogin, events[0].EventType)
		assert.Equal(t, entity.EventUserRegistered, events[1].EventType)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result set", func(t *testing.T) {
		userID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "event_type", "success",
			"ip_address", "user_agent", "metadata", "created_at",
		})

		mock.ExpectQuery(`SELECT .+ FROM auth_audit_log WHERE user_id = \$1`).
			WithArgs(userID, 10, 0).
			WillReturnRows(rows)

		events, err := repo.GetUserEvents(ctx, userID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, events, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuditRepository_GetUserEventsByType(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	t.Run("filter events by type", func(t *testing.T) {
		userID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "event_type", "success",
			"ip_address", "user_agent", "metadata", "created_at",
		}).
			AddRow(1, userID, "user_login", true, nil, nil, []byte(`{}`), time.Now()).
			AddRow(2, userID, "user_login", false, nil, nil, []byte(`{}`), time.Now().Add(-1*time.Hour))

		mock.ExpectQuery(`SELECT .+ FROM auth_audit_log WHERE user_id = \$1 AND event_type = \$2`).
			WithArgs(userID, entity.EventUserLogin, 10).
			WillReturnRows(rows)

		events, err := repo.GetUserEventsByType(ctx, userID, entity.EventUserLogin, 10)
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.True(t, events[0].Success)
		assert.False(t, events[1].Success)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuditRepository_GetEventsByDateRange(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	t.Run("retrieve events within date range", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "event_type", "success",
			"ip_address", "user_agent", "metadata", "created_at",
		}).
			AddRow(1, uuid.New(), "user_login", true, nil, nil, []byte(`{}`), time.Now()).
			AddRow(2, uuid.New(), "user_registered", true, nil, nil, []byte(`{}`), time.Now().Add(-1*time.Hour))

		mock.ExpectQuery(`SELECT .+ FROM auth_audit_log WHERE created_at >= \$1 AND created_at <= \$2`).
			WithArgs(startDate, endDate).
			WillReturnRows(rows)

		events, err := repo.GetEventsByDateRange(ctx, startDate, endDate)
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuditRepository_CountEventsByType(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	t.Run("count events by type and date range", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		rows := sqlmock.NewRows([]string{"count"}).AddRow(42)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth_audit_log`).
			WithArgs(entity.EventUserLogin, startDate, endDate).
			WillReturnRows(rows)

		count, err := repo.CountEventsByType(ctx, entity.EventUserLogin, startDate, endDate)
		assert.NoError(t, err)
		assert.Equal(t, int64(42), count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("zero count for no events", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth_audit_log`).
			WithArgs(entity.EventUserRegistered, startDate, endDate).
			WillReturnRows(rows)

		count, err := repo.CountEventsByType(ctx, entity.EventUserRegistered, startDate, endDate)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuditRepository_GetFailedLoginAttempts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	t.Run("count failed login attempts", func(t *testing.T) {
		userID := uuid.New()
		since := time.Now().Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"count"}).AddRow(3)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth_audit_log`).
			WithArgs(userID, entity.EventUserLogin, since).
			WillReturnRows(rows)

		count, err := repo.GetFailedLoginAttempts(ctx, userID, since)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("zero failed attempts", func(t *testing.T) {
		userID := uuid.New()
		since := time.Now().Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM auth_audit_log`).
			WithArgs(userID, entity.EventUserLogin, since).
			WillReturnRows(rows)

		count, err := repo.GetFailedLoginAttempts(ctx, userID, since)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuditRepository_MetadataHandling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditRepository(db)
	ctx := context.Background()

	t.Run("complex metadata structure", func(t *testing.T) {
		userID := uuid.New()
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

		mock.ExpectQuery(`INSERT INTO auth_audit_log`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := repo.LogEvent(ctx, event)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("metadata unmarshaling", func(t *testing.T) {
		userID := uuid.New()
		metadataJSON := []byte(`{"provider":"google.com","device":{"type":"mobile"}}`)

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "event_type", "success",
			"ip_address", "user_agent", "metadata", "created_at",
		}).AddRow(1, userID, "user_login", true, nil, nil, metadataJSON, time.Now())

		mock.ExpectQuery(`SELECT .+ FROM auth_audit_log WHERE user_id = \$1`).
			WithArgs(userID, 10, 0).
			WillReturnRows(rows)

		events, err := repo.GetUserEvents(ctx, userID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.NotNil(t, events[0].Metadata)
		assert.Equal(t, "google.com", events[0].Metadata["provider"])

		device, ok := events[0].Metadata["device"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "mobile", device["type"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuditEventType_IsValid(t *testing.T) {
	t.Run("valid event types", func(t *testing.T) {
		validEvents := []entity.AuditEventType{
			entity.EventUserRegistered,
			entity.EventUserLogin,
			entity.EventUserLogout,
			entity.EventTokenRefresh,
			entity.EventSessionRevoked,
			entity.EventAccountDeleted,
			entity.EventPasswordReset,
			entity.EventAccountLinked,
		}

		for _, event := range validEvents {
			assert.True(t, event.IsValid(), "Event %s should be valid", event)
		}
	})

	t.Run("invalid event type", func(t *testing.T) {
		invalidEvent := entity.AuditEventType("invalid_event")
		assert.False(t, invalidEvent.IsValid())
	})

	t.Run("empty event type", func(t *testing.T) {
		emptyEvent := entity.AuditEventType("")
		assert.False(t, emptyEvent.IsValid())
	})
}
