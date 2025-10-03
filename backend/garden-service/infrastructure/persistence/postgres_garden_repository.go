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

// PostgresGardenRepository implements repository.GardenRepository using PostgreSQL + PostGIS
type PostgresGardenRepository struct {
	db *sql.DB
}

// NewPostgresGardenRepository creates a new PostgreSQL garden repository
func NewPostgresGardenRepository(db *sql.DB) repository.GardenRepository {
	return &PostgresGardenRepository{db: db}
}

// Create creates a new garden
func (r *PostgresGardenRepository) Create(ctx context.Context, garden *entity.Garden) error {
	// Validate entity
	if err := garden.Validate(); err != nil {
		return entity.NewValidationError("garden", err.Error())
	}

	// Validate GeoJSON before database insert (Gotcha #32)
	if garden.BoundaryGeoJSON != nil {
		if err := database.ValidateGeoJSON(*garden.BoundaryGeoJSON); err != nil {
			return entity.NewSpatialError("boundary_validation", err.Error())
		}
		if err := database.ValidatePolygonClosure(*garden.BoundaryGeoJSON); err != nil {
			return entity.NewSpatialError("boundary_validation", err.Error())
		}
	}

	// Generate ID if not provided
	if garden.GardenID == "" {
		garden.GardenID = uuid.New().String()
	}

	// Set timestamps (Gotcha #11)
	now := time.Now()
	garden.CreatedAt = now
	garden.UpdatedAt = now

	query := `
		INSERT INTO gardens (
			garden_id, user_id, garden_name, boundary, location,
			elevation_m, slope_degrees, aspect, hardiness_zone, garden_type,
			created_at, updated_at
		) VALUES (
			$1, $2, $3,
			CASE WHEN $4::text IS NOT NULL THEN ST_GeomFromGeoJSON($4) ELSE NULL END,
			CASE WHEN $5::text IS NOT NULL THEN ST_GeomFromGeoJSON($5)::geography ELSE NULL END,
			$6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		garden.GardenID,
		garden.UserID,
		garden.GardenName,
		garden.BoundaryGeoJSON,
		garden.LocationGeoJSON,
		garden.ElevationM,
		garden.SlopeDegrees,
		garden.Aspect,
		garden.HardinessZone,
		garden.GardenType,
		garden.CreatedAt,
		garden.UpdatedAt,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_create", err)
	}

	return nil
}

// FindByID finds a garden by ID
func (r *PostgresGardenRepository) FindByID(ctx context.Context, gardenID string) (*entity.Garden, error) {
	query := `
		SELECT
			garden_id, user_id, garden_name,
			ST_AsGeoJSON(boundary) as boundary,
			ST_AsGeoJSON(location::geometry) as location,
			elevation_m, slope_degrees, aspect,
			hardiness_zone, garden_type,
			created_at, updated_at
		FROM gardens
		WHERE garden_id = $1
	`

	var garden entity.Garden
	var boundaryJSON, locationJSON sql.NullString
	var elevationM, slopeDegrees sql.NullFloat64
	var aspect, hardinessZone, gardenType sql.NullString

	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(
		&garden.GardenID,
		&garden.UserID,
		&garden.GardenName,
		&boundaryJSON,
		&locationJSON,
		&elevationM,
		&slopeDegrees,
		&aspect,
		&hardinessZone,
		&gardenType,
		&garden.CreatedAt,
		&garden.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, entity.NewNotFoundError("garden", gardenID)
	}
	if err != nil {
		return nil, entity.NewDatabaseError("garden_find_by_id", err)
	}

	// Map nullable fields
	if boundaryJSON.Valid {
		garden.BoundaryGeoJSON = &boundaryJSON.String
	}
	if locationJSON.Valid {
		garden.LocationGeoJSON = &locationJSON.String
	}
	if elevationM.Valid {
		garden.ElevationM = &elevationM.Float64
	}
	if slopeDegrees.Valid {
		garden.SlopeDegrees = &slopeDegrees.Float64
	}
	if aspect.Valid {
		aspectValue := entity.Aspect(aspect.String)
		garden.Aspect = &aspectValue
	}
	if hardinessZone.Valid {
		garden.HardinessZone = &hardinessZone.String
	}
	if gardenType.Valid {
		typeValue := entity.GardenType(gardenType.String)
		garden.GardenType = &typeValue
	}

	return &garden, nil
}

// FindByUserID finds all gardens for a user with pagination
func (r *PostgresGardenRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Garden, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	query := `
		SELECT
			garden_id, user_id, garden_name,
			ST_AsGeoJSON(boundary) as boundary,
			ST_AsGeoJSON(location::geometry) as location,
			elevation_m, slope_degrees, aspect,
			hardiness_zone, garden_type,
			created_at, updated_at
		FROM gardens
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_find_by_user_id", err)
	}
	defer rows.Close()

	var gardens []*entity.Garden
	for rows.Next() {
		var garden entity.Garden
		var boundaryJSON, locationJSON sql.NullString
		var elevationM, slopeDegrees sql.NullFloat64
		var aspect, hardinessZone, gardenType sql.NullString

		err := rows.Scan(
			&garden.GardenID,
			&garden.UserID,
			&garden.GardenName,
			&boundaryJSON,
			&locationJSON,
			&elevationM,
			&slopeDegrees,
			&aspect,
			&hardinessZone,
			&gardenType,
			&garden.CreatedAt,
			&garden.UpdatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_scan", err)
		}

		// Map nullable fields
		if boundaryJSON.Valid {
			garden.BoundaryGeoJSON = &boundaryJSON.String
		}
		if locationJSON.Valid {
			garden.LocationGeoJSON = &locationJSON.String
		}
		if elevationM.Valid {
			garden.ElevationM = &elevationM.Float64
		}
		if slopeDegrees.Valid {
			garden.SlopeDegrees = &slopeDegrees.Float64
		}
		if aspect.Valid {
			aspectValue := entity.Aspect(aspect.String)
			garden.Aspect = &aspectValue
		}
		if hardinessZone.Valid {
			garden.HardinessZone = &hardinessZone.String
		}
		if gardenType.Valid {
			typeValue := entity.GardenType(gardenType.String)
			garden.GardenType = &typeValue
		}

		gardens = append(gardens, &garden)
	}

	if err = rows.Err(); err != nil {
		return nil, entity.NewDatabaseError("garden_rows_iteration", err)
	}

	return gardens, nil
}

