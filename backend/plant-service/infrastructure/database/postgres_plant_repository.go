package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
	"twigger-backend/backend/plant-service/pkg/types"

	"github.com/google/uuid"
)

// PostgresPlantRepository implements PlantRepository using PostgreSQL
type PostgresPlantRepository struct {
	db *sql.DB
}

// NewPostgresPlantRepository creates a new PostgreSQL plant repository
func NewPostgresPlantRepository(db *sql.DB) *PostgresPlantRepository {
	return &PostgresPlantRepository{
		db: db,
	}
}

// FindByID retrieves a plant by its ID
func (r *PostgresPlantRepository) FindByID(ctx context.Context, plantID string) (*entity.Plant, error) {
	query := `
		SELECT
			p.plant_id,
			p.species_id,
			p.cultivar_id,
			p.full_botanical_name,
			ps.plant_type,
			ps.species_name,
			pg.genus_name,
			pf.family_name,
			c.cultivar_name,
			p.created_at
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		LEFT JOIN cultivars c ON p.cultivar_id = c.cultivar_id
		WHERE p.plant_id = $1
	`

	plant := &entity.Plant{}
	var cultivarID, cultivarName sql.NullString

	err := r.db.QueryRowContext(ctx, query, plantID).Scan(
		&plant.PlantID,
		&plant.SpeciesID,
		&cultivarID,
		&plant.FullBotanicalName,
		&plant.PlantType,
		&plant.SpeciesName,
		&plant.GenusName,
		&plant.FamilyName,
		&cultivarName,
		&plant.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, entity.ErrPlantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant: %w", err)
	}

	// Handle nullable fields
	if cultivarID.Valid {
		plant.CultivarID = &cultivarID.String
	}
	if cultivarName.Valid {
		plant.CultivarName = &cultivarName.String
	}

	// Load common names
	if err := r.loadCommonNames(ctx, plant); err != nil {
		return nil, fmt.Errorf("failed to load common names: %w", err)
	}

	return plant, nil
}

// FindByIDs retrieves multiple plants by their IDs
func (r *PostgresPlantRepository) FindByIDs(ctx context.Context, plantIDs []string) ([]*entity.Plant, error) {
	if len(plantIDs) == 0 {
		return []*entity.Plant{}, nil
	}

	// Build placeholders: $1, $2, $3, ...
	placeholders := make([]string, len(plantIDs))
	args := make([]interface{}, len(plantIDs))
	for i, id := range plantIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT
			p.plant_id,
			p.species_id,
			p.cultivar_id,
			p.full_botanical_name,
			ps.plant_type,
			ps.species_name,
			pg.genus_name,
			pf.family_name,
			c.cultivar_name,
			p.created_at
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		LEFT JOIN cultivars c ON p.cultivar_id = c.cultivar_id
		WHERE p.plant_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query plants: %w", err)
	}
	defer rows.Close()

	plants := make([]*entity.Plant, 0, len(plantIDs))
	for rows.Next() {
		plant := &entity.Plant{}
		var cultivarID, cultivarName sql.NullString

		err := rows.Scan(
			&plant.PlantID,
			&plant.SpeciesID,
			&cultivarID,
			&plant.FullBotanicalName,
			&plant.PlantType,
			&plant.SpeciesName,
			&plant.GenusName,
			&plant.FamilyName,
			&cultivarName,
			&plant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plant: %w", err)
		}

		if cultivarID.Valid {
			plant.CultivarID = &cultivarID.String
		}
		if cultivarName.Valid {
			plant.CultivarName = &cultivarName.String
		}

		plants = append(plants, plant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plants: %w", err)
	}

	// Load common names for all plants
	for _, plant := range plants {
		if err := r.loadCommonNames(ctx, plant); err != nil {
			return nil, fmt.Errorf("failed to load common names: %w", err)
		}
	}

	return plants, nil
}

// Create inserts a new plant
func (r *PostgresPlantRepository) Create(ctx context.Context, plant *entity.Plant) error {
	if err := plant.Validate(); err != nil {
		return fmt.Errorf("invalid plant: %w", err)
	}

	// Generate UUID if not provided
	if plant.PlantID == "" {
		plant.PlantID = uuid.New().String()
	}

	// Update botanical name
	plant.UpdateBotanicalName()

	query := `
		INSERT INTO plants (plant_id, species_id, cultivar_id, full_botanical_name, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`

	var cultivarID interface{}
	if plant.CultivarID != nil {
		cultivarID = *plant.CultivarID
	}

	_, err := r.db.ExecContext(ctx, query, plant.PlantID, plant.SpeciesID, cultivarID, plant.FullBotanicalName)
	if err != nil {
		return fmt.Errorf("failed to create plant: %w", err)
	}

	return nil
}

