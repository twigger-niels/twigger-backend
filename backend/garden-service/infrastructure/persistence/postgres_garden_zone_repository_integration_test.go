// +build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"twigger-backend/backend/garden-service/domain/entity"
	testhelpers "twigger-backend/backend/garden-service/infrastructure/database/testing"
)

func TestGardenZoneRepository_Create(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	// Clean and setup schema
	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	// Seed test data
	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	zoneType := entity.ZoneTypeBed
	irrigationType := entity.IrrigationTypeDrip
	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
		ZoneType:        &zoneType,
		IrrigationType:  &irrigationType,
	}

	err = repo.Create(ctx, zone)
	require.NoError(t, err)
	assert.NotEmpty(t, zone.ZoneID)
	assert.False(t, zone.CreatedAt.IsZero())
}

func TestGardenZoneRepository_Create_InvalidGeoJSON(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	zoneType := entity.ZoneTypeBed
	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.InvalidPolygon, // Unclosed polygon
		ZoneType:        &zoneType,
	}

	err = repo.Create(ctx, zone)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "polygon")
}

func TestGardenZoneRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create zone
	zoneType := entity.ZoneTypeBed
	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
		ZoneType:        &zoneType,
	}
	err = repo.Create(ctx, zone)
	require.NoError(t, err)

	// Find zone
	found, err := repo.FindByID(ctx, zone.ZoneID)
	require.NoError(t, err)
	assert.Equal(t, zone.ZoneID, found.ZoneID)
	assert.Equal(t, zone.GardenID, found.GardenID)
	assert.NotEmpty(t, found.GeometryGeoJSON)
	assert.Equal(t, zoneType, *found.ZoneType)
}

func TestGardenZoneRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	repo := NewPostgresGardenZoneRepository(db)

	_, err = repo.FindByID(ctx, "non-existent-id")
	require.Error(t, err)
	assert.IsType(t, &entity.NotFoundError{}, err)
}

func TestGardenZoneRepository_FindByGardenID(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create multiple zones
	zoneType1 := entity.ZoneTypeBed
	zoneType2 := entity.ZoneTypeLawn
	zones := []*entity.GardenZone{
		{
			GardenID:        gardenID,
			GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
			ZoneType:        &zoneType1,
		},
		{
			GardenID:        gardenID,
			GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
			ZoneType:        &zoneType2,
		},
	}

	for _, zone := range zones {
		err := repo.Create(ctx, zone)
		require.NoError(t, err)
	}

	// Find all zones
	found, err := repo.FindByGardenID(ctx, gardenID)
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestGardenZoneRepository_Update(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create zone
	zoneType := entity.ZoneTypeBed
	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
		ZoneType:        &zoneType,
	}
	err = repo.Create(ctx, zone)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure updated_at changes

	// Update zone
	newZoneType := entity.ZoneTypeBorder
	zone.ZoneType = &newZoneType
	err = repo.Update(ctx, zone)
	require.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(ctx, zone.ZoneID)
	require.NoError(t, err)
	assert.Equal(t, newZoneType, *found.ZoneType)
}

func TestGardenZoneRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create zone
	zoneType := entity.ZoneTypeBed
	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
		ZoneType:        &zoneType,
	}
	err = repo.Create(ctx, zone)
	require.NoError(t, err)

	// Delete zone
	err = repo.Delete(ctx, zone.ZoneID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(ctx, zone.ZoneID)
	require.Error(t, err)
	assert.IsType(t, &entity.NotFoundError{}, err)
}

func TestGardenZoneRepository_CalculateArea(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create zone
	zoneType := entity.ZoneTypeBed
	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
		ZoneType:        &zoneType,
	}
	err = repo.Create(ctx, zone)
	require.NoError(t, err)

	// Calculate area
	area, err := repo.CalculateArea(ctx, zone.ZoneID)
	require.NoError(t, err)
	assert.Greater(t, area, 0.0)
}

func TestGardenZoneRepository_ValidateZoneWithinGarden(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Test valid zone (within garden)
	err = repo.ValidateZoneWithinGarden(ctx, gardenID, testhelpers.TestGeoJSON.ValidZoneGeometry)
	require.NoError(t, err)

	// Test invalid zone (outside garden boundary)
	outsideZone := `{"type":"Polygon","coordinates":[[[-123.0,38.0],[-122.9,38.0],[-122.9,37.9],[-123.0,37.9],[-123.0,38.0]]]}`
	err = repo.ValidateZoneWithinGarden(ctx, gardenID, outsideZone)
	require.Error(t, err)
}

func TestGardenZoneRepository_CheckZoneOverlaps(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create first zone
	zoneType := entity.ZoneTypeBed
	zone1 := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
		ZoneType:        &zoneType,
	}
	err = repo.Create(ctx, zone1)
	require.NoError(t, err)

	// Test overlap with same geometry
	overlaps, err := repo.CheckZoneOverlaps(ctx, gardenID, testhelpers.TestGeoJSON.ValidZoneGeometry, nil)
	require.NoError(t, err)
	assert.True(t, overlaps)

	// Test overlap excluding first zone (should be false)
	overlaps, err = repo.CheckZoneOverlaps(ctx, gardenID, testhelpers.TestGeoJSON.ValidZoneGeometry, &zone1.ZoneID)
	require.NoError(t, err)
	assert.False(t, overlaps)
}

func TestGardenZoneRepository_CalculateTotalArea(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create multiple zones
	zoneType := entity.ZoneTypeBed
	for i := 0; i < 2; i++ {
		zone := &entity.GardenZone{
			GardenID:        gardenID,
			GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
			ZoneType:        &zoneType,
		}
		err := repo.Create(ctx, zone)
		require.NoError(t, err)
	}

	// Calculate total area
	totalArea, err := repo.CalculateTotalArea(ctx, gardenID)
	require.NoError(t, err)
	assert.Greater(t, totalArea, 0.0)
}

func TestGardenZoneRepository_CountByGardenID(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	userID, gardenID := testhelpers.SeedTestGarden(ctx, db, t)
	require.NotEmpty(t, userID)
	require.NotEmpty(t, gardenID)

	repo := NewPostgresGardenZoneRepository(db)

	// Create zones
	zoneType := entity.ZoneTypeBed
	for i := 0; i < 3; i++ {
		zone := &entity.GardenZone{
			GardenID:        gardenID,
			GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
			ZoneType:        &zoneType,
		}
		err := repo.Create(ctx, zone)
		require.NoError(t, err)
	}

	// Count zones
	count, err := repo.CountByGardenID(ctx, gardenID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}
