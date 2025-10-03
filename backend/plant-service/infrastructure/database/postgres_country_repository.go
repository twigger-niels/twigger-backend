package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"

	"github.com/lib/pq"
)

// PostgresCountryRepository implements CountryRepository using PostgreSQL
type PostgresCountryRepository struct {
	db *sql.DB
}

// NewPostgresCountryRepository creates a new PostgreSQL country repository
func NewPostgresCountryRepository(db *sql.DB) repository.CountryRepository {
	return &PostgresCountryRepository{db: db}
}

// FindByID retrieves a country by its UUID
func (r *PostgresCountryRepository) FindByID(ctx context.Context, countryID string) (*entity.Country, error) {
	start := time.Now()
	defer func() {
		LogQueryDuration("FindByID", "Country", time.Since(start), 1)
	}()

	query := `
		SELECT
			country_id,
			country_code,
			country_name,
			climate_systems,
			default_climate_system,
			ST_AsGeoJSON(country_boundary) as country_boundary,
			created_at,
			updated_at
		FROM countries
		WHERE country_id = $1
	`

	var country entity.Country
	var climateSystems pq.StringArray
	var defaultClimateSystem sql.NullString
	var boundaryGeoJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, countryID).Scan(
		&country.CountryID,
		&country.CountryCode,
		&country.CountryName,
		&climateSystems,
		&defaultClimateSystem,
		&boundaryGeoJSON,
		&country.CreatedAt,
		&country.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country not found: %s", countryID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find country: %w", err)
	}

	country.ClimateSystems = []string(climateSystems)
	if defaultClimateSystem.Valid {
		country.DefaultClimateSystem = &defaultClimateSystem.String
	}
	if boundaryGeoJSON.Valid {
		country.CountryBoundaryGeoJSON = &boundaryGeoJSON.String
	}

	return &country, nil
}

// FindByCode retrieves a country by its ISO 3166-1 alpha-2 code
func (r *PostgresCountryRepository) FindByCode(ctx context.Context, countryCode string) (*entity.Country, error) {
	query := `
		SELECT
			country_id,
			country_code,
			country_name,
			climate_systems,
			default_climate_system,
			ST_AsGeoJSON(country_boundary) as country_boundary,
			created_at,
			updated_at
		FROM countries
		WHERE country_code = $1
	`

	var country entity.Country
	var climateSystems pq.StringArray
	var defaultClimateSystem sql.NullString
	var boundaryGeoJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, countryCode).Scan(
		&country.CountryID,
		&country.CountryCode,
		&country.CountryName,
		&climateSystems,
		&defaultClimateSystem,
		&boundaryGeoJSON,
		&country.CreatedAt,
		&country.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country not found: %s", countryCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find country: %w", err)
	}

	country.ClimateSystems = []string(climateSystems)
	if defaultClimateSystem.Valid {
		country.DefaultClimateSystem = &defaultClimateSystem.String
	}
	if boundaryGeoJSON.Valid {
		country.CountryBoundaryGeoJSON = &boundaryGeoJSON.String
	}

	return &country, nil
}

// FindAll retrieves all countries
func (r *PostgresCountryRepository) FindAll(ctx context.Context) ([]*entity.Country, error) {
	query := `
		SELECT
			country_id,
			country_code,
			country_name,
			climate_systems,
			default_climate_system,
			ST_AsGeoJSON(country_boundary) as country_boundary,
			created_at,
			updated_at
		FROM countries
		ORDER BY country_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query countries: %w", err)
	}
	defer rows.Close()

	var countries []*entity.Country
	for rows.Next() {
		var country entity.Country
		var climateSystems pq.StringArray
		var defaultClimateSystem sql.NullString
		var boundaryGeoJSON sql.NullString

		err := rows.Scan(
			&country.CountryID,
			&country.CountryCode,
			&country.CountryName,
			&climateSystems,
			&defaultClimateSystem,
			&boundaryGeoJSON,
			&country.CreatedAt,
			&country.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}

		country.ClimateSystems = []string(climateSystems)
		if defaultClimateSystem.Valid {
			country.DefaultClimateSystem = &defaultClimateSystem.String
		}
		if boundaryGeoJSON.Valid {
			country.CountryBoundaryGeoJSON = &boundaryGeoJSON.String
		}

		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating countries: %w", err)
	}

	return countries, nil
}

// FindByClimateSystem retrieves all countries that support a specific climate system
func (r *PostgresCountryRepository) FindByClimateSystem(ctx context.Context, climateSystem string) ([]*entity.Country, error) {
	query := `
		SELECT
			country_id,
			country_code,
			country_name,
			climate_systems,
			default_climate_system,
			ST_AsGeoJSON(country_boundary) as country_boundary,
			created_at,
			updated_at
		FROM countries
		WHERE $1 = ANY(climate_systems)
		ORDER BY country_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, climateSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to query countries by climate system: %w", err)
	}
	defer rows.Close()

	var countries []*entity.Country
	for rows.Next() {
		var country entity.Country
		var climateSystems pq.StringArray
		var defaultClimateSystem sql.NullString
		var boundaryGeoJSON sql.NullString

		err := rows.Scan(
			&country.CountryID,
			&country.CountryCode,
			&country.CountryName,
			&climateSystems,
			&defaultClimateSystem,
			&boundaryGeoJSON,
			&country.CreatedAt,
			&country.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}

		country.ClimateSystems = []string(climateSystems)
		if defaultClimateSystem.Valid {
			country.DefaultClimateSystem = &defaultClimateSystem.String
		}
		if boundaryGeoJSON.Valid {
			country.CountryBoundaryGeoJSON = &boundaryGeoJSON.String
		}

		countries = append(countries, &country)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating countries: %w", err)
	}

	return countries, nil
}