// Update modifies an existing plant
func (r *PostgresPlantRepository) Update(ctx context.Context, plant *entity.Plant) error {
	if err := plant.Validate(); err != nil {
		return fmt.Errorf("invalid plant: %w", err)
	}

	// Update botanical name
	plant.UpdateBotanicalName()

	query := `
		UPDATE plants
		SET species_id = $2, cultivar_id = $3, full_botanical_name = $4
		WHERE plant_id = $1
	`

	var cultivarID interface{}
	if plant.CultivarID != nil {
		cultivarID = *plant.CultivarID
	}

	result, err := r.db.ExecContext(ctx, query, plant.PlantID, plant.SpeciesID, cultivarID, plant.FullBotanicalName)
	if err != nil {
		return fmt.Errorf("failed to update plant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrPlantNotFound
	}

	return nil
}

// Delete removes a plant
func (r *PostgresPlantRepository) Delete(ctx context.Context, plantID string) error {
	query := `DELETE FROM plants WHERE plant_id = $1`

	result, err := r.db.ExecContext(ctx, query, plantID)
	if err != nil {
		return fmt.Errorf("failed to delete plant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrPlantNotFound
	}

	return nil
}

// FindByBotanicalName finds a plant by its exact botanical name
func (r *PostgresPlantRepository) FindByBotanicalName(ctx context.Context, botanicalName string) (*entity.Plant, error) {
	query := `
		SELECT
			p.plant_id,
			p.species_id,
			p.cultivar_id,
			p.full_botanical_name,
			ps.plant_type,
			ps.species_name,
			pg.genus_name,
			pf.family_name,
			c.cultivar_name,
			p.created_at
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		LEFT JOIN cultivars c ON p.cultivar_id = c.cultivar_id
		WHERE LOWER(p.full_botanical_name) = LOWER($1)
		LIMIT 1
	`

	plant := &entity.Plant{}
	var cultivarID, cultivarName sql.NullString

	err := r.db.QueryRowContext(ctx, query, botanicalName).Scan(
		&plant.PlantID,
		&plant.SpeciesID,
		&cultivarID,
		&plant.FullBotanicalName,
		&plant.PlantType,
		&plant.SpeciesName,
		&plant.GenusName,
		&plant.FamilyName,
		&cultivarName,
		&plant.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, entity.ErrPlantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant by botanical name: %w", err)
	}

	if cultivarID.Valid {
		plant.CultivarID = &cultivarID.String
	}
	if cultivarName.Valid {
		plant.CultivarName = &cultivarName.String
	}

	if err := r.loadCommonNames(ctx, plant); err != nil {
		return nil, fmt.Errorf("failed to load common names: %w", err)
	}

	return plant, nil
}

// FindByCommonName finds plants by common name (case-insensitive partial match)
func (r *PostgresPlantRepository) FindByCommonName(ctx context.Context, commonName string) ([]*entity.Plant, error) {
	// This would require a separate common_names table in the schema
	// For now, return empty results
	// TODO: Implement after common_names table is added
	return []*entity.Plant{}, nil
}

// Count returns the total number of plants matching the filter
func (r *PostgresPlantRepository) Count(ctx context.Context, filter *repository.SearchFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM plants p`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count plants: %w", err)
	}

	return count, nil
}

// BulkCreate inserts multiple plants in a transaction
func (r *PostgresPlantRepository) BulkCreate(ctx context.Context, plants []*entity.Plant) error {
	if len(plants) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO plants (plant_id, species_id, cultivar_id, full_botanical_name, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, plant := range plants {
		if err := plant.Validate(); err != nil {
			return fmt.Errorf("invalid plant %s: %w", plant.PlantID, err)
		}

		if plant.PlantID == "" {
			plant.PlantID = uuid.New().String()
		}

		plant.UpdateBotanicalName()

		var cultivarID interface{}
		if plant.CultivarID != nil {
			cultivarID = *plant.CultivarID
		}

		_, err = stmt.ExecContext(ctx, plant.PlantID, plant.SpeciesID, cultivarID, plant.FullBotanicalName)
		if err != nil {
			return fmt.Errorf("failed to insert plant %s: %w", plant.PlantID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// loadCommonNames loads common names for a plant
// This is a helper method that would query a separate common_names table
func (r *PostgresPlantRepository) loadCommonNames(ctx context.Context, plant *entity.Plant) error {
	// TODO: Implement after common_names table is added to schema
	// For now, just initialize empty slice
	plant.CommonNames = []string{}
	return nil
}