// Update updates a garden
func (r *PostgresGardenRepository) Update(ctx context.Context, garden *entity.Garden) error {
	// Validate entity
	if err := garden.Validate(); err != nil {
		return entity.NewValidationError("garden", err.Error())
	}

	// Validate GeoJSON if provided (Gotcha #32)
	if garden.BoundaryGeoJSON != nil {
		if err := database.ValidateGeoJSON(*garden.BoundaryGeoJSON); err != nil {
			return entity.NewSpatialError("boundary_validation", err.Error())
		}
	}

	// Update timestamp (Gotcha #11)
	garden.UpdatedAt = time.Now()

	query := `
		UPDATE gardens SET
			garden_name = $2,
			boundary = CASE WHEN $3::text IS NOT NULL THEN ST_GeomFromGeoJSON($3) ELSE NULL END,
			location = CASE WHEN $4::text IS NOT NULL THEN ST_GeomFromGeoJSON($4)::geography ELSE NULL END,
			elevation_m = $5,
			slope_degrees = $6,
			aspect = $7,
			hardiness_zone = $8,
			garden_type = $9,
			updated_at = $10
		WHERE garden_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		garden.GardenID,
		garden.GardenName,
		garden.BoundaryGeoJSON,
		garden.LocationGeoJSON,
		garden.ElevationM,
		garden.SlopeDegrees,
		garden.Aspect,
		garden.HardinessZone,
		garden.GardenType,
		garden.UpdatedAt,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_update", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_update_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden", garden.GardenID)
	}

	return nil
}

// Delete deletes a garden
func (r *PostgresGardenRepository) Delete(ctx context.Context, gardenID string) error {
	query := `DELETE FROM gardens WHERE garden_id = $1`

	result, err := r.db.ExecContext(ctx, query, gardenID)
	if err != nil {
		return entity.NewDatabaseError("garden_delete", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_delete_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden", gardenID)
	}

	return nil
}

// FindByIDs finds multiple gardens by IDs (batch loading to avoid N+1)
func (r *PostgresGardenRepository) FindByIDs(ctx context.Context, gardenIDs []string) ([]*entity.Garden, error) {
	if len(gardenIDs) == 0 {
		return []*entity.Garden{}, nil
	}

	query := `
		SELECT
			garden_id, user_id, garden_name,
			ST_AsGeoJSON(boundary) as boundary,
			ST_AsGeoJSON(location::geometry) as location,
			elevation_m, slope_degrees, aspect,
			hardiness_zone, garden_type,
			created_at, updated_at
		FROM gardens
		WHERE garden_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, gardenIDs)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_find_by_ids", err)
	}
	defer rows.Close()

	var gardens []*entity.Garden
	for rows.Next() {
		var garden entity.Garden
		var boundaryJSON, locationJSON sql.NullString
		var elevationM, slopeDegrees sql.NullFloat64
		var aspect, hardinessZone, gardenType sql.NullString

		err := rows.Scan(
			&garden.GardenID,
			&garden.UserID,
			&garden.GardenName,
			&boundaryJSON,
			&locationJSON,
			&elevationM,
			&slopeDegrees,
			&aspect,
			&hardinessZone,
			&gardenType,
			&garden.CreatedAt,
			&garden.UpdatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_scan", err)
		}

		// Map nullable fields (same as FindByID)
		if boundaryJSON.Valid {
			garden.BoundaryGeoJSON = &boundaryJSON.String
		}
		if locationJSON.Valid {
			garden.LocationGeoJSON = &locationJSON.String
		}
		if elevationM.Valid {
			garden.ElevationM = &elevationM.Float64
		}
		if slopeDegrees.Valid {
			garden.SlopeDegrees = &slopeDegrees.Float64
		}
		if aspect.Valid {
			aspectValue := entity.Aspect(aspect.String)
			garden.Aspect = &aspectValue
		}
		if hardinessZone.Valid {
			garden.HardinessZone = &hardinessZone.String
		}
		if gardenType.Valid {
			typeValue := entity.GardenType(gardenType.String)
			garden.GardenType = &typeValue
		}

		gardens = append(gardens, &garden)
	}

	return gardens, nil
}

