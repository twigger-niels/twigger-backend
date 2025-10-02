package mocks

import (
	"context"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
	"twigger-backend/backend/plant-service/pkg/types"

	"github.com/stretchr/testify/mock"
)

// MockPlantRepository is a mock implementation of PlantRepository for testing
type MockPlantRepository struct {
	mock.Mock
}

// FindByID mocks the FindByID method
func (m *MockPlantRepository) FindByID(ctx context.Context, plantID, languageID string, countryID *string) (*entity.Plant, error) {
	args := m.Called(ctx, plantID, languageID, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Plant), args.Error(1)
}

// FindByIDs mocks the FindByIDs method
func (m *MockPlantRepository) FindByIDs(ctx context.Context, plantIDs []string, languageID string, countryID *string) ([]*entity.Plant, error) {
	args := m.Called(ctx, plantIDs, languageID, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Plant), args.Error(1)
}

// Create mocks the Create method
func (m *MockPlantRepository) Create(ctx context.Context, plant *entity.Plant) error {
	args := m.Called(ctx, plant)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockPlantRepository) Update(ctx context.Context, plant *entity.Plant) error {
	args := m.Called(ctx, plant)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockPlantRepository) Delete(ctx context.Context, plantID string) error {
	args := m.Called(ctx, plantID)
	return args.Error(0)
}

// Search mocks the Search method
func (m *MockPlantRepository) Search(ctx context.Context, query string, filter *repository.SearchFilter, languageID string, countryID *string) (*repository.SearchResult, error) {
	args := m.Called(ctx, query, filter, languageID, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SearchResult), args.Error(1)
}

// FindByBotanicalName mocks the FindByBotanicalName method
func (m *MockPlantRepository) FindByBotanicalName(ctx context.Context, botanicalName, languageID string, countryID *string) (*entity.Plant, error) {
	args := m.Called(ctx, botanicalName, languageID, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Plant), args.Error(1)
}

// FindByCommonName mocks the FindByCommonName method
func (m *MockPlantRepository) FindByCommonName(ctx context.Context, commonName, languageID string, countryID *string) ([]*entity.Plant, error) {
	args := m.Called(ctx, commonName, languageID, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Plant), args.Error(1)
}

// FindByFamily mocks the FindByFamily method
func (m *MockPlantRepository) FindByFamily(ctx context.Context, familyName, languageID string, countryID *string, limit, offset int) ([]*entity.Plant, error) {
	args := m.Called(ctx, familyName, languageID, countryID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Plant), args.Error(1)
}

// FindByGenus mocks the FindByGenus method
func (m *MockPlantRepository) FindByGenus(ctx context.Context, genusName, languageID string, countryID *string, limit, offset int) ([]*entity.Plant, error) {
	args := m.Called(ctx, genusName, languageID, countryID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Plant), args.Error(1)
}

// FindBySpecies mocks the FindBySpecies method
func (m *MockPlantRepository) FindBySpecies(ctx context.Context, genusName, speciesName, languageID string, countryID *string) ([]*entity.Plant, error) {
	args := m.Called(ctx, genusName, speciesName, languageID, countryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Plant), args.Error(1)
}

// GetGrowingConditions mocks the GetGrowingConditions method
func (m *MockPlantRepository) GetGrowingConditions(ctx context.Context, plantID, countryID, languageID string) (*types.GrowingConditions, error) {
	args := m.Called(ctx, plantID, countryID, languageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.GrowingConditions), args.Error(1)
}

// FindByGrowingConditions mocks the FindByGrowingConditions method
func (m *MockPlantRepository) FindByGrowingConditions(ctx context.Context, filter *repository.GrowingConditionsFilter) ([]*entity.Plant, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Plant), args.Error(1)
}

// GetPhysicalCharacteristics mocks the GetPhysicalCharacteristics method
func (m *MockPlantRepository) GetPhysicalCharacteristics(ctx context.Context, plantID, languageID string) (*types.PhysicalCharacteristics, error) {
	args := m.Called(ctx, plantID, languageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PhysicalCharacteristics), args.Error(1)
}

// GetCompanions mocks the GetCompanions method
func (m *MockPlantRepository) GetCompanions(ctx context.Context, plantID, languageID string, countryID *string, filter *entity.CompanionFilter) ([]*entity.Companion, error) {
	args := m.Called(ctx, plantID, languageID, countryID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Companion), args.Error(1)
}

// GetCompanionsByType mocks the GetCompanionsByType method
func (m *MockPlantRepository) GetCompanionsByType(ctx context.Context, plantID, languageID string, countryID *string, relType types.RelationshipType) ([]*entity.Companion, error) {
	args := m.Called(ctx, plantID, languageID, countryID, relType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Companion), args.Error(1)
}

// CreateCompanionRelationship mocks the CreateCompanionRelationship method
func (m *MockPlantRepository) CreateCompanionRelationship(ctx context.Context, companion *entity.Companion) error {
	args := m.Called(ctx, companion)
	return args.Error(0)
}

// DeleteCompanionRelationship mocks the DeleteCompanionRelationship method
func (m *MockPlantRepository) DeleteCompanionRelationship(ctx context.Context, relationshipID string) error {
	args := m.Called(ctx, relationshipID)
	return args.Error(0)
}

// BulkCreate mocks the BulkCreate method
func (m *MockPlantRepository) BulkCreate(ctx context.Context, plants []*entity.Plant) error {
	args := m.Called(ctx, plants)
	return args.Error(0)
}

// Count mocks the Count method
func (m *MockPlantRepository) Count(ctx context.Context, filter *repository.SearchFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// NewMockPlantRepository creates a new mock plant repository
func NewMockPlantRepository() *MockPlantRepository {
	return &MockPlantRepository{}
}
