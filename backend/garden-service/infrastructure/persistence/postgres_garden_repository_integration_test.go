// +build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twigger-backend/backend/garden-service/domain/entity"
	testhelpers "twigger-backend/backend/garden-service/infrastructure/database/testing"
)

func setupGardenTest(t *testing.T) (*PostgresGardenRepository, context.Context, func()) {
	ctx := context.Background()
	db := testhelpers.GetTestDB(t)

	// Clean database and create schema
	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err, "Failed to clean database")

	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err, "Failed to create schema")

	// Seed test users
	err = testhelpers.SeedTestUsers(ctx, db, t)
	require.NoError(t, err, "Failed to seed users")

	// Seed test climate zones
	err = testhelpers.SeedTestClimateZones(ctx, db, t)
	require.NoError(t, err, "Failed to seed climate zones")

	repo := NewPostgresGardenRepository(db)

	// Return cleanup function
	cleanup := func() {
		db.Close()
	}

	return repo.(*PostgresGardenRepository), ctx, cleanup
}

func TestGardenRepository_Create(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001" // From SeedTestUsers
	gardenType := entity.GardenTypeVegetable
	aspect := entity.AspectS
	elevation := 100.5

	garden := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "Test Vegetable Garden",
		BoundaryGeoJSON: &testhelpers.TestGeoJSON.ValidGardenBoundary,
		LocationGeoJSON: &testhelpers.TestGeoJSON.ValidGardenLocation,
		ElevationM:      &elevation,
		Aspect:          &aspect,
		GardenType:      &gardenType,
	}

	err := repo.Create(ctx, garden)
	require.NoError(t, err)

	// Verify timestamps were set
	assert.False(t, garden.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, garden.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.Equal(t, garden.CreatedAt, garden.UpdatedAt, "CreatedAt and UpdatedAt should be equal on create")
}

func TestGardenRepository_Create_InvalidGeoJSON(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	garden := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "Invalid Garden",
		BoundaryGeoJSON: &testhelpers.TestGeoJSON.InvalidPolygon, // Unclosed polygon
	}

	err := repo.Create(ctx, garden)
	require.Error(t, err, "Should reject invalid GeoJSON")

	// Check error type
	_, isSpatialError := err.(*entity.SpatialError)
	assert.True(t, isSpatialError, "Should return SpatialError")
}

func TestGardenRepository_FindByID(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Create garden
	original := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "Findable Garden",
		BoundaryGeoJSON: &testhelpers.TestGeoJSON.ValidGardenBoundary,
		LocationGeoJSON: &testhelpers.TestGeoJSON.ValidGardenLocation,
	}

	err := repo.Create(ctx, original)
	require.NoError(t, err)

	// Find by ID
	found, err := repo.FindByID(ctx, original.GardenID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify fields
	assert.Equal(t, original.GardenID, found.GardenID)
	assert.Equal(t, original.UserID, found.UserID)
	assert.Equal(t, original.GardenName, found.GardenName)
	assert.NotNil(t, found.BoundaryGeoJSON, "Boundary should be returned")
	assert.NotNil(t, found.LocationGeoJSON, "Location should be returned")
}

func TestGardenRepository_FindByID_NotFound(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	nonExistentID := uuid.New().String()

	garden, err := repo.FindByID(ctx, nonExistentID)
	require.Error(t, err)
	assert.Nil(t, garden)

	// Check error type
	_, isNotFoundError := err.(*entity.NotFoundError)
	assert.True(t, isNotFoundError, "Should return NotFoundError")
}

