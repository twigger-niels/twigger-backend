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

// PostgresSessionRepository implements SessionRepository using PostgreSQL
type PostgresSessionRepository struct {
	db *sql.DB
}

// NewPostgresSessionRepository creates a new PostgresSessionRepository
func NewPostgresSessionRepository(db *sql.DB) repository.SessionRepository {
	return &PostgresSessionRepository{db: db}
}

// Create inserts a new session
func (r *PostgresSessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO auth_sessions (
			session_id, user_id, device_id, device_info,
			ip_address, user_agent, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	// Generate UUID if not provided
	if session.SessionID == uuid.Nil {
		session.SessionID = uuid.New()
	}

	// Set timestamps
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	// Marshal device_info to JSONB
	deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		session.SessionID,
		session.UserID,
		session.DeviceID,
		deviceInfoJSON,
		session.IPAddress,
		session.UserAgent,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by session_id
func (r *PostgresSessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*entity.Session, error) {
	query := `
		SELECT
			session_id, user_id, device_id, device_info,
			ip_address, user_agent, created_at, expires_at, revoked_at
		FROM auth_sessions
		WHERE session_id = $1
	`

	session := &entity.Session{}
	var deviceInfoJSON []byte

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.DeviceID,
		&deviceInfoJSON,
		&session.IPAddress,
		&session.UserAgent,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Unmarshal device info
	if len(deviceInfoJSON) > 0 {
		if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
		}
	}

	return session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *PostgresSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	query := `
		SELECT
			session_id, user_id, device_id, device_info,
			ip_address, user_agent, created_at, expires_at, revoked_at
		FROM auth_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	return r.querySessions(ctx, query, userID)
}

// GetActiveByUserID retrieves all active sessions for a user
func (r *PostgresSessionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	query := `
		SELECT
			session_id, user_id, device_id, device_info,
			ip_address, user_agent, created_at, expires_at, revoked_at
		FROM auth_sessions
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	return r.querySessions(ctx, query, userID)
}

// Revoke marks a session as revoked
func (r *PostgresSessionRepository) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE auth_sessions
		SET revoked_at = $2
		WHERE session_id = $1 AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, sessionID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already revoked: %s", sessionID)
	}

	return nil
}

// RevokeAllForUser revokes all sessions for a user
func (r *PostgresSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth_sessions
		SET revoked_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to revoke all sessions: %w", err)
	}

	return nil
}

// RevokeByDeviceID revokes all sessions for a specific device
func (r *PostgresSessionRepository) RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	query := `
		UPDATE auth_sessions
		SET revoked_at = $3
		WHERE user_id = $1 AND device_id = $2 AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID, deviceID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to revoke sessions by device: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *PostgresSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM auth_sessions
		WHERE expires_at < NOW()
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// querySessions is a helper function to query multiple sessions
func (r *PostgresSessionRepository) querySessions(ctx context.Context, query string, args ...interface{}) ([]*entity.Session, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*entity.Session
	for rows.Next() {
		session := &entity.Session{}
		var deviceInfoJSON []byte

		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.DeviceID,
			&deviceInfoJSON,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Unmarshal device info
		if len(deviceInfoJSON) > 0 {
			if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
				return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
			}
		}

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}
