package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/backend/garden-service/domain/repository"
	"twigger-backend/backend/garden-service/infrastructure/database"
)

// PostgresGardenZoneRepository implements repository.GardenZoneRepository using PostgreSQL + PostGIS
type PostgresGardenZoneRepository struct {
	db *sql.DB
}

// NewPostgresGardenZoneRepository creates a new PostgreSQL garden zone repository
func NewPostgresGardenZoneRepository(db *sql.DB) repository.GardenZoneRepository {
	return &PostgresGardenZoneRepository{db: db}
}

// Create creates a new garden zone
func (r *PostgresGardenZoneRepository) Create(ctx context.Context, zone *entity.GardenZone) error {
	// Validate entity
	if err := zone.Validate(); err != nil {
		return entity.NewValidationError("garden_zone", err.Error())
	}

	// Validate GeoJSON before database insert (Gotcha #32)
	if err := database.ValidateGeoJSON(zone.GeometryGeoJSON); err != nil {
		return entity.NewSpatialError("geometry_validation", err.Error())
	}
	if err := database.ValidatePolygonClosure(zone.GeometryGeoJSON); err != nil {
		return entity.NewSpatialError("geometry_validation", err.Error())
	}

	// Generate ID if not provided
	if zone.ZoneID == "" {
		zone.ZoneID = uuid.New().String()
	}

	// Set timestamp
	zone.CreatedAt = time.Now()

	query := `
		INSERT INTO garden_zones (
			zone_id, garden_id, zone_name, zone_type, geometry,
			soil_type, soil_amended, irrigation_type,
			sun_hours_summer, sun_hours_winter, created_at
		) VALUES (
			$1, $2, $3, $4, ST_GeomFromGeoJSON($5),
			$6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		zone.ZoneID,
		zone.GardenID,
		zone.ZoneName,
		zone.ZoneType,
		zone.GeometryGeoJSON,
		zone.SoilType,
		zone.SoilAmended,
		zone.IrrigationType,
		zone.SunHoursSummer,
		zone.SunHoursWinter,
		zone.CreatedAt,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_zone_create", err)
	}

	return nil
}

// FindByID finds a garden zone by ID
func (r *PostgresGardenZoneRepository) FindByID(ctx context.Context, zoneID string) (*entity.GardenZone, error) {
	query := `
		SELECT
			zone_id, garden_id, zone_name, zone_type,
			ST_AsGeoJSON(geometry) as geometry,
			ST_Area(geometry::geography) as area_m2,
			soil_type, soil_amended, irrigation_type,
			sun_hours_summer, sun_hours_winter, created_at
		FROM garden_zones
		WHERE zone_id = $1
	`

	var zone entity.GardenZone
	var zoneName, zoneType sql.NullString
	var areaM2 sql.NullFloat64
	var soilType, irrigationType sql.NullString
	var sunHoursSummer, sunHoursWinter sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, zoneID).Scan(
		&zone.ZoneID,
		&zone.GardenID,
		&zoneName,
		&zoneType,
		&zone.GeometryGeoJSON,
		&areaM2,
		&soilType,
		&zone.SoilAmended,
		&irrigationType,
		&sunHoursSummer,
		&sunHoursWinter,
		&zone.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, entity.NewNotFoundError("garden_zone", zoneID)
	}
	if err != nil {
		return nil, entity.NewDatabaseError("garden_zone_find_by_id", err)
	}

	// Map nullable fields
	if zoneName.Valid {
		zone.ZoneName = &zoneName.String
	}
	if zoneType.Valid {
		typeValue := entity.ZoneType(zoneType.String)
		zone.ZoneType = &typeValue
	}
	if areaM2.Valid {
		zone.AreaM2 = &areaM2.Float64
	}
	if soilType.Valid {
		zone.SoilType = &soilType.String
	}
	if irrigationType.Valid {
		irrigationValue := entity.IrrigationType(irrigationType.String)
		zone.IrrigationType = &irrigationValue
	}
	if sunHoursSummer.Valid {
		hours := int(sunHoursSummer.Int32)
		zone.SunHoursSummer = &hours
	}
	if sunHoursWinter.Valid {
		hours := int(sunHoursWinter.Int32)
		zone.SunHoursWinter = &hours
	}

	return &zone, nil
}

// FindByGardenID finds all zones for a garden
func (r *PostgresGardenZoneRepository) FindByGardenID(ctx context.Context, gardenID string) ([]*entity.GardenZone, error) {
	query := `
		SELECT
			zone_id, garden_id, zone_name, zone_type,
			ST_AsGeoJSON(geometry) as geometry,
			ST_Area(geometry::geography) as area_m2,
			soil_type, soil_amended, irrigation_type,
			sun_hours_summer, sun_hours_winter, created_at
		FROM garden_zones
		WHERE garden_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gardenID)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_zone_find_by_garden_id", err)
	}
	defer rows.Close()

	var zones []*entity.GardenZone
	for rows.Next() {
		var zone entity.GardenZone
		var zoneName, zoneType sql.NullString
		var areaM2 sql.NullFloat64
		var soilType, irrigationType sql.NullString
		var sunHoursSummer, sunHoursWinter sql.NullInt32

		err := rows.Scan(
			&zone.ZoneID,
			&zone.GardenID,
			&zoneName,
			&zoneType,
			&zone.GeometryGeoJSON,
			&areaM2,
			&soilType,
			&zone.SoilAmended,
			&irrigationType,
			&sunHoursSummer,
			&sunHoursWinter,
			&zone.CreatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_zone_scan", err)
		}

		// Map nullable fields
		if zoneName.Valid {
			zone.ZoneName = &zoneName.String
		}
		if zoneType.Valid {
			typeValue := entity.ZoneType(zoneType.String)
			zone.ZoneType = &typeValue
		}
		if areaM2.Valid {
			zone.AreaM2 = &areaM2.Float64
		}
		if soilType.Valid {
			zone.SoilType = &soilType.String
		}
		if irrigationType.Valid {
			irrigationValue := entity.IrrigationType(irrigationType.String)
			zone.IrrigationType = &irrigationValue
		}
		if sunHoursSummer.Valid {
			hours := int(sunHoursSummer.Int32)
			zone.SunHoursSummer = &hours
		}
		if sunHoursWinter.Valid {
			hours := int(sunHoursWinter.Int32)
			zone.SunHoursWinter = &hours
		}

		zones = append(zones, &zone)
	}

	if err = rows.Err(); err != nil {
		return nil, entity.NewDatabaseError("garden_zone_rows_iteration", err)
	}

	return zones, nil
}

