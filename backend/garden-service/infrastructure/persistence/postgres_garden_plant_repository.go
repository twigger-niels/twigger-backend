package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/backend/garden-service/domain/repository"
	"twigger-backend/backend/garden-service/infrastructure/database"
)

// PostgresGardenPlantRepository implements repository.GardenPlantRepository using PostgreSQL + PostGIS
type PostgresGardenPlantRepository struct {
	db *sql.DB
}

// NewPostgresGardenPlantRepository creates a new PostgreSQL garden plant repository
func NewPostgresGardenPlantRepository(db *sql.DB) repository.GardenPlantRepository {
	return &PostgresGardenPlantRepository{db: db}
}

// Create creates a new garden plant placement
func (r *PostgresGardenPlantRepository) Create(ctx context.Context, gardenPlant *entity.GardenPlant) error {
	// Validate entity
	if err := gardenPlant.Validate(); err != nil {
		return entity.NewValidationError("garden_plant", err.Error())
	}

	// Validate GeoJSON before database insert (Gotcha #32)
	if err := database.ValidateGeoJSON(gardenPlant.LocationGeoJSON); err != nil {
		return entity.NewSpatialError("location_validation", err.Error())
	}

	// Generate ID if not provided
	if gardenPlant.GardenPlantID == "" {
		gardenPlant.GardenPlantID = uuid.New().String()
	}

	// Set timestamps (Gotcha #11)
	now := time.Now()
	gardenPlant.CreatedAt = now
	gardenPlant.UpdatedAt = now

	query := `
		INSERT INTO garden_plants (
			garden_plant_id, garden_id, zone_id, plant_id, location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, ST_GeomFromGeoJSON($5),
			$6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		gardenPlant.GardenPlantID,
		gardenPlant.GardenID,
		gardenPlant.ZoneID,
		gardenPlant.PlantID,
		gardenPlant.LocationGeoJSON,
		gardenPlant.PlantedDate,
		gardenPlant.RemovedDate,
		gardenPlant.Quantity,
		gardenPlant.PlantSource,
		gardenPlant.HealthStatus,
		gardenPlant.Notes,
		gardenPlant.CreatedAt,
		gardenPlant.UpdatedAt,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_plant_create", err)
	}

	return nil
}

// FindByID finds a garden plant by ID
func (r *PostgresGardenPlantRepository) FindByID(ctx context.Context, gardenPlantID string) (*entity.GardenPlant, error) {
	query := `
		SELECT
			garden_plant_id, garden_id, zone_id, plant_id,
			ST_AsGeoJSON(location) as location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		FROM garden_plants
		WHERE garden_plant_id = $1
	`

	var gardenPlant entity.GardenPlant
	var zoneID sql.NullString
	var plantedDate, removedDate sql.NullTime
	var plantSource, healthStatus, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, gardenPlantID).Scan(
		&gardenPlant.GardenPlantID,
		&gardenPlant.GardenID,
		&zoneID,
		&gardenPlant.PlantID,
		&gardenPlant.LocationGeoJSON,
		&plantedDate,
		&removedDate,
		&gardenPlant.Quantity,
		&plantSource,
		&healthStatus,
		&notes,
		&gardenPlant.CreatedAt,
		&gardenPlant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, entity.NewNotFoundError("garden_plant", gardenPlantID)
	}
	if err != nil {
		return nil, entity.NewDatabaseError("garden_plant_find_by_id", err)
	}

	// Map nullable fields
	if zoneID.Valid {
		gardenPlant.ZoneID = &zoneID.String
	}
	if plantedDate.Valid {
		gardenPlant.PlantedDate = &plantedDate.Time
	}
	if removedDate.Valid {
		gardenPlant.RemovedDate = &removedDate.Time
	}
	if plantSource.Valid {
		gardenPlant.PlantSource = &plantSource.String
	}
	if healthStatus.Valid {
		statusValue := entity.HealthStatus(healthStatus.String)
		gardenPlant.HealthStatus = &statusValue
	}
	if notes.Valid {
		gardenPlant.Notes = &notes.String
	}

	return &gardenPlant, nil
}

// FindByGardenID finds all plants in a garden
func (r *PostgresGardenPlantRepository) FindByGardenID(ctx context.Context, gardenID string, includeRemoved bool) ([]*entity.GardenPlant, error) {
	query := `
		SELECT
			garden_plant_id, garden_id, zone_id, plant_id,
			ST_AsGeoJSON(location) as location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		FROM garden_plants
		WHERE garden_id = $1
	`

	// Filter out removed plants if requested
	if !includeRemoved {
		query += " AND removed_date IS NULL"
	}

	query += " ORDER BY planted_date DESC NULLS LAST, created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, gardenID)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_plant_find_by_garden_id", err)
	}
	defer rows.Close()

	return r.scanGardenPlants(rows)
}

// FindByZoneID finds all plants in a zone
func (r *PostgresGardenPlantRepository) FindByZoneID(ctx context.Context, zoneID string, includeRemoved bool) ([]*entity.GardenPlant, error) {
	query := `
		SELECT
			garden_plant_id, garden_id, zone_id, plant_id,
			ST_AsGeoJSON(location) as location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		FROM garden_plants
		WHERE zone_id = $1
	`

	if !includeRemoved {
		query += " AND removed_date IS NULL"
	}

	query += " ORDER BY planted_date DESC NULLS LAST"

	rows, err := r.db.QueryContext(ctx, query, zoneID)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_plant_find_by_zone_id", err)
	}
	defer rows.Close()

	return r.scanGardenPlants(rows)
}

// FindByPlantID finds all instances of a plant across gardens
func (r *PostgresGardenPlantRepository) FindByPlantID(ctx context.Context, plantID string) ([]*entity.GardenPlant, error) {
	query := `
		SELECT
			garden_plant_id, garden_id, zone_id, plant_id,
			ST_AsGeoJSON(location) as location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		FROM garden_plants
		WHERE plant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, plantID)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_plant_find_by_plant_id", err)
	}
	defer rows.Close()

	return r.scanGardenPlants(rows)
}

