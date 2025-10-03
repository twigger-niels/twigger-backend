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

// PostgresGardenFeatureRepository implements repository.GardenFeatureRepository using PostgreSQL + PostGIS
type PostgresGardenFeatureRepository struct {
	db *sql.DB
}

// NewPostgresGardenFeatureRepository creates a new PostgreSQL garden feature repository
func NewPostgresGardenFeatureRepository(db *sql.DB) repository.GardenFeatureRepository {
	return &PostgresGardenFeatureRepository{db: db}
}

// Create creates a new garden feature
func (r *PostgresGardenFeatureRepository) Create(ctx context.Context, feature *entity.GardenFeature) error {
	// Validate entity
	if err := feature.Validate(); err != nil {
		return entity.NewValidationError("garden_feature", err.Error())
	}

	// Validate GeoJSON before database insert (Gotcha #32)
	// Features can be Point or Polygon, so don't validate polygon closure
	if err := database.ValidateGeoJSON(feature.GeometryGeoJSON); err != nil {
		return entity.NewSpatialError("geometry_validation", err.Error())
	}

	// Generate ID if not provided
	if feature.FeatureID == "" {
		feature.FeatureID = uuid.New().String()
	}

	// Set timestamp
	feature.CreatedAt = time.Now()

	query := `
		INSERT INTO garden_features (
			feature_id, garden_id, feature_type, feature_name, geometry,
			height_m, canopy_diameter_m, deciduous, created_at
		) VALUES (
			$1, $2, $3, $4, ST_GeomFromGeoJSON($5),
			$6, $7, $8, $9
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		feature.FeatureID,
		feature.GardenID,
		feature.FeatureType,
		feature.FeatureName,
		feature.GeometryGeoJSON,
		feature.HeightM,
		feature.CanopyDiameterM,
		feature.Deciduous,
		feature.CreatedAt,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_feature_create", err)
	}

	return nil
}

// FindByID finds a garden feature by ID
func (r *PostgresGardenFeatureRepository) FindByID(ctx context.Context, featureID string) (*entity.GardenFeature, error) {
	query := `
		SELECT
			feature_id, garden_id, feature_type, feature_name,
			ST_AsGeoJSON(geometry) as geometry,
			height_m, canopy_diameter_m, deciduous, created_at
		FROM garden_features
		WHERE feature_id = $1
	`

	var feature entity.GardenFeature
	var featureName sql.NullString
	var heightM, canopyDiameterM sql.NullFloat64
	var deciduous sql.NullBool

	err := r.db.QueryRowContext(ctx, query, featureID).Scan(
		&feature.FeatureID,
		&feature.GardenID,
		&feature.FeatureType,
		&featureName,
		&feature.GeometryGeoJSON,
		&heightM,
		&canopyDiameterM,
		&deciduous,
		&feature.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, entity.NewNotFoundError("garden_feature", featureID)
	}
	if err != nil {
		return nil, entity.NewDatabaseError("garden_feature_find_by_id", err)
	}

	// Map nullable fields
	if featureName.Valid {
		feature.FeatureName = &featureName.String
	}
	if heightM.Valid {
		feature.HeightM = &heightM.Float64
	}
	if canopyDiameterM.Valid {
		feature.CanopyDiameterM = &canopyDiameterM.Float64
	}
	if deciduous.Valid {
		feature.Deciduous = &deciduous.Bool
	}

	return &feature, nil
}

// FindByGardenID finds all features in a garden
func (r *PostgresGardenFeatureRepository) FindByGardenID(ctx context.Context, gardenID string) ([]*entity.GardenFeature, error) {
	query := `
		SELECT
			feature_id, garden_id, feature_type, feature_name,
			ST_AsGeoJSON(geometry) as geometry,
			height_m, canopy_diameter_m, deciduous, created_at
		FROM garden_features
		WHERE garden_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gardenID)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_feature_find_by_garden_id", err)
	}
	defer rows.Close()

	var features []*entity.GardenFeature
	for rows.Next() {
		var feature entity.GardenFeature
		var featureName sql.NullString
		var heightM, canopyDiameterM sql.NullFloat64
		var deciduous sql.NullBool

		err := rows.Scan(
			&feature.FeatureID,
			&feature.GardenID,
			&feature.FeatureType,
			&featureName,
			&feature.GeometryGeoJSON,
			&heightM,
			&canopyDiameterM,
			&deciduous,
			&feature.CreatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_feature_scan", err)
		}

		// Map nullable fields
		if featureName.Valid {
			feature.FeatureName = &featureName.String
		}
		if heightM.Valid {
			feature.HeightM = &heightM.Float64
		}
		if canopyDiameterM.Valid {
			feature.CanopyDiameterM = &canopyDiameterM.Float64
		}
		if deciduous.Valid {
			feature.Deciduous = &deciduous.Bool
		}

		features = append(features, &feature)
	}

	if err = rows.Err(); err != nil {
		return nil, entity.NewDatabaseError("garden_feature_rows_iteration", err)
	}

	return features, nil
}

