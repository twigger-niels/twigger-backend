package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"twigger-backend/backend/auth-service/domain/entity"
	"twigger-backend/backend/auth-service/domain/repository"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) repository.UserRepository {
	return &PostgresUserRepository{db: db}
}

// Create inserts a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			user_id, firebase_uid, email, username, email_verified,
			phone_number, photo_url, provider, preferred_language_id,
			country_id, location, detected_hardiness_zone,
			created_at, last_login_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			ST_GeogFromText($11), $12, $13, $14
		)
	`

	// Generate UUID if not provided
	if user.UserID == uuid.Nil {
		user.UserID = uuid.New()
	}

	// Set timestamps
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	// Validate location WKT format to prevent SQL injection
	if user.Location != nil {
		if err := validateWKT(*user.Location); err != nil {
			return fmt.Errorf("invalid location format: %w", err)
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.FirebaseUID,
		user.Email,
		user.Username,
		user.EmailVerified,
		user.PhoneNumber,
		user.PhotoURL,
		user.Provider,
		user.PreferredLanguageID,
		user.CountryID,
		user.Location,
		user.DetectedHardinessZone,
		user.CreatedAt,
		user.LastLoginAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by user_id
func (r *PostgresUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	query := `
		SELECT
			user_id, firebase_uid, email, username, email_verified,
			phone_number, photo_url, provider, preferred_language_id,
			country_id, ST_AsText(location) as location,
			detected_hardiness_zone, created_at, last_login_at, deleted_at
		FROM users
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.FirebaseUID,
		&user.Email,
		&user.Username,
		&user.EmailVerified,
		&user.PhoneNumber,
		&user.PhotoURL,
		&user.Provider,
		&user.PreferredLanguageID,
		&user.CountryID,
		&user.Location,
		&user.DetectedHardinessZone,
		&user.CreatedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows // Return standard error, don't expose user ID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByFirebaseUID retrieves a user by firebase_uid
func (r *PostgresUserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	query := `
		SELECT
			user_id, firebase_uid, email, username, email_verified,
			phone_number, photo_url, provider, preferred_language_id,
			country_id, ST_AsText(location) as location,
			detected_hardiness_zone, created_at, last_login_at, deleted_at
		FROM users
		WHERE firebase_uid = $1 AND deleted_at IS NULL
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, firebaseUID).Scan(
		&user.UserID,
		&user.FirebaseUID,
		&user.Email,
		&user.Username,
		&user.EmailVerified,
		&user.PhoneNumber,
		&user.PhotoURL,
		&user.Provider,
		&user.PreferredLanguageID,
		&user.CountryID,
		&user.Location,
		&user.DetectedHardinessZone,
		&user.CreatedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows // Return standard error, don't expose Firebase UID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT
			user_id, firebase_uid, email, username, email_verified,
			phone_number, photo_url, provider, preferred_language_id,
			country_id, ST_AsText(location) as location,
			detected_hardiness_zone, created_at, last_login_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.FirebaseUID,
		&user.Email,
		&user.Username,
		&user.EmailVerified,
		&user.PhoneNumber,
		&user.PhotoURL,
		&user.Provider,
		&user.PreferredLanguageID,
		&user.CountryID,
		&user.Location,
		&user.DetectedHardinessZone,
		&user.CreatedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows // Return standard error, don't expose email
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) error {
	// Validate location WKT format to prevent SQL injection
	if user.Location != nil {
		if err := validateWKT(*user.Location); err != nil {
			return fmt.Errorf("invalid location format: %w", err)
		}
	}

	query := `
		UPDATE users
		SET
			firebase_uid = $2,
			email = $3,
			username = $4,
			email_verified = $5,
			phone_number = $6,
			photo_url = $7,
			provider = $8,
			preferred_language_id = $9,
			country_id = $10,
			location = ST_GeogFromText($11),
			detected_hardiness_zone = $12,
			last_login_at = $13
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.FirebaseUID,
		user.Email,
		user.Username,
		user.EmailVerified,
		user.PhoneNumber,
		user.PhotoURL,
		user.Provider,
		user.PreferredLanguageID,
		user.CountryID,
		user.Location,
		user.DetectedHardinessZone,
		user.LastLoginAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Return standard error, don't expose user ID
	}

	return nil
}