// Update updates a garden plant
func (r *PostgresGardenPlantRepository) Update(ctx context.Context, gardenPlant *entity.GardenPlant) error {
	// Validate entity
	if err := gardenPlant.Validate(); err != nil {
		return entity.NewValidationError("garden_plant", err.Error())
	}

	// Validate GeoJSON (Gotcha #32)
	if err := database.ValidateGeoJSON(gardenPlant.LocationGeoJSON); err != nil {
		return entity.NewSpatialError("location_validation", err.Error())
	}

	// Update timestamp (Gotcha #11)
	gardenPlant.UpdatedAt = time.Now()

	query := `
		UPDATE garden_plants SET
			zone_id = $2,
			location = ST_GeomFromGeoJSON($3),
			planted_date = $4,
			removed_date = $5,
			quantity = $6,
			plant_source = $7,
			health_status = $8,
			notes = $9,
			updated_at = $10
		WHERE garden_plant_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		gardenPlant.GardenPlantID,
		gardenPlant.ZoneID,
		gardenPlant.LocationGeoJSON,
		gardenPlant.PlantedDate,
		gardenPlant.RemovedDate,
		gardenPlant.Quantity,
		gardenPlant.PlantSource,
		gardenPlant.HealthStatus,
		gardenPlant.Notes,
		gardenPlant.UpdatedAt,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_plant_update", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_plant_update_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden_plant", gardenPlant.GardenPlantID)
	}

	return nil
}

// Delete deletes a garden plant
func (r *PostgresGardenPlantRepository) Delete(ctx context.Context, gardenPlantID string) error {
	query := `DELETE FROM garden_plants WHERE garden_plant_id = $1`

	result, err := r.db.ExecContext(ctx, query, gardenPlantID)
	if err != nil {
		return entity.NewDatabaseError("garden_plant_delete", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_plant_delete_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden_plant", gardenPlantID)
	}

	return nil
}

// FindByIDs finds multiple garden plants by IDs (batch loading)
func (r *PostgresGardenPlantRepository) FindByIDs(ctx context.Context, gardenPlantIDs []string) ([]*entity.GardenPlant, error) {
	if len(gardenPlantIDs) == 0 {
		return []*entity.GardenPlant{}, nil
	}

	query := `
		SELECT
			garden_plant_id, garden_id, zone_id, plant_id,
			ST_AsGeoJSON(location) as location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		FROM garden_plants
		WHERE garden_plant_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, gardenPlantIDs)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_plant_find_by_ids", err)
	}
	defer rows.Close()

	return r.scanGardenPlants(rows)
}

