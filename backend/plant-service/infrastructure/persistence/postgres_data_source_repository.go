package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

type postgresDataSourceRepository struct {
	db *sql.DB
}

// NewPostgresDataSourceRepository creates a new PostgreSQL data source repository
func NewPostgresDataSourceRepository(db *sql.DB) repository.DataSourceRepository {
	return &postgresDataSourceRepository{db: db}
}

func (r *postgresDataSourceRepository) FindByID(ctx context.Context, sourceID string) (*entity.DataSource, error) {
	query := `
		SELECT source_id, source_name, source_type, website_url,
		       reliability_score, last_verified, created_at
		FROM data_sources
		WHERE source_id = $1
	`

	var source entity.DataSource
	err := r.db.QueryRowContext(ctx, query, sourceID).Scan(
		&source.SourceID,
		&source.SourceName,
		&source.SourceType,
		&source.WebsiteURL,
		&source.ReliabilityScore,
		&source.LastVerified,
		&source.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("data source not found: %s", sourceID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find data source: %w", err)
	}

	return &source, nil
}

func (r *postgresDataSourceRepository) FindAll(ctx context.Context) ([]*entity.DataSource, error) {
	query := `
		SELECT source_id, source_name, source_type, website_url,
		       reliability_score, last_verified, created_at
		FROM data_sources
		ORDER BY source_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query data sources: %w", err)
	}
	defer rows.Close()

	return r.scanDataSources(rows)
}

func (r *postgresDataSourceRepository) FindByType(ctx context.Context, sourceType string) ([]*entity.DataSource, error) {
	query := `
		SELECT source_id, source_name, source_type, website_url,
		       reliability_score, last_verified, created_at
		FROM data_sources
		WHERE source_type = $1
		ORDER BY source_name
	`

	rows, err := r.db.QueryContext(ctx, query, sourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to query data sources by type: %w", err)
	}
	defer rows.Close()

	return r.scanDataSources(rows)
}

func (r *postgresDataSourceRepository) FindVerified(ctx context.Context) ([]*entity.DataSource, error) {
	query := `
		SELECT source_id, source_name, source_type, website_url,
		       reliability_score, last_verified, created_at
		FROM data_sources
		WHERE last_verified IS NOT NULL
		ORDER BY reliability_score DESC, source_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query verified data sources: %w", err)
	}
	defer rows.Close()

	return r.scanDataSources(rows)
}

func (r *postgresDataSourceRepository) Create(ctx context.Context, source *entity.DataSource) error {
	if err := source.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO data_sources (source_id, source_name, source_type, website_url,
		                          reliability_score, last_verified, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	source.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		source.SourceID,
		source.SourceName,
		source.SourceType,
		source.WebsiteURL,
		source.ReliabilityScore,
		source.LastVerified,
		source.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create data source: %w", err)
	}

	return nil
}

func (r *postgresDataSourceRepository) Update(ctx context.Context, source *entity.DataSource) error {
	if err := source.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE data_sources
		SET source_name = $2, source_type = $3, website_url = $4,
		    reliability_score = $5, last_verified = $6
		WHERE source_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		source.SourceID,
		source.SourceName,
		source.SourceType,
		source.WebsiteURL,
		source.ReliabilityScore,
		source.LastVerified,
	)

	if err != nil {
		return fmt.Errorf("failed to update data source: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("data source not found: %s", source.SourceID)
	}

	return nil
}

func (r *postgresDataSourceRepository) Delete(ctx context.Context, sourceID string) error {
	query := `DELETE FROM data_sources WHERE source_id = $1`

	result, err := r.db.ExecContext(ctx, query, sourceID)
	if err != nil {
		return fmt.Errorf("failed to delete data source: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("data source not found: %s", sourceID)
	}

	return nil
}

// Helper method to scan data sources
func (r *postgresDataSourceRepository) scanDataSources(rows *sql.Rows) ([]*entity.DataSource, error) {
	var sources []*entity.DataSource
	for rows.Next() {
		var source entity.DataSource
		if err := rows.Scan(
			&source.SourceID,
			&source.SourceName,
			&source.SourceType,
			&source.WebsiteURL,
			&source.ReliabilityScore,
			&source.LastVerified,
			&source.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan data source: %w", err)
		}
		sources = append(sources, &source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating data sources: %w", err)
	}

	return sources, nil
}
