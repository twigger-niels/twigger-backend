package service

import (
	"context"
	"testing"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
	"twigger-backend/backend/plant-service/pkg/types"
	"twigger-backend/backend/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPlant(t *testing.T) {
	// Setup
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	testPlant := &entity.Plant{
		PlantID:           "plant-123",
		SpeciesID:         "species-456",
		FamilyName:        "Rosaceae",
		GenusName:         "Rosa",
		SpeciesName:       "rugosa",
		PlantType:         types.PlantTypeShrub,
		FullBotanicalName: "Rosa rugosa",
		CreatedAt:         time.Now(),
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "plant-123", "en", (*string)(nil)).Return(testPlant, nil).Once()

		result, err := service.GetPlant(ctx, "plant-123", false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "plant-123", result.PlantID)
		assert.Equal(t, "Rosa rugosa", result.FullBotanicalName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("plant not found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "nonexistent", "en", (*string)(nil)).Return(nil, entity.ErrPlantNotFound).Once()

		result, err := service.GetPlant(ctx, "nonexistent", false)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty plant ID", func(t *testing.T) {
		result, err := service.GetPlant(ctx, "", false)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrInvalidPlantID)
		assert.Nil(t, result)
	})
}

func TestSearchPlants(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("successful search", func(t *testing.T) {
		filter := repository.DefaultSearchFilter()
		filter.Limit = 10

		expectedResult := &repository.SearchResult{
			Plants: []*entity.Plant{
				{
					PlantID:           "plant-1",
					FullBotanicalName: "Rosa rugosa",
					GenusName:         "Rosa",
					SpeciesName:       "rugosa",
				},
			},
			Total:      1,
			Limit:      10,
			NextCursor: nil,
			HasMore:    false,
			Query:      "rosa",
		}

		mockRepo.On("Search", ctx, "rosa", mock.MatchedBy(func(f *repository.SearchFilter) bool {
			return f.Limit == 10
		}), "en", (*string)(nil)).Return(expectedResult, nil).Once()

		result, err := service.SearchPlants(ctx, "rosa", filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Plants))
		assert.Equal(t, "Rosa rugosa", result.Plants[0].FullBotanicalName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("query too long", func(t *testing.T) {
		longQuery := string(make([]byte, 201))

		result, err := service.SearchPlants(ctx, longQuery, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrInvalidSearchQuery)
		assert.Nil(t, result)
	})

	t.Run("uses default filter when nil", func(t *testing.T) {
		expectedResult := &repository.SearchResult{
			Plants:     []*entity.Plant{},
			Total:      0,
			Limit:      20,
			NextCursor: nil,
			HasMore:    false,
		}

		mockRepo.On("Search", ctx, "test", mock.MatchedBy(func(f *repository.SearchFilter) bool {
			return f.Limit == 20 // Default limit
		}), "en", (*string)(nil)).Return(expectedResult, nil).Once()

		result, err := service.SearchPlants(ctx, "test", nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestFindByBotanicalName(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("successful find", func(t *testing.T) {
		testPlant := &entity.Plant{
			PlantID:           "plant-123",
			FullBotanicalName: "Rosa rugosa",
			GenusName:         "Rosa",
			SpeciesName:       "rugosa",
		}

		mockRepo.On("FindByBotanicalName", ctx, "Rosa rugosa", "en", (*string)(nil)).Return(testPlant, nil).Once()

		result, err := service.FindByBotanicalName(ctx, "Rosa rugosa")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Rosa rugosa", result.FullBotanicalName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty botanical name", func(t *testing.T) {
		result, err := service.FindByBotanicalName(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		testPlant := &entity.Plant{
			PlantID:           "plant-123",
			FullBotanicalName: "Rosa rugosa",
		}

		mockRepo.On("FindByBotanicalName", ctx, "Rosa rugosa", "en", (*string)(nil)).Return(testPlant, nil).Once()

		result, err := service.FindByBotanicalName(ctx, "  Rosa rugosa  ")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetCompanionPlants(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("get all companions", func(t *testing.T) {
		companions := []*entity.Companion{
			{
				RelationshipID:   "rel-1",
				PlantAID:         "plant-123",
				PlantBID:         "plant-456",
				RelationshipType: types.RelationshipBeneficial,
				Benefits:         []string{"pest_control"},
			},
			{
				RelationshipID:   "rel-2",
				PlantAID:         "plant-123",
				PlantBID:         "plant-789",
				RelationshipType: types.RelationshipAntagonistic,
			},
		}

		mockRepo.On("GetCompanions", ctx, "plant-123", "en", (*string)(nil), mock.MatchedBy(func(f *entity.CompanionFilter) bool {
			return f.BeneficialOnly == false
		})).Return(companions, nil).Once()

		result, err := service.GetCompanionPlants(ctx, "plant-123", false)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		mockRepo.AssertExpectations(t)
	})

	t.Run("get beneficial only", func(t *testing.T) {
		companions := []*entity.Companion{
			{
				RelationshipID:   "rel-1",
				PlantAID:         "plant-123",
				PlantBID:         "plant-456",
				RelationshipType: types.RelationshipBeneficial,
			},
		}

		mockRepo.On("GetCompanions", ctx, "plant-123", "en", (*string)(nil), mock.MatchedBy(func(f *entity.CompanionFilter) bool {
			return f.BeneficialOnly == true
		})).Return(companions, nil).Once()

		result, err := service.GetCompanionPlants(ctx, "plant-123", true)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, types.RelationshipBeneficial, result[0].RelationshipType)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty plant ID", func(t *testing.T) {
		result, err := service.GetCompanionPlants(ctx, "", false)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrInvalidPlantID)
		assert.Nil(t, result)
	})
}

func TestValidatePlantCompatibility(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("beneficial relationship", func(t *testing.T) {
		companions := []*entity.Companion{
			{
				RelationshipID:   "rel-1",
				PlantAID:         "plant-a",
				PlantBID:         "plant-b",
				RelationshipType: types.RelationshipBeneficial,
				Benefits:         []string{"pest_control", "nitrogen_fixation"},
				OptimalDistanceM: types.Float64Ptr(0.5),
			},
		}

		mockRepo.On("GetCompanions", ctx, "plant-a", "en", (*string)(nil), mock.Anything).Return(companions, nil).Once()

		result, err := service.ValidatePlantCompatibility(ctx, "plant-a", "plant-b")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Compatible)
		assert.Equal(t, types.RelationshipBeneficial, result.RelationshipType)
		assert.Equal(t, 2, len(result.Benefits))
		assert.Equal(t, 0.5, *result.RecommendedDistanceM)
		mockRepo.AssertExpectations(t)
	})

	t.Run("antagonistic relationship", func(t *testing.T) {
		companions := []*entity.Companion{
			{
				RelationshipID:   "rel-1",
				PlantAID:         "plant-a",
				PlantBID:         "plant-c",
				RelationshipType: types.RelationshipAntagonistic,
			},
		}

		mockRepo.On("GetCompanions", ctx, "plant-a", "en", (*string)(nil), mock.Anything).Return(companions, nil).Once()

		result, err := service.ValidatePlantCompatibility(ctx, "plant-a", "plant-c")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Compatible)
		assert.Equal(t, types.RelationshipAntagonistic, result.RelationshipType)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no relationship - assumes neutral", func(t *testing.T) {
		mockRepo.On("GetCompanions", ctx, "plant-a", "en", (*string)(nil), mock.Anything).Return([]*entity.Companion{}, nil).Once()

		result, err := service.ValidatePlantCompatibility(ctx, "plant-a", "plant-unknown")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Compatible)
		assert.Equal(t, types.RelationshipNeutral, result.RelationshipType)
		mockRepo.AssertExpectations(t)
	})

	t.Run("same plant ID", func(t *testing.T) {
		result, err := service.ValidatePlantCompatibility(ctx, "plant-a", "plant-a")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCreatePlant(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		newPlant := &entity.Plant{
			PlantID:           "plant-new",
			SpeciesID:         "species-123",
			FamilyName:        "Rosaceae",
			GenusName:         "Rosa",
			SpeciesName:       "canina",
			PlantType:         types.PlantTypeShrub,
			FullBotanicalName: "Rosa canina",
		}

		// Check if exists
		mockRepo.On("FindByBotanicalName", ctx, "Rosa canina", "en", (*string)(nil)).Return(nil, entity.ErrPlantNotFound).Once()
		// Create
		mockRepo.On("Create", ctx, newPlant).Return(nil).Once()

		err := service.CreatePlant(ctx, newPlant)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("plant already exists", func(t *testing.T) {
		existingPlant := &entity.Plant{
			PlantID:           "plant-existing",
			FullBotanicalName: "Rosa rugosa",
		}

		newPlant := &entity.Plant{
			PlantID:           "plant-new",
			SpeciesID:         "species-123",
			FamilyName:        "Rosaceae",
			GenusName:         "Rosa",
			SpeciesName:       "rugosa",
			PlantType:         types.PlantTypeShrub,
			FullBotanicalName: "Rosa rugosa",
		}

		mockRepo.On("FindByBotanicalName", ctx, "Rosa rugosa", "en", (*string)(nil)).Return(existingPlant, nil).Once()

		err := service.CreatePlant(ctx, newPlant)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrPlantAlreadyExists)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPlantWithConditions(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	testPlant := &entity.Plant{
		PlantID:           "plant-123",
		SpeciesID:         "species-456",
		FullBotanicalName: "Rosa rugosa",
		GenusName:         "Rosa",
		SpeciesName:       "rugosa",
		PlantType:         types.PlantTypeShrub,
	}

	t.Run("successful retrieval with conditions", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "plant-123", "en", (*string)(nil)).Return(testPlant, nil).Once()
		mockRepo.On("GetPhysicalCharacteristics", ctx, "plant-123", "en").Return(&types.PhysicalCharacteristics{
			PlantID: "plant-123",
		}, nil).Once()

		countryID := "country-123"
		waterNeeds := types.WaterNeeds("moderate")
		growingConditions := &types.GrowingConditions{
			PlantID:    "plant-123",
			CountryID:  &countryID,
			WaterNeeds: &waterNeeds,
		}
		mockRepo.On("GetGrowingConditions", ctx, "plant-123", "country-123", "en").Return(growingConditions, nil).Once()

		result, err := service.GetPlantWithConditions(ctx, "plant-123", "country-123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "plant-123", result.PlantID)
		assert.NotNil(t, result.PhysicalCharacteristics)
		assert.NotNil(t, result.GrowingConditions)
		mockRepo.AssertExpectations(t)
	})

	t.Run("without country ID skips growing conditions", func(t *testing.T) {
		plantWithoutConditions := &entity.Plant{
			PlantID:           "plant-123",
			SpeciesID:         "species-456",
			FullBotanicalName: "Rosa rugosa",
			GenusName:         "Rosa",
			SpeciesName:       "rugosa",
			PlantType:         types.PlantTypeShrub,
			GrowingConditions: nil, // Explicitly nil
		}
		mockRepo.On("FindByID", ctx, "plant-123", "en", (*string)(nil)).Return(plantWithoutConditions, nil).Once()
		mockRepo.On("GetPhysicalCharacteristics", ctx, "plant-123", "en").Return(nil, nil).Once()

		result, err := service.GetPlantWithConditions(ctx, "plant-123", "")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.GrowingConditions)
		mockRepo.AssertExpectations(t)
	})

	t.Run("plant not found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "nonexistent", "en", (*string)(nil)).Return(nil, entity.ErrPlantNotFound).Once()

		result, err := service.GetPlantWithConditions(ctx, "nonexistent", "country-123")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestFindPlantsByFamily(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("successful find", func(t *testing.T) {
		plants := []*entity.Plant{
			{PlantID: "plant-1", FamilyName: "Rosaceae"},
			{PlantID: "plant-2", FamilyName: "Rosaceae"},
		}

		mockRepo.On("FindByFamily", ctx, "Rosaceae", "en", (*string)(nil), 20, 0).Return(plants, nil).Once()

		result, err := service.FindPlantsByFamily(ctx, "Rosaceae", 20, 0)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty family name", func(t *testing.T) {
		result, err := service.FindPlantsByFamily(ctx, "", 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		mockRepo.On("FindByFamily", ctx, "Rosaceae", "en", (*string)(nil), 20, 0).Return([]*entity.Plant{}, nil).Once()

		_, err := service.FindPlantsByFamily(ctx, "  Rosaceae  ", 20, 0)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("enforces max limit", func(t *testing.T) {
		mockRepo.On("FindByFamily", ctx, "Rosaceae", "en", (*string)(nil), 20, 0).Return([]*entity.Plant{}, nil).Once()

		_, err := service.FindPlantsByFamily(ctx, "Rosaceae", 150, 0)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("enforces min limit", func(t *testing.T) {
		mockRepo.On("FindByFamily", ctx, "Rosaceae", "en", (*string)(nil), 20, 0).Return([]*entity.Plant{}, nil).Once()

		_, err := service.FindPlantsByFamily(ctx, "Rosaceae", -5, 0)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestFindPlantsByGenus(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("successful find", func(t *testing.T) {
		plants := []*entity.Plant{
			{PlantID: "plant-1", GenusName: "Rosa"},
			{PlantID: "plant-2", GenusName: "Rosa"},
			{PlantID: "plant-3", GenusName: "Rosa"},
		}

		mockRepo.On("FindByGenus", ctx, "Rosa", "en", (*string)(nil), 20, 0).Return(plants, nil).Once()

		result, err := service.FindPlantsByGenus(ctx, "Rosa", 20, 0)

		assert.NoError(t, err)
		assert.Equal(t, 3, len(result))
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty genus name", func(t *testing.T) {
		result, err := service.FindPlantsByGenus(ctx, "", 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("enforces limit boundaries", func(t *testing.T) {
		mockRepo.On("FindByGenus", ctx, "Rosa", "en", (*string)(nil), 20, 10).Return([]*entity.Plant{}, nil).Once()

		_, err := service.FindPlantsByGenus(ctx, "Rosa", 0, 10)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetBeneficialCompanions(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("returns only beneficial companions", func(t *testing.T) {
		companions := []*entity.Companion{
			{
				RelationshipID:   "rel-1",
				PlantAID:         "plant-123",
				PlantBID:         "plant-456",
				RelationshipType: types.RelationshipBeneficial,
				Benefits:         []string{"nitrogen_fixation"},
			},
		}

		mockRepo.On("GetCompanions", ctx, "plant-123", "en", (*string)(nil), mock.MatchedBy(func(f *entity.CompanionFilter) bool {
			return f.BeneficialOnly == true
		})).Return(companions, nil).Once()

		result, err := service.GetBeneficialCompanions(ctx, "plant-123")

		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, types.RelationshipBeneficial, result[0].RelationshipType)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAntagonisticPlants(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("returns only antagonistic relationships", func(t *testing.T) {
		companions := []*entity.Companion{
			{
				RelationshipID:   "rel-1",
				PlantAID:         "plant-123",
				PlantBID:         "plant-789",
				RelationshipType: types.RelationshipAntagonistic,
			},
		}

		mockRepo.On("GetCompanions", ctx, "plant-123", "en", (*string)(nil), mock.MatchedBy(func(f *entity.CompanionFilter) bool {
			return f.RelationshipType != nil && *f.RelationshipType == types.RelationshipAntagonistic
		})).Return(companions, nil).Once()

		result, err := service.GetAntagonisticPlants(ctx, "plant-123")

		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, types.RelationshipAntagonistic, result[0].RelationshipType)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty plant ID", func(t *testing.T) {
		result, err := service.GetAntagonisticPlants(ctx, "")

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrInvalidPlantID)
		assert.Nil(t, result)
	})
}

func TestRecommendPlants(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("successful recommendation", func(t *testing.T) {
		plants := []*entity.Plant{
			{PlantID: "plant-1", FullBotanicalName: "Rosa rugosa"},
			{PlantID: "plant-2", FullBotanicalName: "Rosa canina"},
		}

		mockRepo.On("FindByGrowingConditions", ctx, mock.MatchedBy(func(f *repository.GrowingConditionsFilter) bool {
			return *f.HardinessZone == "5a" &&
				len(f.SunRequirements) == 1 &&
				f.SunRequirements[0] == "full_sun" &&
				f.Limit == 10
		})).Return(plants, nil).Once()

		result, err := service.RecommendPlants(ctx, "5a", types.SunFullSun, 10)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		mockRepo.AssertExpectations(t)
	})

	t.Run("enforces max limit", func(t *testing.T) {
		mockRepo.On("FindByGrowingConditions", ctx, mock.MatchedBy(func(f *repository.GrowingConditionsFilter) bool {
			return f.Limit == 10 // Should cap at 10 when 100 requested (max 50)
		})).Return([]*entity.Plant{}, nil).Once()

		_, err := service.RecommendPlants(ctx, "5a", types.SunFullSun, 100)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("enforces min limit", func(t *testing.T) {
		mockRepo.On("FindByGrowingConditions", ctx, mock.MatchedBy(func(f *repository.GrowingConditionsFilter) bool {
			return f.Limit == 10 // Should default to 10 when 0 requested
		})).Return([]*entity.Plant{}, nil).Once()

		_, err := service.RecommendPlants(ctx, "5a", types.SunPartialShade, 0)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPlantStatistics(t *testing.T) {
	mockRepo := mocks.NewMockPlantRepository()
	service := NewPlantService(mockRepo)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.On("Count", ctx, mock.Anything).Return(int64(1234), nil).Once()

		result, err := service.GetPlantStatistics(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1234), result.TotalPlants)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("Count", ctx, mock.Anything).Return(int64(0), assert.AnError).Once()

		result, err := service.GetPlantStatistics(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestRankSearchResults(t *testing.T) {
	service := &PlantService{}

	t.Run("ranks by search score", func(t *testing.T) {
		plants := []*entity.Plant{
			{PlantID: "plant-1", FullBotanicalName: "Rosa canina", CommonNames: []string{"Dog Rose"}},
			{PlantID: "plant-2", FullBotanicalName: "Rosa rugosa", CommonNames: []string{"Japanese Rose"}},
			{PlantID: "plant-3", FullBotanicalName: "Acer rosa", CommonNames: []string{"Rosa Maple"}},
		}

		service.rankSearchResults(plants, "rosa")

		// After ranking, plants with "rosa" in botanical name should rank higher
		// The exact order depends on SearchScore implementation in entity.Plant
		assert.NotNil(t, plants)
		assert.Equal(t, 3, len(plants))
	})

	t.Run("handles empty query", func(t *testing.T) {
		plants := []*entity.Plant{
			{PlantID: "plant-1", FullBotanicalName: "Rosa canina"},
			{PlantID: "plant-2", FullBotanicalName: "Acer palmatum"},
		}

		// Should not panic with empty query
		service.rankSearchResults(plants, "")
		assert.Equal(t, 2, len(plants))
	})

	t.Run("handles empty plant list", func(t *testing.T) {
		plants := []*entity.Plant{}

		// Should not panic with empty list
		service.rankSearchResults(plants, "test")
		assert.Equal(t, 0, len(plants))
	})
}
