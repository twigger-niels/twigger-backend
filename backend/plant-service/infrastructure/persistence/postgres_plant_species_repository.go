package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

// PostgresPlantSpeciesRepository implements PlantSpeciesRepository using PostgreSQL
type PostgresPlantSpeciesRepository struct {
	db *sql.DB
}

// NewPostgresPlantSpeciesRepository creates a new PostgreSQL plant species repository
func NewPostgresPlantSpeciesRepository(db *sql.DB) repository.PlantSpeciesRepository {
	return &PostgresPlantSpeciesRepository{db: db}
}

func (r *PostgresPlantSpeciesRepository) FindByID(ctx context.Context, speciesID string) (*entity.PlantSpecies, error) {
	query := `
		SELECT species_id, genus_id, species_name, plant_type, created_at
		FROM plant_species
		WHERE species_id = $1
	`

	var species entity.PlantSpecies
	err := r.db.QueryRowContext(ctx, query, speciesID).Scan(
		&species.SpeciesID,
		&species.GenusID,
		&species.SpeciesName,
		&species.PlantType,
		&species.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plant species not found: %s", speciesID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant species: %w", err)
	}

	return &species, nil
}

func (r *PostgresPlantSpeciesRepository) FindByGenus(ctx context.Context, genusID string) ([]*entity.PlantSpecies, error) {
	query := `
		SELECT species_id, genus_id, species_name, plant_type, created_at
		FROM plant_species
		WHERE genus_id = $1
		ORDER BY species_name
	`

	rows, err := r.db.QueryContext(ctx, query, genusID)
	if err != nil {
		return nil, fmt.Errorf("failed to query species by genus: %w", err)
	}
	defer rows.Close()

	return r.scanSpecies(rows)
}

func (r *PostgresPlantSpeciesRepository) FindByType(ctx context.Context, plantType string) ([]*entity.PlantSpecies, error) {
	query := `
		SELECT species_id, genus_id, species_name, plant_type, created_at
		FROM plant_species
		WHERE plant_type = $1
		ORDER BY species_name
	`

	rows, err := r.db.QueryContext(ctx, query, plantType)
	if err != nil {
		return nil, fmt.Errorf("failed to query species by type: %w", err)
	}
	defer rows.Close()

	return r.scanSpecies(rows)
}

func (r *PostgresPlantSpeciesRepository) FindAll(ctx context.Context) ([]*entity.PlantSpecies, error) {
	query := `
		SELECT species_id, genus_id, species_name, plant_type, created_at
		FROM plant_species
		ORDER BY species_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plant species: %w", err)
	}
	defer rows.Close()

	return r.scanSpecies(rows)
}

func (r *PostgresPlantSpeciesRepository) Search(ctx context.Context, query string, limit int) ([]*entity.PlantSpecies, error) {
	sqlQuery := `
		SELECT species_id, genus_id, species_name, plant_type, created_at
		FROM plant_species
		WHERE species_name ILIKE $1
		ORDER BY species_name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search plant species: %w", err)
	}
	defer rows.Close()

	return r.scanSpecies(rows)
}

func (r *PostgresPlantSpeciesRepository) Create(ctx context.Context, species *entity.PlantSpecies) error {
	if err := species.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO plant_species (species_id, genus_id, species_name, plant_type, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	species.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		species.SpeciesID,
		species.GenusID,
		species.SpeciesName,
		species.PlantType,
		species.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create plant species: %w", err)
	}

	return nil
}

func (r *PostgresPlantSpeciesRepository) Update(ctx context.Context, species *entity.PlantSpecies) error {
	if err := species.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE plant_species
		SET genus_id = $2, species_name = $3, plant_type = $4
		WHERE species_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		species.SpeciesID,
		species.GenusID,
		species.SpeciesName,
		species.PlantType,
	)

	if err != nil {
		return fmt.Errorf("failed to update plant species: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant species not found: %s", species.SpeciesID)
	}

	return nil
}

func (r *PostgresPlantSpeciesRepository) Delete(ctx context.Context, speciesID string) error {
	query := `DELETE FROM plant_species WHERE species_id = $1`

	result, err := r.db.ExecContext(ctx, query, speciesID)
	if err != nil {
		return fmt.Errorf("failed to delete plant species: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant species not found: %s", speciesID)
	}

	return nil
}

// Helper method to scan species
func (r *PostgresPlantSpeciesRepository) scanSpecies(rows *sql.Rows) ([]*entity.PlantSpecies, error) {
	var speciesList []*entity.PlantSpecies
	for rows.Next() {
		var species entity.PlantSpecies
		if err := rows.Scan(
			&species.SpeciesID,
			&species.GenusID,
			&species.SpeciesName,
			&species.PlantType,
			&species.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plant species: %w", err)
		}
		speciesList = append(speciesList, &species)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plant species: %w", err)
	}

	return speciesList, nil
}