// BulkCreate creates multiple garden plants in a single transaction
func (r *PostgresGardenPlantRepository) BulkCreate(ctx context.Context, gardenPlants []*entity.GardenPlant) error {
	if len(gardenPlants) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.NewDatabaseError("bulk_create_begin_transaction", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	query := `
		INSERT INTO garden_plants (
			garden_plant_id, garden_id, zone_id, plant_id, location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, ST_GeomFromGeoJSON($5),
			$6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		return entity.NewDatabaseError("bulk_create_prepare", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, gardenPlant := range gardenPlants {
		// Validate each plant
		if err := gardenPlant.Validate(); err != nil {
			_ = tx.Rollback()
			return entity.NewValidationError("garden_plant", err.Error())
		}

		// Validate GeoJSON
		if err := database.ValidateGeoJSON(gardenPlant.LocationGeoJSON); err != nil {
			_ = tx.Rollback()
			return entity.NewSpatialError("location_validation", err.Error())
		}

		// Generate ID if needed
		if gardenPlant.GardenPlantID == "" {
			gardenPlant.GardenPlantID = uuid.New().String()
		}

		// Set timestamps
		gardenPlant.CreatedAt = now
		gardenPlant.UpdatedAt = now

		_, err := stmt.ExecContext(ctx,
			gardenPlant.GardenPlantID,
			gardenPlant.GardenID,
			gardenPlant.ZoneID,
			gardenPlant.PlantID,
			gardenPlant.LocationGeoJSON,
			gardenPlant.PlantedDate,
			gardenPlant.RemovedDate,
			gardenPlant.Quantity,
			gardenPlant.PlantSource,
			gardenPlant.HealthStatus,
			gardenPlant.Notes,
			gardenPlant.CreatedAt,
			gardenPlant.UpdatedAt,
		)

		if err != nil {
			_ = tx.Rollback()
			return entity.NewDatabaseError("bulk_create_exec", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return entity.NewDatabaseError("bulk_create_commit", err)
	}

	return nil
}

// CheckPlantSpacing finds plants within minimum distance of a location (ST_DWithin)
func (r *PostgresGardenPlantRepository) CheckPlantSpacing(ctx context.Context, gardenID, locationGeoJSON string, minDistanceM float64) ([]*entity.GardenPlant, error) {
	// Validate GeoJSON
	if err := database.ValidateGeoJSON(locationGeoJSON); err != nil {
		return nil, entity.NewSpatialError("location_validation", err.Error())
	}

	query := `
		SELECT
			garden_plant_id, garden_id, zone_id, plant_id,
			ST_AsGeoJSON(location) as location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at,
			ST_Distance(location::geography, ST_GeomFromGeoJSON($3)::geography) as distance_m
		FROM garden_plants
		WHERE garden_id = $1
		  AND removed_date IS NULL
		  AND ST_DWithin(
		      location::geography,
		      ST_GeomFromGeoJSON($3)::geography,
		      $2
		  )
		ORDER BY distance_m
	`

	rows, err := r.db.QueryContext(ctx, query, gardenID, minDistanceM, locationGeoJSON)
	if err != nil {
		return nil, entity.NewDatabaseError("check_plant_spacing", err)
	}
	defer rows.Close()

	var plants []*entity.GardenPlant
	for rows.Next() {
		var gardenPlant entity.GardenPlant
		var zoneID sql.NullString
		var plantedDate, removedDate sql.NullTime
		var plantSource, healthStatus, notes sql.NullString
		var distanceM float64

		err := rows.Scan(
			&gardenPlant.GardenPlantID,
			&gardenPlant.GardenID,
			&zoneID,
			&gardenPlant.PlantID,
			&gardenPlant.LocationGeoJSON,
			&plantedDate,
			&removedDate,
			&gardenPlant.Quantity,
			&plantSource,
			&healthStatus,
			&notes,
			&gardenPlant.CreatedAt,
			&gardenPlant.UpdatedAt,
			&distanceM,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_plant_scan", err)
		}

		// Map nullable fields
		if zoneID.Valid {
			gardenPlant.ZoneID = &zoneID.String
		}
		if plantedDate.Valid {
			gardenPlant.PlantedDate = &plantedDate.Time
		}
		if removedDate.Valid {
			gardenPlant.RemovedDate = &removedDate.Time
		}
		if plantSource.Valid {
			gardenPlant.PlantSource = &plantSource.String
		}
		if healthStatus.Valid {
			statusValue := entity.HealthStatus(healthStatus.String)
			gardenPlant.HealthStatus = &statusValue
		}
		if notes.Valid {
			gardenPlant.Notes = &notes.String
		}

		plants = append(plants, &gardenPlant)
	}

	return plants, nil
}

// FindInZone finds plants within a zone using ST_Contains
func (r *PostgresGardenPlantRepository) FindInZone(ctx context.Context, zoneID string) ([]*entity.GardenPlant, error) {
	query := `
		SELECT
			gp.garden_plant_id, gp.garden_id, gp.zone_id, gp.plant_id,
			ST_AsGeoJSON(gp.location) as location,
			gp.planted_date, gp.removed_date, gp.quantity, gp.plant_source,
			gp.health_status, gp.notes, gp.created_at, gp.updated_at
		FROM garden_plants gp
		JOIN garden_zones gz ON gz.zone_id = $1
		WHERE ST_Contains(gz.geometry, gp.location)
		  AND gp.removed_date IS NULL
		ORDER BY gp.planted_date DESC NULLS LAST
	`

	rows, err := r.db.QueryContext(ctx, query, zoneID)
	if err != nil {
		return nil, entity.NewDatabaseError("find_in_zone", err)
	}
	defer rows.Close()

	return r.scanGardenPlants(rows)
}

// ValidatePlantLocation validates that a plant location is within the garden boundary
func (r *PostgresGardenPlantRepository) ValidatePlantLocation(ctx context.Context, gardenID, locationGeoJSON string) error {
	// Validate GeoJSON first
	if err := database.ValidateGeoJSON(locationGeoJSON); err != nil {
		return entity.NewSpatialError("location_validation", err.Error())
	}

	query := `
		SELECT ST_Contains(g.boundary, ST_GeomFromGeoJSON($2)::geometry)
		FROM gardens g
		WHERE g.garden_id = $1
	`

	var isWithin sql.NullBool
	err := r.db.QueryRowContext(ctx, query, gardenID, locationGeoJSON).Scan(&isWithin)

	if err == sql.ErrNoRows {
		return entity.NewNotFoundError("garden", gardenID)
	}
	if err != nil {
		return entity.NewDatabaseError("validate_plant_location", err)
	}

	if !isWithin.Valid || !isWithin.Bool {
		return entity.NewSpatialError("plant_location_validation", "plant location is not within garden boundary")
	}

	return nil
}

// FindByHealthStatus finds plants by health status in a garden
func (r *PostgresGardenPlantRepository) FindByHealthStatus(ctx context.Context, gardenID string, status entity.HealthStatus) ([]*entity.GardenPlant, error) {
	query := `
		SELECT
			garden_plant_id, garden_id, zone_id, plant_id,
			ST_AsGeoJSON(location) as location,
			planted_date, removed_date, quantity, plant_source,
			health_status, notes, created_at, updated_at
		FROM garden_plants
		WHERE garden_id = $1
		  AND health_status = $2
		  AND removed_date IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, gardenID, status)
	if err != nil {
		return nil, entity.NewDatabaseError("find_by_health_status", err)
	}
	defer rows.Close()

	return r.scanGardenPlants(rows)
}

