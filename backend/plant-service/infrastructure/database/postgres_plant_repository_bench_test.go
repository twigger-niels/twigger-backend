// +build integration

package database

import (
	"context"
	"fmt"
	"testing"

	"twigger-backend/backend/plant-service/domain/repository"
	testhelpers "twigger-backend/backend/plant-service/infrastructure/database/testing"

	"github.com/google/uuid"
)

// BenchmarkPlantRepository_FindByIDs_BatchLoading benchmarks batch loading performance
func BenchmarkPlantRepository_FindByIDs_BatchLoading(t *testing.B) {
	db := testhelpers.SetupTestDB(&testing.T{})
	defer testhelpers.TeardownTestDB(&testing.T{}, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(&testing.T{}, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(&testing.T{}, db)

	// Create 100 test plants with common names
	plantIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		plantIDs[i] = uuid.New().String()
		botanicalName := fmt.Sprintf("Rosa benchmark-species-%d", i)

		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
			VALUES ($1, $2, $3, NOW())
		`, plantIDs[i], speciesID, botanicalName)
		if err != nil {
			t.Fatalf("failed to seed plant %d: %v", i, err)
		}

		// Add 3 common names per plant (simulate real-world scenario)
		commonNames := []string{
			fmt.Sprintf("Common Rose %d", i),
			fmt.Sprintf("Garden Rose %d", i),
			fmt.Sprintf("Test Rose %d", i),
		}

		for _, name := range commonNames {
			_, err = db.ExecContext(ctx, `
				INSERT INTO plant_common_names (plant_id, language_id, common_name)
				VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', $2)
			`, plantIDs[i], name)
			if err != nil {
				t.Fatalf("failed to seed common name: %v", err)
			}
		}
	}

	// Benchmark different batch sizes
	benchmarks := []struct {
		name      string
		batchSize int
	}{
		{"BatchSize10", 10},
		{"BatchSize20", 20},
		{"BatchSize50", 50},
		{"BatchSize100", 100},
	}

	for _, bm := range benchmarks {
		t.Run(bm.name, func(b *testing.B) {
			batchIDs := plantIDs[:bm.batchSize]

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := repo.FindByIDs(ctx, batchIDs, "550e8400-e29b-41d4-a716-446655440001", nil)
				if err != nil {
					b.Fatalf("FindByIDs failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPlantRepository_FindByID_SingleVsBatch compares single vs batch loading
func BenchmarkPlantRepository_FindByID_SingleVsBatch(t *testing.B) {
	db := testhelpers.SetupTestDB(&testing.T{})
	defer testhelpers.TeardownTestDB(&testing.T{}, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(&testing.T{}, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(&testing.T{}, db)

	// Create 20 test plants
	plantIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		plantIDs[i] = uuid.New().String()
		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
			VALUES ($1, $2, $3, NOW())
		`, plantIDs[i], speciesID, fmt.Sprintf("Rosa bench-%d", i))
		if err != nil {
			t.Fatalf("failed to seed plant: %v", err)
		}

		// Add common name
		_, err = db.ExecContext(ctx, `
			INSERT INTO plant_common_names (plant_id, language_id, common_name)
			VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', $2)
		`, plantIDs[i], fmt.Sprintf("Rose %d", i))
		if err != nil {
			t.Fatalf("failed to seed common name: %v", err)
		}
	}

	t.Run("N+1_Problem_FindByID_Loop", func(b *testing.B) {
		// Simulate N+1 query problem (old approach)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, id := range plantIDs {
				_, err := repo.FindByID(ctx, id, "550e8400-e29b-41d4-a716-446655440001", nil)
				if err != nil {
					b.Fatalf("FindByID failed: %v", err)
				}
			}
		}
	})

	t.Run("BatchLoading_FindByIDs", func(b *testing.B) {
		// Batch loading approach (new approach)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.FindByIDs(ctx, plantIDs, "550e8400-e29b-41d4-a716-446655440001", nil)
			if err != nil {
				b.Fatalf("FindByIDs failed: %v", err)
			}
		}
	})
}

