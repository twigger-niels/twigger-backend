package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"
	"twigger-backend/backend/auth-service/domain/repository"

	"github.com/google/uuid"
)

// PostgresAuditRepository implements AuditRepository using PostgreSQL
type PostgresAuditRepository struct {
	db *sql.DB
}

// NewPostgresAuditRepository creates a new PostgresAuditRepository
func NewPostgresAuditRepository(db *sql.DB) repository.AuditRepository {
	return &PostgresAuditRepository{db: db}
}

// LogEvent inserts a new audit event
func (r *PostgresAuditRepository) LogEvent(ctx context.Context, event *entity.AuditEvent) error {
	query := `
		INSERT INTO auth_audit_log (
			user_id, event_type, success, ip_address,
			user_agent, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	// Set timestamp if not provided
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Marshal metadata to JSONB
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		event.UserID,
		event.EventType,
		event.Success,
		event.IPAddress,
		event.UserAgent,
		metadataJSON,
		event.CreatedAt,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil
}

// GetUserEvents retrieves audit events for a user with pagination
func (r *PostgresAuditRepository) GetUserEvents(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*entity.AuditEvent, error) {
	query := `
		SELECT
			id, user_id, event_type, success, ip_address,
			user_agent, metadata, created_at
		FROM auth_audit_log
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.queryAuditEvents(ctx, query, userID, limit, offset)
}

// GetUserEventsByType retrieves audit events for a user filtered by event type
func (r *PostgresAuditRepository) GetUserEventsByType(ctx context.Context, userID uuid.UUID, eventType entity.AuditEventType, limit int) ([]*entity.AuditEvent, error) {
	query := `
		SELECT
			id, user_id, event_type, success, ip_address,
			user_agent, metadata, created_at
		FROM auth_audit_log
		WHERE user_id = $1 AND event_type = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	return r.queryAuditEvents(ctx, query, userID, eventType, limit)
}

// GetEventsByDateRange retrieves audit events within a date range
func (r *PostgresAuditRepository) GetEventsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.AuditEvent, error) {
	query := `
		SELECT
			id, user_id, event_type, success, ip_address,
			user_agent, metadata, created_at
		FROM auth_audit_log
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
	`

	return r.queryAuditEvents(ctx, query, startDate, endDate)
}

// CountEventsByType counts audit events by type within a date range
func (r *PostgresAuditRepository) CountEventsByType(ctx context.Context, eventType entity.AuditEventType, startDate, endDate time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM auth_audit_log
		WHERE event_type = $1
		  AND created_at >= $2
		  AND created_at <= $3
	`

	var count int64
	err := r.db.QueryRowContext(ctx, query, eventType, startDate, endDate).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// GetFailedLoginAttempts counts failed login attempts for a user since a given time
func (r *PostgresAuditRepository) GetFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM auth_audit_log
		WHERE user_id = $1
		  AND event_type = $2
		  AND success = false
		  AND created_at >= $3
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, entity.EventUserLogin, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count failed login attempts: %w", err)
	}

	return count, nil
}

// queryAuditEvents is a helper function to query multiple audit events
func (r *PostgresAuditRepository) queryAuditEvents(ctx context.Context, query string, args ...interface{}) ([]*entity.AuditEvent, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []*entity.AuditEvent
	for rows.Next() {
		event := &entity.AuditEvent{}
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.EventType,
			&event.Success,
			&event.IPAddress,
			&event.UserAgent,
			&metadataJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit events: %w", err)
	}

	return events, nil
}