// FindActiveInGarden finds all active (not removed) plants in a garden
func (r *PostgresGardenPlantRepository) FindActiveInGarden(ctx context.Context, gardenID string) ([]*entity.GardenPlant, error) {
	return r.FindByGardenID(ctx, gardenID, false)
}

// CountByGardenID counts plants in a garden
func (r *PostgresGardenPlantRepository) CountByGardenID(ctx context.Context, gardenID string, includeRemoved bool) (int, error) {
	query := `SELECT COUNT(*) FROM garden_plants WHERE garden_id = $1`

	if !includeRemoved {
		query += " AND removed_date IS NULL"
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(&count)

	if err != nil {
		return 0, entity.NewDatabaseError("count_by_garden_id", err)
	}

	return count, nil
}

// CountByPlantID counts instances of a plant across gardens
func (r *PostgresGardenPlantRepository) CountByPlantID(ctx context.Context, plantID string) (int, error) {
	query := `SELECT COUNT(*) FROM garden_plants WHERE plant_id = $1 AND removed_date IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, plantID).Scan(&count)

	if err != nil {
		return 0, entity.NewDatabaseError("count_by_plant_id", err)
	}

	return count, nil
}

// scanGardenPlants is a helper method to scan rows into GardenPlant entities
func (r *PostgresGardenPlantRepository) scanGardenPlants(rows *sql.Rows) ([]*entity.GardenPlant, error) {
	var plants []*entity.GardenPlant

	for rows.Next() {
		var gardenPlant entity.GardenPlant
		var zoneID sql.NullString
		var plantedDate, removedDate sql.NullTime
		var plantSource, healthStatus, notes sql.NullString

		err := rows.Scan(
			&gardenPlant.GardenPlantID,
			&gardenPlant.GardenID,
			&zoneID,
			&gardenPlant.PlantID,
			&gardenPlant.LocationGeoJSON,
			&plantedDate,
			&removedDate,
			&gardenPlant.Quantity,
			&plantSource,
			&healthStatus,
			&notes,
			&gardenPlant.CreatedAt,
			&gardenPlant.UpdatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_plant_scan", err)
		}

		// Map nullable fields
		if zoneID.Valid {
			gardenPlant.ZoneID = &zoneID.String
		}
		if plantedDate.Valid {
			gardenPlant.PlantedDate = &plantedDate.Time
		}
		if removedDate.Valid {
			gardenPlant.RemovedDate = &removedDate.Time
		}
		if plantSource.Valid {
			gardenPlant.PlantSource = &plantSource.String
		}
		if healthStatus.Valid {
			statusValue := entity.HealthStatus(healthStatus.String)
			gardenPlant.HealthStatus = &statusValue
		}
		if notes.Valid {
			gardenPlant.Notes = &notes.String
		}

		plants = append(plants, &gardenPlant)
	}

	if err := rows.Err(); err != nil {
		return nil, entity.NewDatabaseError("garden_plant_rows_iteration", err)
	}

	return plants, nil
}