// BenchmarkPlantRepository_Search benchmarks search performance
func BenchmarkPlantRepository_Search(t *testing.B) {
	db := testhelpers.SetupTestDB(&testing.T{})
	defer testhelpers.TeardownTestDB(&testing.T{}, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(&testing.T{}, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(&testing.T{}, db)

	// Create 200 plants with varying names
	genera := []string{"Rosa", "Prunus", "Malus", "Pyrus", "Crataegus"}
	for i := 0; i < 200; i++ {
		plantID := uuid.New().String()
		genus := genera[i%len(genera)]
		botanicalName := fmt.Sprintf("%s species-%d", genus, i)

		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
			VALUES ($1, $2, $3, NOW())
		`, plantID, speciesID, botanicalName)
		if err != nil {
			t.Fatalf("failed to seed plant: %v", err)
		}

		// Add common names
		_, err = db.ExecContext(ctx, `
			INSERT INTO plant_common_names (plant_id, language_id, common_name)
			VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', $2)
		`, plantID, fmt.Sprintf("%s Common %d", genus, i))
		if err != nil {
			t.Fatalf("failed to seed common name: %v", err)
		}
	}

	t.Run("SearchByBotanicalName", func(b *testing.B) {
		filter := repository.DefaultSearchFilter()
		filter.Limit = 20

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.Search(ctx, "Rosa", filter, "550e8400-e29b-41d4-a716-446655440001", nil)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	})

	t.Run("SearchByCommonName", func(b *testing.B) {
		filter := repository.DefaultSearchFilter()
		filter.Limit = 20

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.Search(ctx, "Common", filter, "550e8400-e29b-41d4-a716-446655440001", nil)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	})
}

// BenchmarkPlantRepository_LocalizationFallback benchmarks localization performance
func BenchmarkPlantRepository_LocalizationFallback(t *testing.B) {
	db := testhelpers.SetupTestDB(&testing.T{})
	defer testhelpers.TeardownTestDB(&testing.T{}, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(&testing.T{}, db)
	testhelpers.SeedTestCountries(&testing.T{}, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(&testing.T{}, db)

	// Create plant with multiple language variants
	plantID := uuid.New().String()
	_, err := db.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, 'Rosa benchmark', NOW())
	`, plantID, speciesID)
	if err != nil {
		t.Fatalf("failed to seed plant: %v", err)
	}

	// Add common names in multiple languages
	names := []struct {
		lang    string
		country *string
		name    string
	}{
		{"550e8400-e29b-41d4-a716-446655440001", nil, "English Rose"},
		{"550e8400-e29b-41d4-a716-446655440002", nil, "Rosa Española"},
		{"550e8400-e29b-41d4-a716-446655440003", nil, "Rose Française"},
	}

	for _, n := range names {
		query := `
			INSERT INTO plant_common_names (plant_id, language_id, country_id, common_name)
			VALUES ($1, $2, $3, $4)
		`
		_, err = db.ExecContext(ctx, query, plantID, n.lang, n.country, n.name)
		if err != nil {
			t.Fatalf("failed to seed common name: %v", err)
		}
	}

	t.Run("EnglishLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.FindByID(ctx, plantID, "550e8400-e29b-41d4-a716-446655440001", nil)
			if err != nil {
				b.Fatalf("FindByID failed: %v", err)
			}
		}
	})

	t.Run("SpanishLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.FindByID(ctx, plantID, "550e8400-e29b-41d4-a716-446655440002", nil)
			if err != nil {
				b.Fatalf("FindByID failed: %v", err)
			}
		}
	})

	t.Run("CountrySpecificLookup", func(b *testing.B) {
		englishID := "550e8400-e29b-41d4-a716-446655440001"
		country := "650e8400-e29b-41d4-a716-446655440001" // US UUID
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.FindByID(ctx, plantID, englishID, &country)
			if err != nil {
				b.Fatalf("FindByID failed: %v", err)
			}
		}
	})
}

// BenchmarkPlantRepository_Search_CommonNames benchmarks search with common names
func BenchmarkPlantRepository_Search_CommonNames(t *testing.B) {
	db := testhelpers.SetupTestDB(&testing.T{})
	defer testhelpers.TeardownTestDB(&testing.T{}, db)

	repo := NewPostgresPlantRepository(db)
	ctx := context.Background()

	// Seed test data
	testhelpers.SeedTestLanguages(&testing.T{}, db)
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(&testing.T{}, db)

	// Create 50 test plants with common names for realistic search scenario
	plantIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		plantIDs[i] = uuid.New().String()
		botanicalName := fmt.Sprintf("Rosa search-benchmark-%d", i)

		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
			VALUES ($1, $2, $3, NOW())
		`, plantIDs[i], speciesID, botanicalName)
		if err != nil {
			t.Fatalf("failed to seed plant %d: %v", i, err)
		}

		// Add common names - include "Japanese" in every 10th plant
		commonName := fmt.Sprintf("Common Rose %d", i)
		if i%10 == 0 {
			commonName = fmt.Sprintf("Japanese Rose %d", i)
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO plant_common_names (plant_id, language_id, common_name)
			VALUES ($1, '550e8400-e29b-41d4-a716-446655440001', $2)
		`, plantIDs[i], commonName)
		if err != nil {
			t.Fatalf("failed to seed common name for plant %d: %v", i, err)
		}
	}

	englishID := "550e8400-e29b-41d4-a716-446655440001"
	filter := repository.DefaultSearchFilter()
	filter.Limit = 20

	t.Run("SearchByBotanicalName", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.Search(ctx, "Rosa", filter, englishID, nil)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	})

	t.Run("SearchByCommonName", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := repo.Search(ctx, "Japanese", filter, englishID, nil)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
			if len(result.Plants) == 0 {
				b.Fatalf("Expected to find plants with 'Japanese' in common name")
			}
		}
	})

	t.Run("SearchEmptyQuery", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := repo.Search(ctx, "", filter, englishID, nil)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	})
}