// FindByType finds features by type in a garden
func (r *PostgresGardenFeatureRepository) FindByType(ctx context.Context, gardenID string, featureType entity.FeatureType) ([]*entity.GardenFeature, error) {
	query := `
		SELECT
			feature_id, garden_id, feature_type, feature_name,
			ST_AsGeoJSON(geometry) as geometry,
			height_m, canopy_diameter_m, deciduous, created_at
		FROM garden_features
		WHERE garden_id = $1 AND feature_type = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gardenID, featureType)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_feature_find_by_type", err)
	}
	defer rows.Close()

	var features []*entity.GardenFeature
	for rows.Next() {
		var feature entity.GardenFeature
		var featureName sql.NullString
		var heightM, canopyDiameterM sql.NullFloat64
		var deciduous sql.NullBool

		err := rows.Scan(
			&feature.FeatureID,
			&feature.GardenID,
			&feature.FeatureType,
			&featureName,
			&feature.GeometryGeoJSON,
			&heightM,
			&canopyDiameterM,
			&deciduous,
			&feature.CreatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_feature_scan", err)
		}

		// Map nullable fields
		if featureName.Valid {
			feature.FeatureName = &featureName.String
		}
		if heightM.Valid {
			feature.HeightM = &heightM.Float64
		}
		if canopyDiameterM.Valid {
			feature.CanopyDiameterM = &canopyDiameterM.Float64
		}
		if deciduous.Valid {
			feature.Deciduous = &deciduous.Bool
		}

		features = append(features, &feature)
	}

	return features, nil
}

// Update updates a garden feature
func (r *PostgresGardenFeatureRepository) Update(ctx context.Context, feature *entity.GardenFeature) error {
	// Validate entity
	if err := feature.Validate(); err != nil {
		return entity.NewValidationError("garden_feature", err.Error())
	}

	// Validate GeoJSON (Gotcha #32)
	if err := database.ValidateGeoJSON(feature.GeometryGeoJSON); err != nil {
		return entity.NewSpatialError("geometry_validation", err.Error())
	}

	query := `
		UPDATE garden_features SET
			feature_type = $2,
			feature_name = $3,
			geometry = ST_GeomFromGeoJSON($4),
			height_m = $5,
			canopy_diameter_m = $6,
			deciduous = $7
		WHERE feature_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		feature.FeatureID,
		feature.FeatureType,
		feature.FeatureName,
		feature.GeometryGeoJSON,
		feature.HeightM,
		feature.CanopyDiameterM,
		feature.Deciduous,
	)

	if err != nil {
		return entity.NewDatabaseError("garden_feature_update", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_feature_update_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden_feature", feature.FeatureID)
	}

	return nil
}

