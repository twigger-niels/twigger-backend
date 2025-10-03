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

func TestGardenPlantRepository_Create(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	healthStatus := entity.HealthStatusHealthy
	quantity := 1
	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		HealthStatus:    &healthStatus,
		Quantity:        &quantity,
	}

	err = repo.Create(ctx, gardenPlant)
	require.NoError(t, err)
	assert.NotEmpty(t, gardenPlant.GardenPlantID)
	assert.False(t, gardenPlant.CreatedAt.IsZero())
	assert.False(t, gardenPlant.UpdatedAt.IsZero())
	assert.False(t, gardenPlant.PlantedAt.IsZero())
}

func TestGardenPlantRepository_Create_InvalidGeoJSON(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.InvalidPolygon,
	}

	err = repo.Create(ctx, gardenPlant)
	require.Error(t, err)
}

func TestGardenPlantRepository_FindByID(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create garden plant
	healthStatus := entity.HealthStatusHealthy
	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		HealthStatus:    &healthStatus,
	}
	err = repo.Create(ctx, gardenPlant)
	require.NoError(t, err)

	// Find garden plant
	found, err := repo.FindByID(ctx, gardenPlant.GardenPlantID)
	require.NoError(t, err)
	assert.Equal(t, gardenPlant.GardenPlantID, found.GardenPlantID)
	assert.Equal(t, gardenPlant.GardenID, found.GardenID)
	assert.Equal(t, gardenPlant.PlantID, found.PlantID)
	assert.NotEmpty(t, found.LocationGeoJSON)
	assert.Equal(t, healthStatus, *found.HealthStatus)
}

func TestGardenPlantRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	repo := NewPostgresGardenPlantRepository(db)

	_, err = repo.FindByID(ctx, "non-existent-id")
	require.Error(t, err)
	assert.IsType(t, &entity.NotFoundError{}, err)
}

func TestGardenPlantRepository_FindByGardenID(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create multiple garden plants
	for i := 0; i < 3; i++ {
		gardenPlant := &entity.GardenPlant{
			GardenID:        gardenID,
			PlantID:         plantID,
			LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		}
		err := repo.Create(ctx, gardenPlant)
		require.NoError(t, err)
	}

	// Find all plants
	found, err := repo.FindByGardenID(ctx, gardenID)
	require.NoError(t, err)
	assert.Len(t, found, 3)
}

func TestGardenPlantRepository_Update(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create garden plant
	healthStatus := entity.HealthStatusHealthy
	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		HealthStatus:    &healthStatus,
	}
	err = repo.Create(ctx, gardenPlant)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Update health status
	newStatus := entity.HealthStatusThriving
	gardenPlant.HealthStatus = &newStatus
	err = repo.Update(ctx, gardenPlant)
	require.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(ctx, gardenPlant.GardenPlantID)
	require.NoError(t, err)
	assert.Equal(t, newStatus, *found.HealthStatus)
}

func TestGardenPlantRepository_Delete(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create garden plant
	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, gardenPlant)
	require.NoError(t, err)

	// Delete garden plant
	err = repo.Delete(ctx, gardenPlant.GardenPlantID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(ctx, gardenPlant.GardenPlantID)
	require.Error(t, err)
	assert.IsType(t, &entity.NotFoundError{}, err)
}

func TestGardenPlantRepository_CheckPlantSpacing(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create garden plant
	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, gardenPlant)
	require.NoError(t, err)

	// Check spacing - should find the existing plant (within 1km)
	nearby, err := repo.CheckPlantSpacing(ctx, gardenID, testhelpers.TestGeoJSON.ValidPlantLocation, 1000)
	require.NoError(t, err)
	assert.Len(t, nearby, 1)
	assert.Equal(t, gardenPlant.GardenPlantID, nearby[0].GardenPlantID)

	// Check spacing with tight radius - should find nothing
	nearby, err = repo.CheckPlantSpacing(ctx, gardenID, testhelpers.TestGeoJSON.ValidPlantLocation, 0.01)
	require.NoError(t, err)
	assert.Len(t, nearby, 0) // No plants within 1cm of exact location (excluding itself)
}