// FindByLocation finds gardens near a location using ST_DWithin
func (r *PostgresGardenRepository) FindByLocation(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Garden, error) {
	// Validate coordinates (Gotcha #33)
	if err := database.ValidateCoordinates(lat, lng); err != nil {
		return nil, entity.NewInvalidInputError("coordinates", err.Error())
	}

	radiusMeters := radiusKm * 1000

	query := `
		SELECT
			garden_id, user_id, garden_name,
			ST_AsGeoJSON(boundary) as boundary,
			ST_AsGeoJSON(location::geometry) as location,
			elevation_m, slope_degrees, aspect,
			hardiness_zone, garden_type,
			created_at, updated_at,
			ST_Distance(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_m
		FROM gardens
		WHERE ST_DWithin(
			location,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3
		)
		ORDER BY distance_m
	`

	rows, err := r.db.QueryContext(ctx, query, lng, lat, radiusMeters)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_find_by_location", err)
	}
	defer rows.Close()

	var gardens []*entity.Garden
	for rows.Next() {
		var garden entity.Garden
		var boundaryJSON, locationJSON sql.NullString
		var elevationM, slopeDegrees sql.NullFloat64
		var aspect, hardinessZone, gardenType sql.NullString
		var distanceM float64

		err := rows.Scan(
			&garden.GardenID,
			&garden.UserID,
			&garden.GardenName,
			&boundaryJSON,
			&locationJSON,
			&elevationM,
			&slopeDegrees,
			&aspect,
			&hardinessZone,
			&gardenType,
			&garden.CreatedAt,
			&garden.UpdatedAt,
			&distanceM,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_scan", err)
		}

		// Map nullable fields
		if boundaryJSON.Valid {
			garden.BoundaryGeoJSON = &boundaryJSON.String
		}
		if locationJSON.Valid {
			garden.LocationGeoJSON = &locationJSON.String
		}
		if elevationM.Valid {
			garden.ElevationM = &elevationM.Float64
		}
		if slopeDegrees.Valid {
			garden.SlopeDegrees = &slopeDegrees.Float64
		}
		if aspect.Valid {
			aspectValue := entity.Aspect(aspect.String)
			garden.Aspect = &aspectValue
		}
		if hardinessZone.Valid {
			garden.HardinessZone = &hardinessZone.String
		}
		if gardenType.Valid {
			typeValue := entity.GardenType(gardenType.String)
			garden.GardenType = &typeValue
		}

		gardens = append(gardens, &garden)
	}

	return gardens, nil
}

