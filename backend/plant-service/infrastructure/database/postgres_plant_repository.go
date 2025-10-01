package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"

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

// FindByID retrieves a plant by its ID with localized common names
func (r *PostgresPlantRepository) FindByID(ctx context.Context, plantID, languageID string, countryID *string) (*entity.Plant, error) {
	// Validate inputs
	if err := ValidatePlantID(plantID); err != nil {
		return nil, fmt.Errorf("invalid plant_id: %w", err)
	}
	if err := ValidateLanguageID(languageID); err != nil {
		return nil, fmt.Errorf("invalid language_id: %w", err)
	}
	if err := ValidateCountryID(countryID); err != nil {
		return nil, fmt.Errorf("invalid country_id: %w", err)
	}

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

	// Load localized common names
	if err := r.loadCommonNames(ctx, plant, languageID, countryID); err != nil {
		return nil, fmt.Errorf("failed to load common names: %w", err)
	}

	return plant, nil
}

// FindByIDs retrieves multiple plants by their IDs with localized common names
func (r *PostgresPlantRepository) FindByIDs(ctx context.Context, plantIDs []string, languageID string, countryID *string) ([]*entity.Plant, error) {
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

	// Load localized common names for all plants (batch loading to avoid N+1 query)
	if err := r.loadCommonNamesForMultiplePlants(ctx, plants, languageID, countryID); err != nil {
		return nil, fmt.Errorf("failed to load common names: %w", err)
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

// FindByBotanicalName finds a plant by its exact botanical name with localized common names
func (r *PostgresPlantRepository) FindByBotanicalName(ctx context.Context, botanicalName, languageID string, countryID *string) (*entity.Plant, error) {
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

	if err := r.loadCommonNames(ctx, plant, languageID, countryID); err != nil {
		return nil, fmt.Errorf("failed to load common names: %w", err)
	}

	return plant, nil
}

// FindByCommonName finds plants by common name with language context and fallback
func (r *PostgresPlantRepository) FindByCommonName(ctx context.Context, commonName, languageID string, countryID *string) ([]*entity.Plant, error) {
	// Validate inputs
	if err := ValidateLanguageID(languageID); err != nil {
		return nil, fmt.Errorf("invalid language_id: %w", err)
	}
	if err := ValidateCountryID(countryID); err != nil {
		return nil, fmt.Errorf("invalid country_id: %w", err)
	}
	if commonName == "" {
		return nil, fmt.Errorf("common_name is required")
	}

	// Search in requested language/country first
	query := `
		SELECT DISTINCT p.plant_id
		FROM plants p
		INNER JOIN plant_common_names pcn ON p.plant_id = pcn.plant_id
		WHERE pcn.language_id = $1
		  AND pcn.common_name ILIKE $2
		  AND (
			  pcn.country_id = $3 OR
			  ($3 IS NOT NULL AND pcn.country_id IS NULL) OR
			  $3 IS NULL
		  )
		ORDER BY p.plant_id
		LIMIT 100
	`

	var countryIDParam interface{}
	if countryID != nil {
		countryIDParam = *countryID
	}

	rows, err := r.db.QueryContext(ctx, query, languageID, "%"+commonName+"%", countryIDParam)
	if err != nil {
		return nil, fmt.Errorf("failed to query plants by common name: %w", err)
	}
	defer rows.Close()

	plantIDs := make([]string, 0)
	for rows.Next() {
		var plantID string
		if err := rows.Scan(&plantID); err != nil {
			return nil, fmt.Errorf("failed to scan plant ID: %w", err)
		}
		plantIDs = append(plantIDs, plantID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plant IDs: %w", err)
	}

	// If no results and not English, try English fallback
	if len(plantIDs) == 0 && languageID != "en" {
		return r.findByCommonNameEnglishFallback(ctx, commonName, languageID, countryID)
	}

	// Load full plant details
	if len(plantIDs) == 0 {
		return []*entity.Plant{}, nil
	}

	return r.FindByIDs(ctx, plantIDs, languageID, countryID)
}

// findByCommonNameEnglishFallback searches in English when no results in requested language
func (r *PostgresPlantRepository) findByCommonNameEnglishFallback(ctx context.Context, commonName, originalLanguageID string, countryID *string) ([]*entity.Plant, error) {
	query := `
		SELECT DISTINCT p.plant_id
		FROM plants p
		INNER JOIN plant_common_names pcn ON p.plant_id = pcn.plant_id
		INNER JOIN languages l ON pcn.language_id = l.language_id
		WHERE l.language_code = 'en'
		  AND pcn.common_name ILIKE $1
		  AND pcn.country_id IS NULL
		ORDER BY p.plant_id
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, "%"+commonName+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query English fallback: %w", err)
	}
	defer rows.Close()

	plantIDs := make([]string, 0)
	for rows.Next() {
		var plantID string
		if err := rows.Scan(&plantID); err != nil {
			return nil, fmt.Errorf("failed to scan plant ID: %w", err)
		}
		plantIDs = append(plantIDs, plantID)
	}

	if len(plantIDs) == 0 {
		return []*entity.Plant{}, nil
	}

	// Return with original language context for common names
	return r.FindByIDs(ctx, plantIDs, originalLanguageID, countryID)
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

// loadCommonNames loads common names for a plant with language context and fallback
func (r *PostgresPlantRepository) loadCommonNames(ctx context.Context, plant *entity.Plant, languageID string, countryID *string) error {
	query := `
		SELECT common_name, is_primary, country_id IS NOT NULL AS is_country_specific
		FROM plant_common_names
		WHERE plant_id = $1
		  AND language_id = $2
		  AND (
			  country_id = $3 OR
			  ($3 IS NOT NULL AND country_id IS NULL) OR
			  $3 IS NULL
		  )
		ORDER BY
			country_id IS NOT NULL DESC, -- Country-specific first
			is_primary DESC,
			common_name
	`

	var countryIDParam interface{}
	if countryID != nil {
		countryIDParam = *countryID
	}

	rows, err := r.db.QueryContext(ctx, query, plant.PlantID, languageID, countryIDParam)
	if err != nil {
		return fmt.Errorf("failed to query common names: %w", err)
	}
	defer rows.Close()

	commonNames := make([]string, 0)
	for rows.Next() {
		var name string
		var isPrimary bool
		var isCountrySpecific bool

		if err := rows.Scan(&name, &isPrimary, &isCountrySpecific); err != nil {
			return fmt.Errorf("failed to scan common name: %w", err)
		}

		commonNames = append(commonNames, name)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating common names: %w", err)
	}

	// Fallback to English if no names found in requested language
	if len(commonNames) == 0 && languageID != "en" {
		return r.loadCommonNamesEnglishFallback(ctx, plant)
	}

	plant.CommonNames = commonNames
	return nil
}

// loadCommonNamesForMultiplePlants loads common names for multiple plants in a single query (batch loading)
// This prevents N+1 query problems when loading lists of plants
func (r *PostgresPlantRepository) loadCommonNamesForMultiplePlants(ctx context.Context, plants []*entity.Plant, languageID string, countryID *string) error {
	if len(plants) == 0 {
		return nil
	}

	// Collect all plant IDs
	plantIDs := make([]string, len(plants))
	plantMap := make(map[string]*entity.Plant)
	for i, p := range plants {
		plantIDs[i] = p.PlantID
		plantMap[p.PlantID] = p
		// Initialize empty slices
		p.CommonNames = []string{}
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(plantIDs))
	args := make([]interface{}, len(plantIDs)+2)
	for i, id := range plantIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(plantIDs)] = languageID

	var countryIDParam interface{}
	if countryID != nil {
		countryIDParam = *countryID
	}
	args[len(plantIDs)+1] = countryIDParam

	query := fmt.Sprintf(`
		SELECT plant_id, common_name, is_primary, country_id IS NOT NULL AS is_country_specific
		FROM plant_common_names
		WHERE plant_id IN (%s)
		  AND language_id = $%d
		  AND (
			  country_id = $%d OR
			  ($%d IS NOT NULL AND country_id IS NULL) OR
			  $%d IS NULL
		  )
		ORDER BY
			plant_id,
			country_id IS NOT NULL DESC,
			is_primary DESC,
			common_name
	`, strings.Join(placeholders, ","), len(plantIDs)+1, len(plantIDs)+2, len(plantIDs)+2, len(plantIDs)+2)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch query common names: %w", err)
	}
	defer rows.Close()

	// Group results by plant_id
	namesByPlant := make(map[string][]string)
	for rows.Next() {
		var plantID, name string
		var isPrimary, isCountrySpecific bool

		if err := rows.Scan(&plantID, &name, &isPrimary, &isCountrySpecific); err != nil {
			return fmt.Errorf("failed to scan common name: %w", err)
		}

		namesByPlant[plantID] = append(namesByPlant[plantID], name)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating common names: %w", err)
	}

	// Assign names to plants
	plantsNeedingFallback := make([]*entity.Plant, 0)
	for _, plant := range plants {
		if names, ok := namesByPlant[plant.PlantID]; ok && len(names) > 0 {
			plant.CommonNames = names
		} else if languageID != "en" {
			// Track plants that need English fallback
			plantsNeedingFallback = append(plantsNeedingFallback, plant)
		}
	}

	// Handle English fallback for plants with no names
	if len(plantsNeedingFallback) > 0 {
		return r.loadEnglishFallbackForMultiplePlants(ctx, plantsNeedingFallback)
	}

	return nil
}

// loadEnglishFallbackForMultiplePlants loads English common names for multiple plants as fallback
func (r *PostgresPlantRepository) loadEnglishFallbackForMultiplePlants(ctx context.Context, plants []*entity.Plant) error {
	if len(plants) == 0 {
		return nil
	}

	// Collect plant IDs
	plantIDs := make([]string, len(plants))
	plantMap := make(map[string]*entity.Plant)
	for i, p := range plants {
		plantIDs[i] = p.PlantID
		plantMap[p.PlantID] = p
	}

	// Build placeholders
	placeholders := make([]string, len(plantIDs))
	args := make([]interface{}, len(plantIDs))
	for i, id := range plantIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT pcn.plant_id, pcn.common_name
		FROM plant_common_names pcn
		INNER JOIN languages l ON pcn.language_id = l.language_id
		WHERE pcn.plant_id IN (%s)
		  AND l.language_code = 'en'
		  AND pcn.country_id IS NULL
		ORDER BY pcn.plant_id, pcn.is_primary DESC, pcn.common_name
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query English fallback: %w", err)
	}
	defer rows.Close()

	// Group by plant_id
	namesByPlant := make(map[string][]string)
	for rows.Next() {
		var plantID, name string
		if err := rows.Scan(&plantID, &name); err != nil {
			return fmt.Errorf("failed to scan fallback name: %w", err)
		}
		namesByPlant[plantID] = append(namesByPlant[plantID], name)
	}

	// Assign to plants
	for _, plant := range plants {
		if names, ok := namesByPlant[plant.PlantID]; ok {
			plant.CommonNames = names
		}
	}

	return nil
}

// loadCommonNamesEnglishFallback loads English common names as fallback
func (r *PostgresPlantRepository) loadCommonNamesEnglishFallback(ctx context.Context, plant *entity.Plant) error {
	query := `
		SELECT common_name
		FROM plant_common_names pcn
		INNER JOIN languages l ON pcn.language_id = l.language_id
		WHERE pcn.plant_id = $1
		  AND l.language_code = 'en'
		  AND pcn.country_id IS NULL
		ORDER BY pcn.is_primary DESC, pcn.common_name
	`

	rows, err := r.db.QueryContext(ctx, query, plant.PlantID)
	if err != nil {
		return fmt.Errorf("failed to query English fallback names: %w", err)
	}
	defer rows.Close()

	commonNames := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan fallback name: %w", err)
		}
		commonNames = append(commonNames, name)
	}

	plant.CommonNames = commonNames
	return nil
}