// Update updates a garden zone
func (r *PostgresGardenZoneRepository) Update(ctx context.Context, zone *entity.GardenZone) error {
	// Validate entity
	if err := zone.Validate(); err != nil {
		return entity.NewValidationError("garden_zone", err.Error())
	}

	// Validate GeoJSON (Gotcha #32)
	if err := database.ValidateGeoJSON(zone.GeometryGeoJSON); err != nil {
		return entity.NewSpatialError("geometry_validation", err.Error())
	}

	query := `
		UPDATE garden_zones SET
			zone_name = $2,
			zone_type = $3,
			geometry = ST_GeomFromGeoJSON($4),
			soil_type = $5,
			soil_amended = $6,
			irrigation_type = $7,
			sun_hours_summer = $8,
			sun_hours_winter = $9
		WHERE zone_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		zone.ZoneID,
		zone.ZoneName,
		zone.ZoneType,
		zone.GeometryGeoJSON,
		zone.SoilType,
		zone.SoilAmended,
		zone.IrrigationType,
		zone.SunHoursSummer,
		zone.SunHoursWinter,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_zone_update", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_zone_update_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden_zone", zone.ZoneID)
	}

	return nil
}

// Delete deletes a garden zone
func (r *PostgresGardenZoneRepository) Delete(ctx context.Context, zoneID string) error {
	query := `DELETE FROM garden_zones WHERE zone_id = $1`

	result, err := r.db.ExecContext(ctx, query, zoneID)
	if err != nil {
		return entity.NewDatabaseError("garden_zone_delete", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_zone_delete_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden_zone", zoneID)
	}

	return nil
}

// FindByIDs finds multiple zones by IDs (batch loading)
func (r *PostgresGardenZoneRepository) FindByIDs(ctx context.Context, zoneIDs []string) ([]*entity.GardenZone, error) {
	if len(zoneIDs) == 0 {
		return []*entity.GardenZone{}, nil
	}

	query := `
		SELECT
			zone_id, garden_id, zone_name, zone_type,
			ST_AsGeoJSON(geometry) as geometry,
			ST_Area(geometry::geography) as area_m2,
			soil_type, soil_amended, irrigation_type,
			sun_hours_summer, sun_hours_winter, created_at
		FROM garden_zones
		WHERE zone_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, zoneIDs)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_zone_find_by_ids", err)
	}
	defer rows.Close()

	var zones []*entity.GardenZone
	for rows.Next() {
		var zone entity.GardenZone
		var zoneName, zoneType sql.NullString
		var areaM2 sql.NullFloat64
		var soilType, irrigationType sql.NullString
		var sunHoursSummer, sunHoursWinter sql.NullInt32

		err := rows.Scan(
			&zone.ZoneID,
			&zone.GardenID,
			&zoneName,
			&zoneType,
			&zone.GeometryGeoJSON,
			&areaM2,
			&soilType,
			&zone.SoilAmended,
			&irrigationType,
			&sunHoursSummer,
			&sunHoursWinter,
			&zone.CreatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_zone_scan", err)
		}

		// Map nullable fields (same as FindByID)
		if zoneName.Valid {
			zone.ZoneName = &zoneName.String
		}
		if zoneType.Valid {
			typeValue := entity.ZoneType(zoneType.String)
			zone.ZoneType = &typeValue
		}
		if areaM2.Valid {
			zone.AreaM2 = &areaM2.Float64
		}
		if soilType.Valid {
			zone.SoilType = &soilType.String
		}
		if irrigationType.Valid {
			irrigationValue := entity.IrrigationType(irrigationType.String)
			zone.IrrigationType = &irrigationValue
		}
		if sunHoursSummer.Valid {
			hours := int(sunHoursSummer.Int32)
			zone.SunHoursSummer = &hours
		}
		if sunHoursWinter.Valid {
			hours := int(sunHoursWinter.Int32)
			zone.SunHoursWinter = &hours
		}

		zones = append(zones, &zone)
	}

	return zones, nil
}