// CalculateArea calculates the area of a garden using ST_Area
func (r *PostgresGardenRepository) CalculateArea(ctx context.Context, gardenID string) (float64, error) {
	query := `
		SELECT ST_Area(boundary::geography)
		FROM gardens
		WHERE garden_id = $1
	`

	var areaM2 sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(&areaM2)

	if err == sql.ErrNoRows {
		return 0, entity.NewNotFoundError("garden", gardenID)
	}
	if err != nil {
		return 0, entity.NewDatabaseError("garden_calculate_area", err)
	}

	if !areaM2.Valid {
		return 0, nil // No boundary defined
	}

	return areaM2.Float64, nil
}

// GetCenterPoint gets the center point of a garden
func (r *PostgresGardenRepository) GetCenterPoint(ctx context.Context, gardenID string) (lat, lng float64, err error) {
	query := `
		SELECT
			ST_Y(ST_Centroid(boundary)) as lat,
			ST_X(ST_Centroid(boundary)) as lng
		FROM gardens
		WHERE garden_id = $1
	`

	err = r.db.QueryRowContext(ctx, query, gardenID).Scan(&lat, &lng)

	if err == sql.ErrNoRows {
		return 0, 0, entity.NewNotFoundError("garden", gardenID)
	}
	if err != nil {
		return 0, 0, entity.NewDatabaseError("garden_get_center_point", err)
	}

	return lat, lng, nil
}

// DetectHardinessZone detects the hardiness zone using ST_Contains
func (r *PostgresGardenRepository) DetectHardinessZone(ctx context.Context, gardenID string) (string, error) {
	query := `
		SELECT cz.zone_code
		FROM gardens g
		JOIN climate_zones cz ON ST_Contains(cz.zone_geometry, g.boundary)
		WHERE g.garden_id = $1
		LIMIT 1
	`

	var zoneCode string
	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(&zoneCode)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no climate zone found for garden")
	}
	if err != nil {
		return "", entity.NewDatabaseError("garden_detect_hardiness_zone", err)
	}

	return zoneCode, nil
}

// ValidateBoundary validates a GeoJSON boundary string
func (r *PostgresGardenRepository) ValidateBoundary(ctx context.Context, boundaryGeoJSON string) error {
	if err := database.ValidateGeoJSON(boundaryGeoJSON); err != nil {
		return entity.NewSpatialError("boundary_validation", err.Error())
	}

	if err := database.ValidatePolygonClosure(boundaryGeoJSON); err != nil {
		return entity.NewSpatialError("boundary_validation", err.Error())
	}

	return nil
}

// CheckBoundaryValid checks if a garden's boundary is valid
func (r *PostgresGardenRepository) CheckBoundaryValid(ctx context.Context, gardenID string) (bool, error) {
	query := `
		SELECT ST_IsValid(boundary)
		FROM gardens
		WHERE garden_id = $1
	`

	var isValid sql.NullBool
	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(&isValid)

	if err == sql.ErrNoRows {
		return false, entity.NewNotFoundError("garden", gardenID)
	}
	if err != nil {
		return false, entity.NewDatabaseError("garden_check_boundary_valid", err)
	}

	if !isValid.Valid {
		return false, nil // No boundary
	}

	return isValid.Bool, nil
}

// CountByUserID counts gardens for a user
func (r *PostgresGardenRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM gardens WHERE user_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)

	if err != nil {
		return 0, entity.NewDatabaseError("garden_count_by_user_id", err)
	}

	return count, nil
}

// GetTotalArea gets total area of all gardens for a user
func (r *PostgresGardenRepository) GetTotalArea(ctx context.Context, userID string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(ST_Area(boundary::geography)), 0)
		FROM gardens
		WHERE user_id = $1
	`

	var totalArea float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&totalArea)

	if err != nil {
		return 0, entity.NewDatabaseError("garden_get_total_area", err)
	}

	return totalArea, nil
}