// Delete deletes a garden feature
func (r *PostgresGardenFeatureRepository) Delete(ctx context.Context, featureID string) error {
	query := `DELETE FROM garden_features WHERE feature_id = $1`

	result, err := r.db.ExecContext(ctx, query, featureID)
	if err != nil {
		return entity.NewDatabaseError("garden_feature_delete", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entity.NewDatabaseError("garden_feature_delete_rows_affected", err)
	}

	if rowsAffected == 0 {
		return entity.NewNotFoundError("garden_feature", featureID)
	}

	return nil
}

// FindByIDs finds multiple features by IDs (batch loading)
func (r *PostgresGardenFeatureRepository) FindByIDs(ctx context.Context, featureIDs []string) ([]*entity.GardenFeature, error) {
	if len(featureIDs) == 0 {
		return []*entity.GardenFeature{}, nil
	}

	query := `
		SELECT
			feature_id, garden_id, feature_type, feature_name,
			ST_AsGeoJSON(geometry) as geometry,
			height_m, canopy_diameter_m, deciduous, created_at
		FROM garden_features
		WHERE feature_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, featureIDs)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_feature_find_by_ids", err)
	}
	defer rows.Close()

	var features []*entity.GardenFeature
	for rows.Next() {
		var feature entity.GardenFeature
		var featureName sql.NullString
		var heightM, canopyDiameterM sql.NullFloat64
		var deciduous sql.NullBool

		err := rows.Scan(
			&feature.FeatureID,
			&feature.GardenID,
			&feature.FeatureType,
			&featureName,
			&feature.GeometryGeoJSON,
			&heightM,
			&canopyDiameterM,
			&deciduous,
			&feature.CreatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_feature_scan", err)
		}

		// Map nullable fields
		if featureName.Valid {
			feature.FeatureName = &featureName.String
		}
		if heightM.Valid {
			feature.HeightM = &heightM.Float64
		}
		if canopyDiameterM.Valid {
			feature.CanopyDiameterM = &canopyDiameterM.Float64
		}
		if deciduous.Valid {
			feature.Deciduous = &deciduous.Bool
		}

		features = append(features, &feature)
	}

	return features, nil
}

// FindFeaturesWithHeight finds features that have height data (for shade calculations)
func (r *PostgresGardenFeatureRepository) FindFeaturesWithHeight(ctx context.Context, gardenID string) ([]*entity.GardenFeature, error) {
	query := `
		SELECT
			feature_id, garden_id, feature_type, feature_name,
			ST_AsGeoJSON(geometry) as geometry,
			height_m, canopy_diameter_m, deciduous, created_at
		FROM garden_features
		WHERE garden_id = $1
		  AND height_m IS NOT NULL
		  AND height_m > 0
		ORDER BY height_m DESC
	`

	rows, err := r.db.QueryContext(ctx, query, gardenID)
	if err != nil {
		return nil, entity.NewDatabaseError("garden_feature_find_with_height", err)
	}
	defer rows.Close()

	var features []*entity.GardenFeature
	for rows.Next() {
		var feature entity.GardenFeature
		var featureName sql.NullString
		var heightM, canopyDiameterM sql.NullFloat64
		var deciduous sql.NullBool

		err := rows.Scan(
			&feature.FeatureID,
			&feature.GardenID,
			&feature.FeatureType,
			&featureName,
			&feature.GeometryGeoJSON,
			&heightM,
			&canopyDiameterM,
			&deciduous,
			&feature.CreatedAt,
		)
		if err != nil {
			return nil, entity.NewDatabaseError("garden_feature_scan", err)
		}

		// Map nullable fields
		if featureName.Valid {
			feature.FeatureName = &featureName.String
		}
		if heightM.Valid {
			feature.HeightM = &heightM.Float64
		}
		if canopyDiameterM.Valid {
			feature.CanopyDiameterM = &canopyDiameterM.Float64
		}
		if deciduous.Valid {
			feature.Deciduous = &deciduous.Bool
		}

		features = append(features, &feature)
	}

	return features, nil
}

// FindTreesInGarden finds all tree features in a garden
func (r *PostgresGardenFeatureRepository) FindTreesInGarden(ctx context.Context, gardenID string) ([]*entity.GardenFeature, error) {
	return r.FindByType(ctx, gardenID, entity.FeatureTypeTree)
}

// CountByGardenID counts features in a garden
func (r *PostgresGardenFeatureRepository) CountByGardenID(ctx context.Context, gardenID string) (int, error) {
	query := `SELECT COUNT(*) FROM garden_features WHERE garden_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, gardenID).Scan(&count)

	if err != nil {
		return 0, entity.NewDatabaseError("count_features_by_garden_id", err)
	}

	return count, nil
}
