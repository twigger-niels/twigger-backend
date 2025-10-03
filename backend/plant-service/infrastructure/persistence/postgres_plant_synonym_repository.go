package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

// PostgresPlantSynonymRepository implements PlantSynonymRepository using PostgreSQL
type PostgresPlantSynonymRepository struct {
	db *sql.DB
}

// NewPostgresPlantSynonymRepository creates a new PostgreSQL plant synonym repository
func NewPostgresPlantSynonymRepository(db *sql.DB) repository.PlantSynonymRepository {
	return &PostgresPlantSynonymRepository{db: db}
}

func (r *PostgresPlantSynonymRepository) FindByID(ctx context.Context, synonymID string) (*entity.PlantSynonym, error) {
	query := `
		SELECT synonym_id, current_plant_id, old_name, date_deprecated, created_at
		FROM plant_synonyms
		WHERE synonym_id = $1
	`

	var synonym entity.PlantSynonym
	err := r.db.QueryRowContext(ctx, query, synonymID).Scan(
		&synonym.SynonymID,
		&synonym.CurrentPlantID,
		&synonym.OldName,
		&synonym.DateDeprecated,
		&synonym.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plant synonym not found: %s", synonymID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant synonym: %w", err)
	}

	return &synonym, nil
}

func (r *PostgresPlantSynonymRepository) FindByCurrentPlant(ctx context.Context, currentPlantID string) ([]*entity.PlantSynonym, error) {
	query := `
		SELECT synonym_id, current_plant_id, old_name, date_deprecated, created_at
		FROM plant_synonyms
		WHERE current_plant_id = $1
		ORDER BY old_name
	`

	rows, err := r.db.QueryContext(ctx, query, currentPlantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query synonyms by current plant: %w", err)
	}
	defer rows.Close()

	return r.scanSynonyms(rows)
}

func (r *PostgresPlantSynonymRepository) FindByOldName(ctx context.Context, oldName string) ([]*entity.PlantSynonym, error) {
	query := `
		SELECT synonym_id, current_plant_id, old_name, date_deprecated, created_at
		FROM plant_synonyms
		WHERE old_name ILIKE $1
		ORDER BY old_name
	`

	rows, err := r.db.QueryContext(ctx, query, "%"+oldName+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query synonyms by old name: %w", err)
	}
	defer rows.Close()

	return r.scanSynonyms(rows)
}

func (r *PostgresPlantSynonymRepository) Create(ctx context.Context, synonym *entity.PlantSynonym) error {
	if err := synonym.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO plant_synonyms (synonym_id, current_plant_id, old_name, date_deprecated, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	synonym.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		synonym.SynonymID,
		synonym.CurrentPlantID,
		synonym.OldName,
		synonym.DateDeprecated,
		synonym.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create plant synonym: %w", err)
	}

	return nil
}

func (r *PostgresPlantSynonymRepository) Update(ctx context.Context, synonym *entity.PlantSynonym) error {
	if err := synonym.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE plant_synonyms
		SET current_plant_id = $2, old_name = $3, date_deprecated = $4
		WHERE synonym_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		synonym.SynonymID,
		synonym.CurrentPlantID,
		synonym.OldName,
		synonym.DateDeprecated,
	)

	if err != nil {
		return fmt.Errorf("failed to update plant synonym: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant synonym not found: %s", synonym.SynonymID)
	}

	return nil
}

func (r *PostgresPlantSynonymRepository) Delete(ctx context.Context, synonymID string) error {
	query := `DELETE FROM plant_synonyms WHERE synonym_id = $1`

	result, err := r.db.ExecContext(ctx, query, synonymID)
	if err != nil {
		return fmt.Errorf("failed to delete plant synonym: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant synonym not found: %s", synonymID)
	}

	return nil
}

// Helper method to scan synonyms
func (r *PostgresPlantSynonymRepository) scanSynonyms(rows *sql.Rows) ([]*entity.PlantSynonym, error) {
	var synonyms []*entity.PlantSynonym
	for rows.Next() {
		var synonym entity.PlantSynonym
		if err := rows.Scan(
			&synonym.SynonymID,
			&synonym.CurrentPlantID,
			&synonym.OldName,
			&synonym.DateDeprecated,
			&synonym.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plant synonym: %w", err)
		}
		synonyms = append(synonyms, &synonym)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plant synonyms: %w", err)
	}

	return synonyms, nil
}