// ValidateZoneWithinGarden validates that a zone is within the garden boundary using ST_Contains
func (r *PostgresGardenZoneRepository) ValidateZoneWithinGarden(ctx context.Context, gardenID, zoneGeometryGeoJSON string) error {
	// Validate GeoJSON first
	if err := database.ValidateGeoJSON(zoneGeometryGeoJSON); err != nil {
		return entity.NewSpatialError("zone_geometry_validation", err.Error())
	}

	query := `
		SELECT ST_Contains(g.boundary, ST_GeomFromGeoJSON($2)::geometry)
		FROM gardens g
		WHERE g.garden_id = $1
	`

	var isWithin sql.NullBool
	err := r.db.QueryRowContext(ctx, query, gardenID, zoneGeometryGeoJSON).Scan(&isWithin)

	if err == sql.ErrNoRows {
		return entity.NewNotFoundError("garden", gardenID)
	}
	if err != nil {
		return entity.NewDatabaseError("validate_zone_within_garden", err)
	}

	if !isWithin.Valid || !isWithin.Bool {
		return entity.NewSpatialError("zone_validation", "zone geometry is not within garden boundary")
	}

	return nil
}

// CheckZoneOverlaps checks if a zone overlaps with existing zones using ST_Overlaps
func (r *PostgresGardenZoneRepository) CheckZoneOverlaps(ctx context.Context, gardenID, zoneGeometryGeoJSON string, excludeZoneID *string) (bool, error) {
	// Validate GeoJSON first
	if err := database.ValidateGeoJSON(zoneGeometryGeoJSON); err != nil {
		return false, entity.NewSpatialError("zone_geometry_validation", err.Error())
	}

	// Build query with optional zone exclusion (for updates)
	query := `
		SELECT EXISTS(
			SELECT 1 FROM garden_zones
			WHERE garden_id = $1
			  AND ($2::uuid IS NULL OR zone_id != $2)
			  AND ST_Overlaps(geometry, ST_GeomFromGeoJSON($3)::geometry)
		)
	`

	var overlaps bool
	err := r.db.QueryRowContext(ctx, query, gardenID, excludeZoneID, zoneGeometryGeoJSON).Scan(&overlaps)

	if err != nil {
		return false, entity.NewDatabaseError("check_zone_overlaps", err)
	}

	return overlaps, nil
}

// CalculateTotalArea calculates the total area of all zones in a garden
func (r *PostgresGardenZoneRepository) CalculateTotalArea(ctx context.Context, gardenID string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(ST_Area(geometry::geography)), 0)
		FROM garden_zones
		WHERE garden_id = $1
	`

	var totalArea float64
	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(&totalArea)

	if err != nil {
		return 0, entity.NewDatabaseError("calculate_total_zone_area", err)
	}

	return totalArea, nil
}

// CalculateZoneArea calculates the area of a specific zone
func (r *PostgresGardenZoneRepository) CalculateZoneArea(ctx context.Context, zoneID string) (float64, error) {
	query := `
		SELECT ST_Area(geometry::geography)
		FROM garden_zones
		WHERE zone_id = $1
	`

	var areaM2 sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, zoneID).Scan(&areaM2)

	if err == sql.ErrNoRows {
		return 0, entity.NewNotFoundError("garden_zone", zoneID)
	}
	if err != nil {
		return 0, entity.NewDatabaseError("calculate_zone_area", err)
	}

	if !areaM2.Valid {
		return 0, fmt.Errorf("zone has no geometry")
	}

	return areaM2.Float64, nil
}

// CountByGardenID counts zones in a garden
func (r *PostgresGardenZoneRepository) CountByGardenID(ctx context.Context, gardenID string) (int, error) {
	query := `SELECT COUNT(*) FROM garden_zones WHERE garden_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(&count)

	if err != nil {
		return 0, entity.NewDatabaseError("count_zones_by_garden_id", err)
	}

	return count, nil
}