func TestGardenPlantRepository_FindInZone(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	// Create a zone
	zoneRepo := NewPostgresGardenZoneRepository(db)
	zoneType := entity.ZoneTypeBed
	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
		ZoneType:        &zoneType,
	}
	err = zoneRepo.Create(ctx, zone)
	require.NoError(t, err)

	repo := NewPostgresGardenPlantRepository(db)

	// Create plant within zone
	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		ZoneID:          &zone.ZoneID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, gardenPlant)
	require.NoError(t, err)

	// Find plants in zone
	found, err := repo.FindInZone(ctx, zone.ZoneID)
	require.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, gardenPlant.GardenPlantID, found[0].GardenPlantID)
}

func TestGardenPlantRepository_ValidatePlantLocation(t *testing.T) {
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

	repo := NewPostgresGardenPlantRepository(db)

	// Test valid location (within garden)
	err = repo.ValidatePlantLocation(ctx, gardenID, testhelpers.TestGeoJSON.ValidPlantLocation)
	require.NoError(t, err)

	// Test invalid location (outside garden boundary)
	outsideLocation := `{"type":"Point","coordinates":[-123.0,38.0]}`
	err = repo.ValidatePlantLocation(ctx, gardenID, outsideLocation)
	require.Error(t, err)
}

func TestGardenPlantRepository_FindByHealthStatus(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create plants with different health statuses
	healthyStatus := entity.HealthStatusHealthy
	healthyPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		HealthStatus:    &healthyStatus,
	}
	err = repo.Create(ctx, healthyPlant)
	require.NoError(t, err)

	strugglingStatus := entity.HealthStatusStruggling
	strugglingPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		HealthStatus:    &strugglingStatus,
	}
	err = repo.Create(ctx, strugglingPlant)
	require.NoError(t, err)

	// Find only healthy plants
	healthy, err := repo.FindByHealthStatus(ctx, gardenID, entity.HealthStatusHealthy)
	require.NoError(t, err)
	assert.Len(t, healthy, 1)
	assert.Equal(t, healthyPlant.GardenPlantID, healthy[0].GardenPlantID)
}

func TestGardenPlantRepository_FindActivePlants(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create active plant (no removed_date)
	activePlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, activePlant)
	require.NoError(t, err)

	// Create removed plant
	removedDate := time.Now().Add(-24 * time.Hour)
	removedPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		RemovedDate:     &removedDate,
	}
	err = repo.Create(ctx, removedPlant)
	require.NoError(t, err)

	// Find active plants
	active, err := repo.FindActivePlants(ctx, gardenID)
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, activePlant.GardenPlantID, active[0].GardenPlantID)
}

func TestGardenPlantRepository_BulkCreate(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create multiple plants
	plants := []*entity.GardenPlant{
		{
			GardenID:        gardenID,
			PlantID:         plantID,
			LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		},
		{
			GardenID:        gardenID,
			PlantID:         plantID,
			LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		},
		{
			GardenID:        gardenID,
			PlantID:         plantID,
			LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		},
	}

	err = repo.BulkCreate(ctx, plants)
	require.NoError(t, err)

	// Verify all plants were created
	for _, plant := range plants {
		assert.NotEmpty(t, plant.GardenPlantID)
	}

	// Verify count
	found, err := repo.FindByGardenID(ctx, gardenID)
	require.NoError(t, err)
	assert.Len(t, found, 3)
}

func TestGardenPlantRepository_CountByGardenID(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create plants
	for i := 0; i < 5; i++ {
		gardenPlant := &entity.GardenPlant{
			GardenID:        gardenID,
			PlantID:         plantID,
			LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		}
		err := repo.Create(ctx, gardenPlant)
		require.NoError(t, err)
	}

	// Count plants
	count, err := repo.CountByGardenID(ctx, gardenID)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestGardenPlantRepository_FindByIDs(t *testing.T) {
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

	plantID := testhelpers.SeedTestPlant(ctx, db, t)
	require.NotEmpty(t, plantID)

	repo := NewPostgresGardenPlantRepository(db)

	// Create plants
	var ids []string
	for i := 0; i < 3; i++ {
		gardenPlant := &entity.GardenPlant{
			GardenID:        gardenID,
			PlantID:         plantID,
			LocationGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		}
		err := repo.Create(ctx, gardenPlant)
		require.NoError(t, err)
		ids = append(ids, gardenPlant.GardenPlantID)
	}

	// Find by IDs
	found, err := repo.FindByIDs(ctx, ids)
	require.NoError(t, err)
	assert.Len(t, found, 3)
}
