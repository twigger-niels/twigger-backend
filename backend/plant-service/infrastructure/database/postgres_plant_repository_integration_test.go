// +build integration

package database

import (
	"context"
	"testing"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
	testhelpers "twigger-backend/backend/plant-service/infrastructure/database/testing"
	"twigger-backend/backend/plant-service/pkg/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgresPlantRepository_FindByID_Integration tests FindByID with real database
func TestPostgresPlantRepository_FindByID_Integration(t *testing.T) {
	// Setup
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(t, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	// Create test plant
	plantID := uuid.New().String()
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Rosa rugosa', NOW())
	`, plantID, speciesID)
	require.NoError(t, err)

	// Insert English common name
	_, err = db.ExecContext(ctx, `
		INSERT INTO plant_common_names (plant_id, language_id, common_name)
		VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', 'Japanese Rose')
	`, plantID)
	require.NoError(t, err)

	t.Run("successful retrieval with English common names", func(t *testing.T) {
		// Use English language UUID from seeded test data
		englishID := "550e8400-e29b-41d4-a716-446655440001"
		plant, err := repo.FindByID(ctx, plantID, englishID, nil)

		require.NoError(t, err)
		assert.NotNil(t, plant)
		assert.Equal(t, plantID, plant.PlantID)
		assert.Equal(t, speciesID, plant.SpeciesID)
		assert.Equal(t, "Rosa rugosa", plant.FullBotanicalName)
		assert.Equal(t, "Rosa", plant.GenusName)
		assert.Equal(t, "rugosa", plant.SpeciesName)
		assert.Equal(t, "Rosaceae", plant.FamilyName)
		assert.Equal(t, types.PlantTypeShrub, plant.PlantType)

		// Verify common names loaded
		require.NotNil(t, plant.CommonNames)
		assert.Contains(t, plant.CommonNames, "Japanese Rose")
	})

	t.Run("plant not found", func(t *testing.T) {
		englishID := "550e8400-e29b-41d4-a716-446655440001"
		nonExistentID := uuid.New().String()
		plant, err := repo.FindByID(ctx, nonExistentID, englishID, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrPlantNotFound)
		assert.Nil(t, plant)
	})

	t.Run("invalid plant ID", func(t *testing.T) {
		plant, err := repo.FindByID(ctx, "not-a-uuid", "550e8400-e29b-41d4-a716-446655440001", nil)

		assert.Error(t, err)
		assert.Nil(t, plant)
	})
}

// TestPostgresPlantRepository_FindByIDs_Integration tests batch retrieval
func TestPostgresPlantRepository_FindByIDs_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(t, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	// Create multiple test plants
	plantIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		plantIDs[i] = uuid.New().String()
		botanicalName := "Rosa test-species-" + string(rune('A'+i))

		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
			VALUES ($1, $2, $3, NOW())
		`, plantIDs[i], speciesID, botanicalName)
		require.NoError(t, err)

		// Add English common name
		commonName := "Test Rose " + string(rune('A'+i))
		_, err = db.ExecContext(ctx, `
			INSERT INTO plant_common_names (plant_id, language_id, common_name)
			VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', $2)
		`, plantIDs[i], commonName)
		require.NoError(t, err)
	}

	t.Run("retrieve multiple plants with batch loading", func(t *testing.T) {
		plants, err := repo.FindByIDs(ctx, plantIDs, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.Len(t, plants, 3)

		// Verify all plants loaded with common names
		for i, plant := range plants {
			assert.Equal(t, plantIDs[i], plant.PlantID)
			assert.NotEmpty(t, plant.CommonNames, "Plant %d should have common names", i)
		}
	})

	t.Run("empty IDs array returns empty result", func(t *testing.T) {
		plants, err := repo.FindByIDs(ctx, []string{}, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.Empty(t, plants)
	})

	t.Run("partial match returns only found plants", func(t *testing.T) {
		mixedIDs := []string{plantIDs[0], uuid.New().String(), plantIDs[1]}
		plants, err := repo.FindByIDs(ctx, mixedIDs, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.Len(t, plants, 2, "Should only return the 2 existing plants")
	})
}

// TestPostgresPlantRepository_Localization_Integration tests localization fallback
func TestPostgresPlantRepository_Localization_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(t, db)
	testhelpers.SeedTestCountries(t, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	plantID := uuid.New().String()
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Solanum lycopersicum', NOW())
	`, plantID, speciesID)
	require.NoError(t, err)

	// Insert localized common names
	// English (global)
	_, err = db.ExecContext(ctx, `
		INSERT INTO plant_common_names (plant_id, language_id, country_id, common_name)
		VALUES
			($1, '550e8400-e29b-41d4-a716-446655440001', NULL, 'Tomato'),
			($1, '550e8400-e29b-41d4-a716-446655440002', NULL, 'Tomate'),
			($1, '550e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440002', 'Jitomate')
	`, plantID)
	require.NoError(t, err)

	t.Run("English global", func(t *testing.T) {
		plant, err := repo.FindByID(ctx, plantID, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.Contains(t, plant.CommonNames, "Tomato")
	})

	t.Run("Spanish global", func(t *testing.T) {
		spanishID := "550e8400-e29b-41d4-a716-446655440002"
		plant, err := repo.FindByID(ctx, plantID, spanishID, nil)

		require.NoError(t, err)
		assert.Contains(t, plant.CommonNames, "Tomate")
	})

	t.Run("Spanish Mexico - country specific", func(t *testing.T) {
		spanishID := "550e8400-e29b-41d4-a716-446655440002"
		countryMX := "650e8400-e29b-41d4-a716-446655440002" // Mexico UUID
		plant, err := repo.FindByID(ctx, plantID, spanishID, &countryMX)

		require.NoError(t, err)
		// Should include country-specific name
		assert.Contains(t, plant.CommonNames, "Jitomate")
	})

	t.Run("language without translations falls back to English", func(t *testing.T) {
		germanID := "550e8400-e29b-41d4-a716-446655440004"
		plant, err := repo.FindByID(ctx, plantID, germanID, nil)

		require.NoError(t, err)
		// German has no translations, should fall back to English per 4-tier fallback chain
		assert.Contains(t, plant.CommonNames, "Tomato")
	})
}

// TestPostgresPlantRepository_Create_Integration tests plant creation
func TestPostgresPlantRepository_Create_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	t.Run("successful creation", func(t *testing.T) {
		newPlant := &entity.Plant{
			PlantID:           uuid.New().String(),
			SpeciesID:         speciesID,
			FamilyName:        "Rosaceae",
			GenusName:         "Rosa",
			SpeciesName:       "canina",
			PlantType:         types.PlantTypeShrub,
			FullBotanicalName: "Rosa canina",
			CreatedAt:         time.Now(),
		}

		err := repo.Create(ctx, newPlant)
		require.NoError(t, err)

		// Verify plant was created
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM plants WHERE plant_id = $1", newPlant.PlantID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("duplicate plant ID fails", func(t *testing.T) {
		plantID := uuid.New().String()

		plant1 := &entity.Plant{
			PlantID:           plantID,
			SpeciesID:         speciesID,
			FamilyName:        "Rosaceae",
			GenusName:         "Rosa",
			SpeciesName:       "test",
			PlantType:         types.PlantTypeShrub,
			FullBotanicalName: "Rosa test",
			CreatedAt:         time.Now(),
		}

		err := repo.Create(ctx, plant1)
		require.NoError(t, err)

		// Try to create duplicate
		plant2 := &entity.Plant{
			PlantID:           plantID,
			SpeciesID:         speciesID,
			FamilyName:        "Rosaceae",
			GenusName:         "Rosa",
			SpeciesName:       "test2",
			PlantType:         types.PlantTypeShrub,
			FullBotanicalName: "Rosa test2",
			CreatedAt:         time.Now(),
		}

		err = repo.Create(ctx, plant2)
		assert.Error(t, err, "Should fail on duplicate plant_id")
	})
}

// TestPostgresPlantRepository_FindByBotanicalName_Integration tests botanical name search
func TestPostgresPlantRepository_FindByBotanicalName_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(t, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	plantID := uuid.New().String()
	botanicalName := "Rosa rugosa"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, $3, NOW())
	`, plantID, speciesID, botanicalName)
	require.NoError(t, err)

	t.Run("find by exact botanical name", func(t *testing.T) {
		plant, err := repo.FindByBotanicalName(ctx, botanicalName, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.NotNil(t, plant)
		assert.Equal(t, plantID, plant.PlantID)
		assert.Equal(t, botanicalName, plant.FullBotanicalName)
	})

	t.Run("case insensitive search", func(t *testing.T) {
		plant, err := repo.FindByBotanicalName(ctx, "rosa RUGOSA", "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.NotNil(t, plant)
		assert.Equal(t, botanicalName, plant.FullBotanicalName)
	})

	t.Run("botanical name not found", func(t *testing.T) {
		plant, err := repo.FindByBotanicalName(ctx, "Nonexistent plantus", "550e8400-e29b-41d4-a716-446655440001", nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrPlantNotFound)
		assert.Nil(t, plant)
	})
}

// TestPostgresPlantRepository_Search_Integration tests search functionality
func TestPostgresPlantRepository_Search_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(t, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	// Create multiple plants for search testing
	testPlants := []struct {
		BotanicalName string
		CommonName    string
	}{
		{"Rosa rugosa", "Japanese Rose"},
		{"Rosa canina", "Dog Rose"},
		{"Rosa multiflora", "Multiflora Rose"},
	}

	for _, tp := range testPlants {
		plantID := uuid.New().String()
		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
			VALUES ($1, $2, $3, NOW())
		`, plantID, speciesID, tp.BotanicalName)
		require.NoError(t, err)

		// Add common name
		_, err = db.ExecContext(ctx, `
			INSERT INTO plant_common_names (plant_id, language_id, common_name)
			VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', $2)
		`, plantID, tp.CommonName)
		require.NoError(t, err)
	}

	t.Run("search by botanical name substring", func(t *testing.T) {
		filter := repository.DefaultSearchFilter()
		result, err := repo.Search(ctx, "Rosa", filter, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, len(result.Plants), 3, "Should find at least 3 Rosa plants")
	})

	// Search by common name - now implemented with CTE to search both botanical and common names
	t.Run("search by common name", func(t *testing.T) {
		filter := repository.DefaultSearchFilter()
		result, err := repo.Search(ctx, "Japanese", filter, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, len(result.Plants), 1, "Should find at least one plant with 'Japanese' in common name")
		// Verify the result includes Rosa rugosa which has 'Japanese Rose' as common name
		found := false
		for _, plant := range result.Plants {
			if plant.FullBotanicalName == "Rosa rugosa" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find Rosa rugosa when searching for 'Japanese'")
	})

	t.Run("search with limit", func(t *testing.T) {
		filter := repository.DefaultSearchFilter()
		filter.Limit = 2

		result, err := repo.Search(ctx, "Rosa", filter, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		assert.LessOrEqual(t, len(result.Plants), 2)
	})

	t.Run("empty search query returns all results", func(t *testing.T) {
		filter := repository.DefaultSearchFilter()
		result, err := repo.Search(ctx, "", filter, "550e8400-e29b-41d4-a716-446655440001", nil)

		require.NoError(t, err)
		// Empty query should return all plants (3 Rosa plants created in setup)
		assert.GreaterOrEqual(t, len(result.Plants), 3)
	})
}

// TestPostgresPlantRepository_Update_Integration tests plant updates
func TestPostgresPlantRepository_Update_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	// Create initial plant
	plantID := uuid.New().String()
	initialName := "Rosa initial"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, $3, NOW())
	`, plantID, speciesID, initialName)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updatedPlant := &entity.Plant{
			PlantID:           plantID,
			SpeciesID:         speciesID,
			FamilyName:        "Rosaceae",
			GenusName:         "Rosa",
			SpeciesName:       "updated",
			PlantType:         types.PlantTypeShrub,
			FullBotanicalName: "Rosa updated",
			UpdatedAt:         time.Now(),
		}

		err := repo.Update(ctx, updatedPlant)
		require.NoError(t, err)

		// Verify update
		var botanicalName string
		err = db.QueryRowContext(ctx, "SELECT full_botanical_name FROM plants WHERE plant_id = $1", plantID).Scan(&botanicalName)
		require.NoError(t, err)
		assert.Equal(t, "Rosa updated", botanicalName)
	})

	t.Run("update non-existent plant fails", func(t *testing.T) {
		nonExistentPlant := &entity.Plant{
			PlantID:           uuid.New().String(),
			SpeciesID:         speciesID,
			FamilyName:        "Rosaceae",
			GenusName:         "Rosa",
			SpeciesName:       "test",
			PlantType:         types.PlantTypeShrub,
			FullBotanicalName: "Rosa test",
			UpdatedAt:         time.Now(),
		}

		err := repo.Update(ctx, nonExistentPlant)
		assert.Error(t, err)
	})
}

// TestPostgresPlantRepository_Delete_Integration tests plant deletion
func TestPostgresPlantRepository_Delete_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	plantID := uuid.New().String()
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Rosa deleteme', NOW())
	`, plantID, speciesID)
	require.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := repo.Delete(ctx, plantID)
		require.NoError(t, err)

		// Verify deletion
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM plants WHERE plant_id = $1", plantID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete non-existent plant", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		err := repo.Delete(ctx, nonExistentID)
		assert.Error(t, err)
	})
}

// TestPostgresPlantRepository_FindByGrowingConditions_Integration tests growing conditions filtering
func TestPostgresPlantRepository_FindByGrowingConditions_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(t, db)
	testhelpers.SeedTestCountries(t, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)

	// Create test plants with different growing conditions
	plantID1 := uuid.New().String()
	plantID2 := uuid.New().String()
	plantID3 := uuid.New().String()

	// Plant 1: Full sun, low water, drought tolerant, hardiness zone 5-7
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Lavandula angustifolia', NOW())
	`, plantID1, speciesID)
	require.NoError(t, err)

	// Plant 2: Part shade, moderate water, not drought tolerant, hardiness zone 6-8
	_, err = db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Hosta plantaginea', NOW())
	`, plantID2, speciesID)
	require.NoError(t, err)

	// Plant 3: Full sun, wet/high water, salt tolerant, hardiness zone 9-11
	_, err = db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Hibiscus rosa-sinensis', NOW())
	`, plantID3, speciesID)
	require.NoError(t, err)

	// Create country_plants entries
	usaID := "650e8400-e29b-41d4-a716-446655440001"
	countryPlant1 := uuid.New().String()
	countryPlant2 := uuid.New().String()
	countryPlant3 := uuid.New().String()

	_, err = db.ExecContext(ctx, `
		INSERT INTO country_plants (country_plant_id, plant_id, country_id, native_status, legal_status)
		VALUES ($1, $2, $3, 'naturalized', 'unrestricted')
	`, countryPlant1, plantID1, usaID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO country_plants (country_plant_id, plant_id, country_id, native_status, legal_status)
		VALUES ($1, $2, $3, 'native', 'unrestricted')
	`, countryPlant2, plantID2, usaID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO country_plants (country_plant_id, plant_id, country_id, native_status, legal_status)
		VALUES ($1, $2, $3, 'introduced', 'unrestricted')
	`, countryPlant3, plantID3, usaID)
	require.NoError(t, err)

	// Create data source
	sourceID := uuid.New().String()
	_, err = db.ExecContext(ctx, `
		INSERT INTO data_sources (source_id, source_name, source_type, reliability_score)
		VALUES ($1, 'Test Source', 'botanical_garden', 0.9)
	`, sourceID)
	require.NoError(t, err)

	// Add growing conditions for plant 1 (Lavender)
	_, err = db.ExecContext(ctx, `
		INSERT INTO growing_conditions_assertions (
			assertion_id, country_plant_id, source_id, confidence,
			hardiness_zones, sun_requirements, water_needs,
			drought_tolerant, salt_tolerant, wind_tolerant,
			flowering_months, ph_preference
		) VALUES (
			$1, $2, $3, 'confirmed',
			'5a,5b,6a,6b,7a,7b',
			ARRAY['full_sun']::sun_requirement[],
			'dry',
			true, false, true,
			ARRAY[6, 7, 8],
			ROW(6.5, 8.0, 7.0)::ph_range
		)
	`, uuid.New().String(), countryPlant1, sourceID)
	require.NoError(t, err)

	// Add growing conditions for plant 2 (Hosta)
	_, err = db.ExecContext(ctx, `
		INSERT INTO growing_conditions_assertions (
			assertion_id, country_plant_id, source_id, confidence,
			hardiness_zones, sun_requirements, water_needs,
			drought_tolerant, salt_tolerant, wind_tolerant,
			flowering_months, ph_preference
		) VALUES (
			$1, $2, $3, 'confirmed',
			'6a,6b,7a,7b,8a,8b',
			ARRAY['partial_shade', 'full_shade']::sun_requirement[],
			'moderate',
			false, false, false,
			ARRAY[7, 8],
			ROW(6.0, 7.5, 6.5)::ph_range
		)
	`, uuid.New().String(), countryPlant2, sourceID)
	require.NoError(t, err)

	// Add growing conditions for plant 3 (Hibiscus)
	_, err = db.ExecContext(ctx, `
		INSERT INTO growing_conditions_assertions (
			assertion_id, country_plant_id, source_id, confidence,
			hardiness_zones, sun_requirements, water_needs,
			drought_tolerant, salt_tolerant, wind_tolerant,
			flowering_months, ph_preference
		) VALUES (
			$1, $2, $3, 'probable',
			'9a,9b,10a,10b,11a,11b',
			ARRAY['full_sun']::sun_requirement[],
			'wet',
			false, true, false,
			ARRAY[5, 6, 7, 8, 9],
			ROW(5.5, 6.5, 6.0)::ph_range
		)
	`, uuid.New().String(), countryPlant3, sourceID)
	require.NoError(t, err)

	t.Run("filter by hardiness zone", func(t *testing.T) {
		zone := "6a"
		filter := &repository.GrowingConditionsFilter{
			HardinessZone: &zone,
			Limit:         10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 2) // Lavender and Hosta both have zone 6a

		plantIDs := []string{plants[0].PlantID, plants[1].PlantID}
		assert.Contains(t, plantIDs, plantID1)
		assert.Contains(t, plantIDs, plantID2)
	})

	t.Run("filter by sun requirements", func(t *testing.T) {
		filter := &repository.GrowingConditionsFilter{
			SunRequirements: []types.SunRequirement{types.SunFullSun},
			Limit:           10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 2) // Lavender and Hibiscus both need full sun

		plantIDs := []string{plants[0].PlantID, plants[1].PlantID}
		assert.Contains(t, plantIDs, plantID1)
		assert.Contains(t, plantIDs, plantID3)
	})

	t.Run("filter by water needs", func(t *testing.T) {
		waterNeeds := types.WaterDry
		filter := &repository.GrowingConditionsFilter{
			WaterNeeds: &waterNeeds,
			Limit:      10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 1)
		assert.Equal(t, plantID1, plants[0].PlantID) // Only Lavender has dry/low water needs
	})

	t.Run("filter by drought tolerant", func(t *testing.T) {
		droughtTolerant := true
		filter := &repository.GrowingConditionsFilter{
			DroughtTolerant: &droughtTolerant,
			Limit:           10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 1)
		assert.Equal(t, plantID1, plants[0].PlantID) // Only Lavender is drought tolerant
	})

	t.Run("filter by salt tolerant", func(t *testing.T) {
		saltTolerant := true
		filter := &repository.GrowingConditionsFilter{
			SaltTolerant: &saltTolerant,
			Limit:        10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 1)
		assert.Equal(t, plantID3, plants[0].PlantID) // Only Hibiscus is salt tolerant
	})

	t.Run("filter by flowering month", func(t *testing.T) {
		month := 7 // July
		filter := &repository.GrowingConditionsFilter{
			FloweringMonth: &month,
			Limit:          10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 3) // All three plants flower in July
	})

	t.Run("filter by pH range", func(t *testing.T) {
		minPH := 6.0
		maxPH := 7.0
		filter := &repository.GrowingConditionsFilter{
			MinPH: &minPH,
			MaxPH: &maxPH,
			Limit: 10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(plants), 2) // At least Lavender and Hosta
	})

	t.Run("combined filters", func(t *testing.T) {
		zone := "6a"
		droughtTolerant := true
		filter := &repository.GrowingConditionsFilter{
			HardinessZone:   &zone,
			DroughtTolerant: &droughtTolerant,
			Limit:           10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 1)
		assert.Equal(t, plantID1, plants[0].PlantID) // Only Lavender matches both
	})

	t.Run("confidence level filter", func(t *testing.T) {
		minConfidence := types.ConfidenceConfirmed
		filter := &repository.GrowingConditionsFilter{
			MinConfidence: &minConfidence,
			Limit:         10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 2) // Lavender and Hosta have "confirmed", Hibiscus has "probable"
	})

	t.Run("no matching plants", func(t *testing.T) {
		zone := "1a" // Very cold zone, none of our test plants
		filter := &repository.GrowingConditionsFilter{
			HardinessZone: &zone,
			Limit:         10,
		}

		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, plants, 0)
	})

	t.Run("cursor pagination", func(t *testing.T) {
		filter := &repository.GrowingConditionsFilter{
			Limit: 2,
		}

		// First page
		plants, err := repo.FindByGrowingConditions(ctx, filter)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(plants), 2)

		if len(plants) == 2 {
			// Second page
			cursor := plants[1].PlantID
			filter.Cursor = &cursor
			filter.Limit = 10

			plantsPage2, err := repo.FindByGrowingConditions(ctx, filter)
			require.NoError(t, err)
			// Should not include plants from first page
			for _, p := range plantsPage2 {
				assert.NotEqual(t, plants[0].PlantID, p.PlantID)
				assert.NotEqual(t, plants[1].PlantID, p.PlantID)
			}
		}
	})
}

