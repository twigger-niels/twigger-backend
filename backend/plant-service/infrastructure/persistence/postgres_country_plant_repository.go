package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

type postgresCountryPlantRepository struct {
	db *sql.DB
}

// NewPostgresCountryPlantRepository creates a new PostgreSQL country-plant repository
func NewPostgresCountryPlantRepository(db *sql.DB) repository.CountryPlantRepository {
	return &postgresCountryPlantRepository{db: db}
}

func (r *postgresCountryPlantRepository) FindByID(ctx context.Context, countryPlantID string) (*entity.CountryPlant, error) {
	query := `
		SELECT country_plant_id, country_id, plant_id, native_status, legal_status,
		       ST_AsGeoJSON(native_range_geojson), created_at, updated_at
		FROM country_plants
		WHERE country_plant_id = $1
	`

	var cp entity.CountryPlant
	err := r.db.QueryRowContext(ctx, query, countryPlantID).Scan(
		&cp.CountryPlantID,
		&cp.CountryID,
		&cp.PlantID,
		&cp.NativeStatus,
		&cp.LegalStatus,
		&cp.NativeRangeGeoJSON,
		&cp.CreatedAt,
		&cp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country-plant relationship not found: %s", countryPlantID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find country-plant relationship: %w", err)
	}

	return &cp, nil
}

func (r *postgresCountryPlantRepository) FindByCountry(ctx context.Context, countryID string) ([]*entity.CountryPlant, error) {
	query := `
		SELECT country_plant_id, country_id, plant_id, native_status, legal_status,
		       ST_AsGeoJSON(native_range_geojson), created_at, updated_at
		FROM country_plants
		WHERE country_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query country-plants by country: %w", err)
	}
	defer rows.Close()

	return r.scanCountryPlants(rows)
}

func (r *postgresCountryPlantRepository) FindByPlant(ctx context.Context, plantID string) ([]*entity.CountryPlant, error) {
	query := `
		SELECT country_plant_id, country_id, plant_id, native_status, legal_status,
		       ST_AsGeoJSON(native_range_geojson), created_at, updated_at
		FROM country_plants
		WHERE plant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, plantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query country-plants by plant: %w", err)
	}
	defer rows.Close()

	return r.scanCountryPlants(rows)
}

func (r *postgresCountryPlantRepository) FindByCountryAndPlant(ctx context.Context, countryID, plantID string) (*entity.CountryPlant, error) {
	query := `
		SELECT country_plant_id, country_id, plant_id, native_status, legal_status,
		       ST_AsGeoJSON(native_range_geojson), created_at, updated_at
		FROM country_plants
		WHERE country_id = $1 AND plant_id = $2
	`

	var cp entity.CountryPlant
	err := r.db.QueryRowContext(ctx, query, countryID, plantID).Scan(
		&cp.CountryPlantID,
		&cp.CountryID,
		&cp.PlantID,
		&cp.NativeStatus,
		&cp.LegalStatus,
		&cp.NativeRangeGeoJSON,
		&cp.CreatedAt,
		&cp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country-plant relationship not found for country %s and plant %s", countryID, plantID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find country-plant relationship: %w", err)
	}

	return &cp, nil
}

func (r *postgresCountryPlantRepository) FindByNativeStatus(ctx context.Context, countryID, nativeStatus string) ([]*entity.CountryPlant, error) {
	query := `
		SELECT country_plant_id, country_id, plant_id, native_status, legal_status,
		       ST_AsGeoJSON(native_range_geojson), created_at, updated_at
		FROM country_plants
		WHERE country_id = $1 AND native_status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, countryID, nativeStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to query country-plants by native status: %w", err)
	}
	defer rows.Close()

	return r.scanCountryPlants(rows)
}

func (r *postgresCountryPlantRepository) FindByLegalStatus(ctx context.Context, countryID, legalStatus string) ([]*entity.CountryPlant, error) {
	query := `
		SELECT country_plant_id, country_id, plant_id, native_status, legal_status,
		       ST_AsGeoJSON(native_range_geojson), created_at, updated_at
		FROM country_plants
		WHERE country_id = $1 AND legal_status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, countryID, legalStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to query country-plants by legal status: %w", err)
	}
	defer rows.Close()

	return r.scanCountryPlants(rows)
}

func (r *postgresCountryPlantRepository) Create(ctx context.Context, countryPlant *entity.CountryPlant) error {
	if err := countryPlant.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO country_plants (country_plant_id, country_id, plant_id, native_status,
		                            legal_status, native_range_geojson, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, ST_GeomFromGeoJSON($6), $7, $8)
	`

	now := time.Now()
	countryPlant.CreatedAt = now
	countryPlant.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		countryPlant.CountryPlantID,
		countryPlant.CountryID,
		countryPlant.PlantID,
		countryPlant.NativeStatus,
		countryPlant.LegalStatus,
		countryPlant.NativeRangeGeoJSON,
		countryPlant.CreatedAt,
		countryPlant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create country-plant relationship: %w", err)
	}

	return nil
}

func (r *postgresCountryPlantRepository) Update(ctx context.Context, countryPlant *entity.CountryPlant) error {
	if err := countryPlant.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE country_plants
		SET country_id = $2, plant_id = $3, native_status = $4,
		    legal_status = $5, native_range_geojson = ST_GeomFromGeoJSON($6), updated_at = $7
		WHERE country_plant_id = $1
	`

	countryPlant.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		countryPlant.CountryPlantID,
		countryPlant.CountryID,
		countryPlant.PlantID,
		countryPlant.NativeStatus,
		countryPlant.LegalStatus,
		countryPlant.NativeRangeGeoJSON,
		countryPlant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update country-plant relationship: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("country-plant relationship not found: %s", countryPlant.CountryPlantID)
	}

	return nil
}

func (r *postgresCountryPlantRepository) Delete(ctx context.Context, countryPlantID string) error {
	query := `DELETE FROM country_plants WHERE country_plant_id = $1`

	result, err := r.db.ExecContext(ctx, query, countryPlantID)
	if err != nil {
		return fmt.Errorf("failed to delete country-plant relationship: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("country-plant relationship not found: %s", countryPlantID)
	}

	return nil
}

// Helper method to scan country-plants
func (r *postgresCountryPlantRepository) scanCountryPlants(rows *sql.Rows) ([]*entity.CountryPlant, error) {
	var countryPlants []*entity.CountryPlant
	for rows.Next() {
		var cp entity.CountryPlant
		if err := rows.Scan(
			&cp.CountryPlantID,
			&cp.CountryID,
			&cp.PlantID,
			&cp.NativeStatus,
			&cp.LegalStatus,
			&cp.NativeRangeGeoJSON,
			&cp.CreatedAt,
			&cp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan country-plant relationship: %w", err)
		}
		countryPlants = append(countryPlants, &cp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating country-plant relationships: %w", err)
	}

	return countryPlants, nil
}
