package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

// PostgresLanguageRepository implements LanguageRepository using PostgreSQL
type PostgresLanguageRepository struct {
	db *sql.DB
}

// NewPostgresLanguageRepository creates a new PostgreSQL language repository
func NewPostgresLanguageRepository(db *sql.DB) repository.LanguageRepository {
	return &PostgresLanguageRepository{db: db}
}

// FindByID retrieves a language by its UUID
func (r *PostgresLanguageRepository) FindByID(ctx context.Context, languageID string) (*entity.Language, error) {
	query := `
		SELECT language_id, language_code, language_name, native_name, is_active, created_at
		FROM languages
		WHERE language_id = $1
	`

	var lang entity.Language
	err := r.db.QueryRowContext(ctx, query, languageID).Scan(
		&lang.LanguageID,
		&lang.LanguageCode,
		&lang.LanguageName,
		&lang.NativeName,
		&lang.IsActive,
		&lang.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("language not found: %s", languageID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find language: %w", err)
	}

	return &lang, nil
}

// FindByCode retrieves a language by its ISO 639 language code
func (r *PostgresLanguageRepository) FindByCode(ctx context.Context, languageCode string) (*entity.Language, error) {
	query := `
		SELECT language_id, language_code, language_name, native_name, is_active, created_at
		FROM languages
		WHERE language_code = $1
	`

	var lang entity.Language
	err := r.db.QueryRowContext(ctx, query, languageCode).Scan(
		&lang.LanguageID,
		&lang.LanguageCode,
		&lang.LanguageName,
		&lang.NativeName,
		&lang.IsActive,
		&lang.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("language not found: %s", languageCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find language by code: %w", err)
	}

	return &lang, nil
}

// FindAll retrieves all languages
func (r *PostgresLanguageRepository) FindAll(ctx context.Context) ([]*entity.Language, error) {
	query := `
		SELECT language_id, language_code, language_name, native_name, is_active, created_at
		FROM languages
		ORDER BY language_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query languages: %w", err)
	}
	defer rows.Close()

	var languages []*entity.Language
	for rows.Next() {
		var lang entity.Language
		if err := rows.Scan(
			&lang.LanguageID,
			&lang.LanguageCode,
			&lang.LanguageName,
			&lang.NativeName,
			&lang.IsActive,
			&lang.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan language: %w", err)
		}
		languages = append(languages, &lang)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating languages: %w", err)
	}

	return languages, nil
}

// FindActive retrieves all active languages
func (r *PostgresLanguageRepository) FindActive(ctx context.Context) ([]*entity.Language, error) {
	query := `
		SELECT language_id, language_code, language_name, native_name, is_active, created_at
		FROM languages
		WHERE is_active = true
		ORDER BY language_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active languages: %w", err)
	}
	defer rows.Close()

	var languages []*entity.Language
	for rows.Next() {
		var lang entity.Language
		if err := rows.Scan(
			&lang.LanguageID,
			&lang.LanguageCode,
			&lang.LanguageName,
			&lang.NativeName,
			&lang.IsActive,
			&lang.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan language: %w", err)
		}
		languages = append(languages, &lang)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active languages: %w", err)
	}

	return languages, nil
}

// Create creates a new language
func (r *PostgresLanguageRepository) Create(ctx context.Context, language *entity.Language) error {
	if err := language.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO languages (language_id, language_code, language_name, native_name, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	language.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		language.LanguageID,
		language.LanguageCode,
		language.LanguageName,
		language.NativeName,
		language.IsActive,
		language.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create language: %w", err)
	}

	return nil
}

// Update updates an existing language
func (r *PostgresLanguageRepository) Update(ctx context.Context, language *entity.Language) error {
	if err := language.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE languages
		SET language_code = $2, language_name = $3, native_name = $4, is_active = $5
		WHERE language_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		language.LanguageID,
		language.LanguageCode,
		language.LanguageName,
		language.NativeName,
		language.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update language: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("language not found: %s", language.LanguageID)
	}

	return nil
}

// Delete deletes a language by ID
func (r *PostgresLanguageRepository) Delete(ctx context.Context, languageID string) error {
	query := `DELETE FROM languages WHERE language_id = $1`

	result, err := r.db.ExecContext(ctx, query, languageID)
	if err != nil {
		return fmt.Errorf("failed to delete language: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("language not found: %s", languageID)
	}

	return nil
}