func TestGardenRepository_FindByUserID(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Create multiple gardens
	for i := 0; i < 3; i++ {
		garden := &entity.Garden{
			GardenID:        uuid.New().String(),
			UserID:          userID,
			GardenName:      "Garden " + string(rune('A'+i)),
			LocationGeoJSON: &testhelpers.TestGeoJSON.ValidGardenLocation,
		}
		err := repo.Create(ctx, garden)
		require.NoError(t, err)

		// Sleep to ensure different created_at timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Find by user ID
	gardens, err := repo.FindByUserID(ctx, userID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, gardens, 3, "Should find all 3 gardens")

	// Verify ordered by created_at DESC (most recent first)
	assert.Equal(t, "Garden C", gardens[0].GardenName)
	assert.Equal(t, "Garden A", gardens[2].GardenName)
}

func TestGardenRepository_FindByUserID_Pagination(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Create 5 gardens
	for i := 0; i < 5; i++ {
		garden := &entity.Garden{
			GardenID:   uuid.New().String(),
			UserID:     userID,
			GardenName: "Garden " + string(rune('1'+i)),
		}
		err := repo.Create(ctx, garden)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Page 1: Limit 2, Offset 0
	page1, err := repo.FindByUserID(ctx, userID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2, "Page 1 should have 2 gardens")

	// Page 2: Limit 2, Offset 2
	page2, err := repo.FindByUserID(ctx, userID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2, "Page 2 should have 2 gardens")

	// Page 3: Limit 2, Offset 4
	page3, err := repo.FindByUserID(ctx, userID, 2, 4)
	require.NoError(t, err)
	assert.Len(t, page3, 1, "Page 3 should have 1 garden")

	// Verify no overlap
	assert.NotEqual(t, page1[0].GardenID, page2[0].GardenID)
	assert.NotEqual(t, page2[0].GardenID, page3[0].GardenID)
}

func TestGardenRepository_Update(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Create garden
	garden := &entity.Garden{
		GardenID:   uuid.New().String(),
		UserID:     userID,
		GardenName: "Original Name",
	}
	err := repo.Create(ctx, garden)
	require.NoError(t, err)

	originalUpdatedAt := garden.UpdatedAt

	// Wait to ensure timestamp difference
	time.Sleep(100 * time.Millisecond)

	// Update garden
	garden.GardenName = "Updated Name"
	gardenType := entity.GardenTypeOrnamental
	garden.GardenType = &gardenType

	err = repo.Update(ctx, garden)
	require.NoError(t, err)

	// Verify updated_at changed
	assert.True(t, garden.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be newer")

	// Fetch and verify
	updated, err := repo.FindByID(ctx, garden.GardenID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.GardenName)
	assert.Equal(t, entity.GardenTypeOrnamental, *updated.GardenType)
}

func TestGardenRepository_Delete(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Create garden
	garden := &entity.Garden{
		GardenID:   uuid.New().String(),
		UserID:     userID,
		GardenName: "Deletable Garden",
	}
	err := repo.Create(ctx, garden)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, garden.GardenID)
	require.NoError(t, err)

	// Verify deleted
	found, err := repo.FindByID(ctx, garden.GardenID)
	require.Error(t, err)
	assert.Nil(t, found)

	// Check error type
	_, isNotFoundError := err.(*entity.NotFoundError)
	assert.True(t, isNotFoundError, "Should return NotFoundError")
}

func TestGardenRepository_FindByLocation(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Garden 1: San Francisco (37.7749, -122.4194)
	sfLocation := testhelpers.TestGeoJSON.ValidGardenLocation
	garden1 := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "SF Garden",
		LocationGeoJSON: &sfLocation,
	}
	err := repo.Create(ctx, garden1)
	require.NoError(t, err)

	// Garden 2: ~1.5km away (37.7849, -122.4094)
	nearbyLocation := `{"type":"Point","coordinates":[-122.4094, 37.7849]}`
	garden2 := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "Nearby Garden",
		LocationGeoJSON: &nearbyLocation,
	}
	err = repo.Create(ctx, garden2)
	require.NoError(t, err)

	// Garden 3: New York (40.7128, -74.0060) - far away
	nyLocation := `{"type":"Point","coordinates":[-74.0060, 40.7128]}`
	garden3 := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "NY Garden",
		LocationGeoJSON: &nyLocation,
	}
	err = repo.Create(ctx, garden3)
	require.NoError(t, err)

	// Search within 5km of SF
	gardens, err := repo.FindByLocation(ctx, 37.7749, -122.4194, 5.0)
	require.NoError(t, err)
	assert.Len(t, gardens, 2, "Should find 2 gardens within 5km")

	// Verify correct gardens found
	gardenNames := []string{gardens[0].GardenName, gardens[1].GardenName}
	assert.Contains(t, gardenNames, "SF Garden")
	assert.Contains(t, gardenNames, "Nearby Garden")
	assert.NotContains(t, gardenNames, "NY Garden")
}