// SoftDelete marks a user as deleted
func (r *PostgresUserRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = $2
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Return standard error, don't expose user ID
	}

	return nil
}

// UpdateLastLogin updates the last_login_at timestamp
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $2
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Return standard error, don't expose user ID
	}

	return nil
}

// LinkProvider links a social provider to a user account
func (r *PostgresUserRepository) LinkProvider(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error {
	query := `
		INSERT INTO linked_accounts (user_id, provider, provider_user_id, linked_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, provider_user_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, provider, providerUserID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to link provider: %w", err)
	}

	return nil
}

// GetLinkedAccounts retrieves all linked accounts for a user
func (r *PostgresUserRepository) GetLinkedAccounts(ctx context.Context, userID uuid.UUID) ([]*entity.LinkedAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, linked_at
		FROM linked_accounts
		WHERE user_id = $1
		ORDER BY linked_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*entity.LinkedAccount
	for rows.Next() {
		account := &entity.LinkedAccount{}
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderUserID,
			&account.LinkedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan linked account: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating linked accounts: %w", err)
	}

	return accounts, nil
}

// GetUserWorkspaces retrieves all workspaces for a user
func (r *PostgresUserRepository) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error) {
	query := `
		SELECT w.workspace_id, w.owner_id, w.name, w.created_at, w.updated_at
		FROM workspaces w
		INNER JOIN workspace_members wm ON w.workspace_id = wm.workspace_id
		WHERE wm.user_id = $1
		ORDER BY w.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*entity.Workspace
	for rows.Next() {
		workspace := &entity.Workspace{}
		err := rows.Scan(
			&workspace.WorkspaceID,
			&workspace.OwnerID,
			&workspace.Name,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspaces: %w", err)
	}

	return workspaces, nil
}

// scanUser is a helper function to scan a user row (not used in current implementation but useful for future)
func scanUser(scanner interface{ Scan(...interface{}) error }) (*entity.User, error) {
	user := &entity.User{}
	var deviceInfo []byte

	err := scanner.Scan(
		&user.UserID,
		&user.FirebaseUID,
		&user.Email,
		&user.Username,
		&user.EmailVerified,
		&user.PhoneNumber,
		&user.PhotoURL,
		&user.Provider,
		&user.PreferredLanguageID,
		&user.CountryID,
		&user.Location,
		&user.DetectedHardinessZone,
		&user.CreatedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse device info JSONB if present
	if len(deviceInfo) > 0 {
		var info map[string]interface{}
		if err := json.Unmarshal(deviceInfo, &info); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
		}
	}

	return user, nil
}

// validateWKT validates a Well-Known Text (WKT) string to prevent SQL injection
// Accepts POINT format: "POINT(longitude latitude)"
// Example: "POINT(-122.4194 37.7749)"
func validateWKT(wkt string) error {
	if wkt == "" {
		return nil // Empty string is valid (NULL location)
	}

	// Trim whitespace
	wkt = strings.TrimSpace(wkt)

	// Validate POINT format with regex
	// Pattern: POINT(decimal decimal) where decimal can be negative and have optional decimal places
	// Longitude range: -180 to 180
	// Latitude range: -90 to 90
	pointRegex := regexp.MustCompile(`^POINT\s*\(\s*(-?[0-9]{1,3}(?:\.[0-9]+)?)\s+(-?[0-9]{1,2}(?:\.[0-9]+)?)\s*\)$`)

	matches := pointRegex.FindStringSubmatch(wkt)
	if matches == nil {
		return fmt.Errorf("invalid WKT format, expected POINT(longitude latitude)")
	}

	// Extract and validate coordinates
	// Note: matches[0] is the full match, matches[1] is longitude, matches[2] is latitude
	// Actual coordinate validation would require parsing to float and checking bounds
	// For SQL injection prevention, regex validation is sufficient

	return nil
}