// TestPostgresPlantRepository_SearchWithPhysicalCharacteristics_Integration tests physical characteristic filters in Search
func TestPostgresPlantRepository_SearchWithPhysicalCharacteristics_Integration(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer testhelpers.TeardownTestDB(t, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(t, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(t, db)
	englishID := "550e8400-e29b-41d4-a716-446655440001"

	// Create test plants with different physical characteristics
	plantID1 := uuid.New().String() // Tall tree
	plantID2 := uuid.New().String() // Medium shrub
	plantID3 := uuid.New().String() // Small groundcover

	// Plant 1: Tall tree (10-20m), fast growth, evergreen, non-toxic
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Pinus strobus', NOW())
	`, plantID1, speciesID)
	require.NoError(t, err)

	// Plant 2: Medium shrub (1-3m), moderate growth, deciduous, toxic
	_, err = db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Rhododendron catawbiense', NOW())
	`, plantID2, speciesID)
	require.NoError(t, err)

	// Plant 3: Small groundcover (0.1-0.3m), slow growth, evergreen, non-toxic
	_, err = db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Thymus serpyllum', NOW())
	`, plantID3, speciesID)
	require.NoError(t, err)

	// Add physical characteristics for plant 1 (Pine)
	_, err = db.ExecContext(ctx, `
		INSERT INTO physical_characteristics (
			plant_id, mature_height, mature_spread, growth_rate, traits
		) VALUES (
			$1,
			ROW(10.0, 15.0, 20.0)::size_range,
			ROW(5.0, 7.0, 10.0)::size_range,
			'fast',
			'{"evergreen": true, "deciduous": false, "toxic": false}'::jsonb
		)
	`, plantID1)
	require.NoError(t, err)

	// Add physical characteristics for plant 2 (Rhododendron)
	_, err = db.ExecContext(ctx, `
		INSERT INTO physical_characteristics (
			plant_id, mature_height, mature_spread, growth_rate, traits
		) VALUES (
			$1,
			ROW(1.0, 2.0, 3.0)::size_range,
			ROW(1.0, 1.5, 2.0)::size_range,
			'moderate',
			'{"evergreen": false, "deciduous": true, "toxic": true}'::jsonb
		)
	`, plantID2)
	require.NoError(t, err)

	// Add physical characteristics for plant 3 (Thyme)
	_, err = db.ExecContext(ctx, `
		INSERT INTO physical_characteristics (
			plant_id, mature_height, mature_spread, growth_rate, traits
		) VALUES (
			$1,
			ROW(0.1, 0.2, 0.3)::size_range,
			ROW(0.3, 0.5, 0.7)::size_range,
			'slow',
			'{"evergreen": true, "deciduous": false, "toxic": false}'::jsonb
		)
	`, plantID3)
	require.NoError(t, err)

	t.Run("filter by minimum height", func(t *testing.T) {
		minHeight := 5.0 // Should match pine only
		filter := &repository.SearchFilter{
			MinHeight: &minHeight,
			Limit:     10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 1)
		assert.Equal(t, plantID1, result.Plants[0].PlantID)
	})

	t.Run("filter by maximum height", func(t *testing.T) {
		maxHeight := 1.0 // Should match thyme only
		filter := &repository.SearchFilter{
			MaxHeight: &maxHeight,
			Limit:     10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 1)
		assert.Equal(t, plantID3, result.Plants[0].PlantID)
	})

	t.Run("filter by height range", func(t *testing.T) {
		minHeight := 0.5
		maxHeight := 5.0 // Should match rhododendron only
		filter := &repository.SearchFilter{
			MinHeight: &minHeight,
			MaxHeight: &maxHeight,
			Limit:     10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 1)
		assert.Equal(t, plantID2, result.Plants[0].PlantID)
	})

	t.Run("filter by growth rate", func(t *testing.T) {
		growthRate := types.GrowthFast
		filter := &repository.SearchFilter{
			GrowthRate: &growthRate,
			Limit:      10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 1)
		assert.Equal(t, plantID1, result.Plants[0].PlantID)
	})

	t.Run("filter by evergreen", func(t *testing.T) {
		evergreen := true
		filter := &repository.SearchFilter{
			Evergreen: &evergreen,
			Limit:     10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 2) // Pine and Thyme

		plantIDs := []string{result.Plants[0].PlantID, result.Plants[1].PlantID}
		assert.Contains(t, plantIDs, plantID1)
		assert.Contains(t, plantIDs, plantID3)
	})

	t.Run("filter by deciduous", func(t *testing.T) {
		deciduous := true
		filter := &repository.SearchFilter{
			Deciduous: &deciduous,
			Limit:     10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 1)
		assert.Equal(t, plantID2, result.Plants[0].PlantID)
	})

	t.Run("filter by toxic", func(t *testing.T) {
		toxic := true
		filter := &repository.SearchFilter{
			Toxic: &toxic,
			Limit: 10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 1)
		assert.Equal(t, plantID2, result.Plants[0].PlantID)
	})

	t.Run("filter by non-toxic", func(t *testing.T) {
		toxic := false
		filter := &repository.SearchFilter{
			Toxic: &toxic,
			Limit: 10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 2) // Pine and Thyme

		plantIDs := []string{result.Plants[0].PlantID, result.Plants[1].PlantID}
		assert.Contains(t, plantIDs, plantID1)
		assert.Contains(t, plantIDs, plantID3)
	})

	t.Run("combined physical filters", func(t *testing.T) {
		maxHeight := 1.0
		evergreen := true
		filter := &repository.SearchFilter{
			MaxHeight: &maxHeight,
			Evergreen: &evergreen,
			Limit:     10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)
		assert.Len(t, result.Plants, 1)
		assert.Equal(t, plantID3, result.Plants[0].PlantID) // Only thyme matches both
	})

	t.Run("no physical characteristics data", func(t *testing.T) {
		// Create a plant without physical characteristics
		plantID4 := uuid.New().String()
		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
			VALUES ($1, $2, 'Acer rubrum', NOW())
		`, plantID4, speciesID)
		require.NoError(t, err)

		// Search with physical filter should not include this plant
		evergreen := true
		filter := &repository.SearchFilter{
			Evergreen: &evergreen,
			Limit:     10,
		}

		result, err := repo.Search(ctx, "", filter, englishID, nil)
		require.NoError(t, err)

		// Should still get pine and thyme, not the new plant
		for _, p := range result.Plants {
			assert.NotEqual(t, plantID4, p.PlantID)
		}
	})
}