func TestGardenRepository_CalculateArea(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	garden := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "Area Test Garden",
		BoundaryGeoJSON: &testhelpers.TestGeoJSON.ValidGardenBoundary,
	}
	err := repo.Create(ctx, garden)
	require.NoError(t, err)

	// Calculate area
	area, err := repo.CalculateArea(ctx, garden.GardenID)
	require.NoError(t, err)
	assert.Greater(t, area, 0.0, "Area should be positive")

	// Area should be approximately correct for a ~100m x ~100m rectangle
	// (coordinates span ~0.001 degrees lat/lng which is ~100m at SF latitude)
	assert.InDelta(t, 10000, area, 5000, "Area should be approximately 10,000 m² ±5,000 m²")
}

func TestGardenRepository_DetectHardinessZone(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Create garden within SF climate zone (10a)
	garden := &entity.Garden{
		GardenID:        uuid.New().String(),
		UserID:          userID,
		GardenName:      "SF Garden",
		BoundaryGeoJSON: &testhelpers.TestGeoJSON.ValidGardenBoundary, // SF coordinates
	}
	err := repo.Create(ctx, garden)
	require.NoError(t, err)

	// Detect hardiness zone
	zone, err := repo.DetectHardinessZone(ctx, garden.GardenID)
	require.NoError(t, err)
	assert.Equal(t, "10a", zone, "Should detect USDA zone 10a for SF")
}

func TestGardenRepository_ValidateBoundary(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	// Valid boundary
	err := repo.ValidateBoundary(ctx, testhelpers.TestGeoJSON.ValidGardenBoundary)
	assert.NoError(t, err, "Valid boundary should pass")

	// Invalid boundary (unclosed ring)
	err = repo.ValidateBoundary(ctx, testhelpers.TestGeoJSON.InvalidPolygon)
	assert.Error(t, err, "Invalid boundary should fail")

	// Check error type
	_, isSpatialError := err.(*entity.SpatialError)
	assert.True(t, isSpatialError, "Should return SpatialError")
}

func TestGardenRepository_CountByUserID(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Initially 0
	count, err := repo.CountByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create 3 gardens
	for i := 0; i < 3; i++ {
		garden := &entity.Garden{
			GardenID:   uuid.New().String(),
			UserID:     userID,
			GardenName: "Garden " + string(rune('1'+i)),
		}
		err := repo.Create(ctx, garden)
		require.NoError(t, err)
	}

	// Should count 3
	count, err = repo.CountByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestGardenRepository_GetTotalArea(t *testing.T) {
	repo, ctx, cleanup := setupGardenTest(t)
	defer cleanup()

	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Create 2 gardens with boundaries
	for i := 0; i < 2; i++ {
		garden := &entity.Garden{
			GardenID:        uuid.New().String(),
			UserID:          userID,
			GardenName:      "Garden " + string(rune('1'+i)),
			BoundaryGeoJSON: &testhelpers.TestGeoJSON.ValidGardenBoundary,
		}
		err := repo.Create(ctx, garden)
		require.NoError(t, err)
	}

	// Get total area
	totalArea, err := repo.GetTotalArea(ctx, userID)
	require.NoError(t, err)
	assert.Greater(t, totalArea, 0.0, "Total area should be positive")

	// Should be approximately 2x the individual area
	assert.InDelta(t, 20000, totalArea, 10000, "Total area should be ~20,000 m² ±10,000 m²")
}