// FindByPoint retrieves the country containing a specific geographic point
// Requires GIST index: idx_countries_boundary USING GIST (country_boundary)
func (r *PostgresCountryRepository) FindByPoint(ctx context.Context, latitude, longitude float64) (*entity.Country, error) {
	// Validate coordinate bounds
	if err := ValidateCoordinates(latitude, longitude); err != nil {
		return nil, fmt.Errorf("invalid coordinates: %w", err)
	}

	query := `
		SELECT
			country_id,
			country_code,
			country_name,
			climate_systems,
			default_climate_system,
			ST_AsGeoJSON(country_boundary) as country_boundary,
			created_at,
			updated_at
		FROM countries
		WHERE ST_Contains(country_boundary, ST_SetSRID(ST_MakePoint($1, $2), 4326))
		LIMIT 1
	`

	var country entity.Country
	var climateSystems pq.StringArray
	var defaultClimateSystem sql.NullString
	var boundaryGeoJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, longitude, latitude).Scan(
		&country.CountryID,
		&country.CountryCode,
		&country.CountryName,
		&climateSystems,
		&defaultClimateSystem,
		&boundaryGeoJSON,
		&country.CreatedAt,
		&country.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no country found at location: lat=%f, lng=%f", latitude, longitude)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find country by point: %w", err)
	}

	country.ClimateSystems = []string(climateSystems)
	if defaultClimateSystem.Valid {
		country.DefaultClimateSystem = &defaultClimateSystem.String
	}
	if boundaryGeoJSON.Valid {
		country.CountryBoundaryGeoJSON = &boundaryGeoJSON.String
	}

	return &country, nil
}

// Create creates a new country
func (r *PostgresCountryRepository) Create(ctx context.Context, country *entity.Country) error {
	if err := country.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO countries (
			country_code,
			country_name,
			climate_systems,
			default_climate_system,
			country_boundary
		) VALUES ($1, $2, $3, $4, ST_GeomFromGeoJSON($5))
		RETURNING country_id, created_at, updated_at
	`

	var boundaryGeoJSON interface{}
	if country.CountryBoundaryGeoJSON != nil {
		// Validate GeoJSON before passing to database
		if err := ValidateGeoJSON(*country.CountryBoundaryGeoJSON); err != nil {
			return fmt.Errorf("invalid country boundary geojson: %w", err)
		}
		boundaryGeoJSON = *country.CountryBoundaryGeoJSON
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		country.CountryCode,
		country.CountryName,
		country.ClimateSystemsArray(),
		country.DefaultClimateSystem,
		boundaryGeoJSON,
	).Scan(&country.CountryID, &country.CreatedAt, &country.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create country: %w", err)
	}

	return nil
}

// Update updates an existing country
func (r *PostgresCountryRepository) Update(ctx context.Context, country *entity.Country) error {
	if err := country.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE countries SET
			country_code = $1,
			country_name = $2,
			climate_systems = $3,
			default_climate_system = $4,
			country_boundary = ST_GeomFromGeoJSON($5),
			updated_at = CURRENT_TIMESTAMP
		WHERE country_id = $6
		RETURNING updated_at
	`

	var boundaryGeoJSON interface{}
	if country.CountryBoundaryGeoJSON != nil {
		// Validate GeoJSON before passing to database
		if err := ValidateGeoJSON(*country.CountryBoundaryGeoJSON); err != nil {
			return fmt.Errorf("invalid country boundary geojson: %w", err)
		}
		boundaryGeoJSON = *country.CountryBoundaryGeoJSON
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		country.CountryCode,
		country.CountryName,
		country.ClimateSystemsArray(),
		country.DefaultClimateSystem,
		boundaryGeoJSON,
		country.CountryID,
	).Scan(&country.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("country not found: %s", country.CountryID)
	}
	if err != nil {
		return fmt.Errorf("failed to update country: %w", err)
	}

	return nil
}

// Delete deletes a country by ID
func (r *PostgresCountryRepository) Delete(ctx context.Context, countryID string) error {
	query := `DELETE FROM countries WHERE country_id = $1`

	result, err := r.db.ExecContext(ctx, query, countryID)
	if err != nil {
		return fmt.Errorf("failed to delete country: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("country not found: %s", countryID)
	}

	return nil
}
