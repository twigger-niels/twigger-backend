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

func TestGardenFeatureRepository_Create(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	featureType := entity.FeatureTypeTree
	featureName := "Oak Tree"
	heightM := 15.0
	canopyDiameterM := 8.0
	deciduous := true

	feature := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     featureType,
		FeatureName:     &featureName,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation, // Point geometry
		HeightM:         &heightM,
		CanopyDiameterM: &canopyDiameterM,
		Deciduous:       &deciduous,
	}

	err = repo.Create(ctx, feature)
	require.NoError(t, err)
	assert.NotEmpty(t, feature.FeatureID)
	assert.False(t, feature.CreatedAt.IsZero())
}

func TestGardenFeatureRepository_FindByID(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create feature
	featureType := entity.FeatureTypeTree
	featureName := "Maple Tree"
	feature := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     featureType,
		FeatureName:     &featureName,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, feature)
	require.NoError(t, err)

	// Find feature
	found, err := repo.FindByID(ctx, feature.FeatureID)
	require.NoError(t, err)
	assert.Equal(t, feature.FeatureID, found.FeatureID)
	assert.Equal(t, feature.GardenID, found.GardenID)
	assert.Equal(t, featureType, found.FeatureType)
	assert.Equal(t, featureName, *found.FeatureName)
	assert.NotEmpty(t, found.GeometryGeoJSON)
}

func TestGardenFeatureRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testhelpers.SetupTestDB(t)
	defer db.Close()

	err := testhelpers.CleanDatabase(ctx, db, t)
	require.NoError(t, err)
	err = testhelpers.CreateTestSchema(ctx, db, t)
	require.NoError(t, err)

	repo := NewPostgresGardenFeatureRepository(db)

	_, err = repo.FindByID(ctx, "non-existent-id")
	require.Error(t, err)
	assert.IsType(t, &entity.NotFoundError{}, err)
}

func TestGardenFeatureRepository_FindByGardenID(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create multiple features
	features := []entity.FeatureType{entity.FeatureTypeTree, entity.FeatureTypeBuilding}
	for _, ft := range features {
		feature := &entity.GardenFeature{
			GardenID:        gardenID,
			FeatureType:     ft,
			GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		}
		err := repo.Create(ctx, feature)
		require.NoError(t, err)
	}

	// Find all features
	found, err := repo.FindByGardenID(ctx, gardenID)
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestGardenFeatureRepository_FindByType(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create features of different types
	treeFeature := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     entity.FeatureTypeTree,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, treeFeature)
	require.NoError(t, err)

	buildingFeature := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     entity.FeatureTypeBuilding,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry, // Polygon
	}
	err = repo.Create(ctx, buildingFeature)
	require.NoError(t, err)

	// Find only trees
	trees, err := repo.FindByType(ctx, gardenID, entity.FeatureTypeTree)
	require.NoError(t, err)
	assert.Len(t, trees, 1)
	assert.Equal(t, entity.FeatureTypeTree, trees[0].FeatureType)
}

func TestGardenFeatureRepository_Update(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create feature
	featureType := entity.FeatureTypeTree
	oldName := "Young Oak"
	heightM := 5.0
	feature := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     featureType,
		FeatureName:     &oldName,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		HeightM:         &heightM,
	}
	err = repo.Create(ctx, feature)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Update feature
	newName := "Mature Oak"
	newHeight := 15.0
	feature.FeatureName = &newName
	feature.HeightM = &newHeight
	err = repo.Update(ctx, feature)
	require.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(ctx, feature.FeatureID)
	require.NoError(t, err)
	assert.Equal(t, newName, *found.FeatureName)
	assert.Equal(t, newHeight, *found.HeightM)
}

func TestGardenFeatureRepository_Delete(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create feature
	feature := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     entity.FeatureTypeTree,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, feature)
	require.NoError(t, err)

	// Delete feature
	err = repo.Delete(ctx, feature.FeatureID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(ctx, feature.FeatureID)
	require.Error(t, err)
	assert.IsType(t, &entity.NotFoundError{}, err)
}

func TestGardenFeatureRepository_FindFeaturesWithHeight(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create feature with height
	heightM := 10.0
	featureWithHeight := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     entity.FeatureTypeTree,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		HeightM:         &heightM,
	}
	err = repo.Create(ctx, featureWithHeight)
	require.NoError(t, err)

	// Create feature without height
	featureWithoutHeight := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     entity.FeatureTypeFence,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, featureWithoutHeight)
	require.NoError(t, err)

	// Find features with height
	found, err := repo.FindFeaturesWithHeight(ctx, gardenID)
	require.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, heightM, *found[0].HeightM)
}

func TestGardenFeatureRepository_FindTreesInGarden(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create tree
	tree := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     entity.FeatureTypeTree,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
	}
	err = repo.Create(ctx, tree)
	require.NoError(t, err)

	// Create building
	building := &entity.GardenFeature{
		GardenID:        gardenID,
		FeatureType:     entity.FeatureTypeBuilding,
		GeometryGeoJSON: testhelpers.TestGeoJSON.ValidZoneGeometry,
	}
	err = repo.Create(ctx, building)
	require.NoError(t, err)

	// Find only trees
	trees, err := repo.FindTreesInGarden(ctx, gardenID)
	require.NoError(t, err)
	assert.Len(t, trees, 1)
	assert.Equal(t, entity.FeatureTypeTree, trees[0].FeatureType)
}

func TestGardenFeatureRepository_CountByGardenID(t *testing.T) {
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

	repo := NewPostgresGardenFeatureRepository(db)

	// Create features
	for i := 0; i < 3; i++ {
		feature := &entity.GardenFeature{
			GardenID:        gardenID,
			FeatureType:     entity.FeatureTypeTree,
			GeometryGeoJSON: testhelpers.TestGeoJSON.ValidPlantLocation,
		}
		err := repo.Create(ctx, feature)
		require.NoError(t, err)
	}

	// Count features
	count, err := repo.CountByGardenID(ctx, gardenID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}
