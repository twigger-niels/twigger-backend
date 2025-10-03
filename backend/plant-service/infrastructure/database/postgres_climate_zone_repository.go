package database

import (
	"context"
	"database/sql"
	"fmt"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

// PostgresClimateZoneRepository implements ClimateZoneRepository using PostgreSQL
type PostgresClimateZoneRepository struct {
	db *sql.DB
}

// NewPostgresClimateZoneRepository creates a new PostgreSQL climate zone repository
func NewPostgresClimateZoneRepository(db *sql.DB) repository.ClimateZoneRepository {
	return &PostgresClimateZoneRepository{db: db}
}

// FindByID retrieves a climate zone by its UUID
func (r *PostgresClimateZoneRepository) FindByID(ctx context.Context, zoneID string) (*entity.ClimateZone, error) {
	query := `
		SELECT
			zone_id,
			country_id,
			zone_system,
			zone_code,
			ST_AsGeoJSON(zone_geometry) as zone_geometry,
			min_temp_c,
			max_temp_c,
			created_at
		FROM climate_zones
		WHERE zone_id = $1
	`

	var zone entity.ClimateZone
	var geometryJSON sql.NullString
	var minTempC, maxTempC sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, zoneID).Scan(
		&zone.ZoneID,
		&zone.CountryID,
		&zone.ZoneSystem,
		&zone.ZoneCode,
		&geometryJSON,
		&minTempC,
		&maxTempC,
		&zone.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("climate zone not found: %s", zoneID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find climate zone: %w", err)
	}

	if geometryJSON.Valid {
		zone.ZoneGeometryJSON = &geometryJSON.String
	}
	if minTempC.Valid {
		zone.MinTempC = &minTempC.Float64
	}
	if maxTempC.Valid {
		zone.MaxTempC = &maxTempC.Float64
	}

	return &zone, nil
}

// FindByCountry retrieves climate zones for a specific country with pagination
func (r *PostgresClimateZoneRepository) FindByCountry(ctx context.Context, countryID string, limit, offset int) ([]*entity.ClimateZone, error) {
	// Apply default limit if not specified or invalid
	if limit <= 0 || limit > 1000 {
		limit = 100 // Default page size
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			zone_id,
			country_id,
			zone_system,
			zone_code,
			ST_AsGeoJSON(zone_geometry) as zone_geometry,
			min_temp_c,
			max_temp_c,
			created_at
		FROM climate_zones
		WHERE country_id = $1
		ORDER BY zone_system, zone_code
		LIMIT $2 OFFSET $3
	`

	return r.queryZones(ctx, query, countryID, limit, offset)
}

// FindByCountryAndSystem retrieves zones for a country and climate system
func (r *PostgresClimateZoneRepository) FindByCountryAndSystem(ctx context.Context, countryID, zoneSystem string) ([]*entity.ClimateZone, error) {
	query := `
		SELECT
			zone_id,
			country_id,
			zone_system,
			zone_code,
			ST_AsGeoJSON(zone_geometry) as zone_geometry,
			min_temp_c,
			max_temp_c,
			created_at
		FROM climate_zones
		WHERE country_id = $1 AND zone_system = $2
		ORDER BY zone_code
	`

	return r.queryZones(ctx, query, countryID, zoneSystem)
}

// FindByCode retrieves a zone by country, system, and code
func (r *PostgresClimateZoneRepository) FindByCode(ctx context.Context, countryID, zoneSystem, zoneCode string) (*entity.ClimateZone, error) {
	query := `
		SELECT
			zone_id,
			country_id,
			zone_system,
			zone_code,
			ST_AsGeoJSON(zone_geometry) as zone_geometry,
			min_temp_c,
			max_temp_c,
			created_at
		FROM climate_zones
		WHERE country_id = $1 AND zone_system = $2 AND zone_code = $3
	`

	var zone entity.ClimateZone
	var geometryJSON sql.NullString
	var minTempC, maxTempC sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, countryID, zoneSystem, zoneCode).Scan(
		&zone.ZoneID,
		&zone.CountryID,
		&zone.ZoneSystem,
		&zone.ZoneCode,
		&geometryJSON,
		&minTempC,
		&maxTempC,
		&zone.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("climate zone not found: %s/%s/%s", countryID, zoneSystem, zoneCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find climate zone: %w", err)
	}

	if geometryJSON.Valid {
		zone.ZoneGeometryJSON = &geometryJSON.String
	}
	if minTempC.Valid {
		zone.MinTempC = &minTempC.Float64
	}
	if maxTempC.Valid {
		zone.MaxTempC = &maxTempC.Float64
	}

	return &zone, nil
}

// FindByPoint retrieves the climate zone containing a specific geographic point
// Requires GIST index: idx_climate_zones_geometry USING GIST (zone_geometry)
func (r *PostgresClimateZoneRepository) FindByPoint(ctx context.Context, latitude, longitude float64, zoneSystem string) (*entity.ClimateZone, error) {
	// Validate coordinate bounds
	if err := ValidateCoordinates(latitude, longitude); err != nil {
		return nil, fmt.Errorf("invalid coordinates: %w", err)
	}

	query := `
		SELECT
			zone_id,
			country_id,
			zone_system,
			zone_code,
			ST_AsGeoJSON(zone_geometry) as zone_geometry,
			min_temp_c,
			max_temp_c,
			created_at
		FROM climate_zones
		WHERE ST_Contains(zone_geometry, ST_SetSRID(ST_MakePoint($1, $2), 4326))
			AND zone_system = $3
		LIMIT 1
	`

	var zone entity.ClimateZone
	var geometryJSON sql.NullString
	var minTempC, maxTempC sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, longitude, latitude, zoneSystem).Scan(
		&zone.ZoneID,
		&zone.CountryID,
		&zone.ZoneSystem,
		&zone.ZoneCode,
		&geometryJSON,
		&minTempC,
		&maxTempC,
		&zone.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no climate zone found at location: lat=%f, lng=%f, system=%s", latitude, longitude, zoneSystem)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find climate zone by point: %w", err)
	}

	if geometryJSON.Valid {
		zone.ZoneGeometryJSON = &geometryJSON.String
	}
	if minTempC.Valid {
		zone.MinTempC = &minTempC.Float64
	}
	if maxTempC.Valid {
		zone.MaxTempC = &maxTempC.Float64
	}

	return &zone, nil
}

// Create creates a new climate zone
func (r *PostgresClimateZoneRepository) Create(ctx context.Context, zone *entity.ClimateZone) error {
	if err := zone.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO climate_zones (
			country_id,
			zone_system,
			zone_code,
			zone_geometry,
			min_temp_c,
			max_temp_c
		) VALUES ($1, $2, $3, ST_GeomFromGeoJSON($4), $5, $6)
		RETURNING zone_id, created_at
	`

	var geometryJSON interface{}
	if zone.ZoneGeometryJSON != nil {
		// Validate GeoJSON before passing to database
		if err := ValidateGeoJSON(*zone.ZoneGeometryJSON); err != nil {
			return fmt.Errorf("invalid zone geometry geojson: %w", err)
		}
		geometryJSON = *zone.ZoneGeometryJSON
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		zone.CountryID,
		zone.ZoneSystem,
		zone.ZoneCode,
		geometryJSON,
		zone.MinTempC,
		zone.MaxTempC,
	).Scan(&zone.ZoneID, &zone.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create climate zone: %w", err)
	}

	return nil
}

// Update updates an existing climate zone
func (r *PostgresClimateZoneRepository) Update(ctx context.Context, zone *entity.ClimateZone) error {
	if err := zone.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE climate_zones SET
			country_id = $1,
			zone_system = $2,
			zone_code = $3,
			zone_geometry = ST_GeomFromGeoJSON($4),
			min_temp_c = $5,
			max_temp_c = $6
		WHERE zone_id = $7
	`

	var geometryJSON interface{}
	if zone.ZoneGeometryJSON != nil {
		// Validate GeoJSON before passing to database
		if err := ValidateGeoJSON(*zone.ZoneGeometryJSON); err != nil {
			return fmt.Errorf("invalid zone geometry geojson: %w", err)
		}
		geometryJSON = *zone.ZoneGeometryJSON
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		zone.CountryID,
		zone.ZoneSystem,
		zone.ZoneCode,
		geometryJSON,
		zone.MinTempC,
		zone.MaxTempC,
		zone.ZoneID,
	)

	if err != nil {
		return fmt.Errorf("failed to update climate zone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("climate zone not found: %s", zone.ZoneID)
	}

	return nil
}

// Delete deletes a climate zone by ID
func (r *PostgresClimateZoneRepository) Delete(ctx context.Context, zoneID string) error {
	query := `DELETE FROM climate_zones WHERE zone_id = $1`

	result, err := r.db.ExecContext(ctx, query, zoneID)
	if err != nil {
		return fmt.Errorf("failed to delete climate zone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("climate zone not found: %s", zoneID)
	}

	return nil
}

// Helper method to query multiple zones
func (r *PostgresClimateZoneRepository) queryZones(ctx context.Context, query string, args ...interface{}) ([]*entity.ClimateZone, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query climate zones: %w", err)
	}
	defer rows.Close()

	var zones []*entity.ClimateZone
	for rows.Next() {
		var zone entity.ClimateZone
		var geometryJSON sql.NullString
		var minTempC, maxTempC sql.NullFloat64

		err := rows.Scan(
			&zone.ZoneID,
			&zone.CountryID,
			&zone.ZoneSystem,
			&zone.ZoneCode,
			&geometryJSON,
			&minTempC,
			&maxTempC,
			&zone.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan climate zone: %w", err)
		}

		if geometryJSON.Valid {
			zone.ZoneGeometryJSON = &geometryJSON.String
		}
		if minTempC.Valid {
			zone.MinTempC = &minTempC.Float64
		}
		if maxTempC.Valid {
			zone.MaxTempC = &maxTempC.Float64
		}

		zones = append(zones, &zone)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating climate zones: %w", err)
	}

	return zones, nil
}
