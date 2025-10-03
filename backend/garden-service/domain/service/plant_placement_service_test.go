package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"twigger-backend/backend/garden-service/domain/entity"
)

// MockGardenPlantRepository is a mock implementation of repository.GardenPlantRepository
type MockGardenPlantRepository struct {
	mock.Mock
}

func (m *MockGardenPlantRepository) Create(ctx context.Context, gardenPlant *entity.GardenPlant) error {
	args := m.Called(ctx, gardenPlant)
	return args.Error(0)
}

func (m *MockGardenPlantRepository) FindByID(ctx context.Context, gardenPlantID string) (*entity.GardenPlant, error) {
	args := m.Called(ctx, gardenPlantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GardenPlant), args.Error(1)
}

func (m *MockGardenPlantRepository) FindByGardenID(ctx context.Context, gardenID string) ([]*entity.GardenPlant, error) {
	args := m.Called(ctx, gardenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenPlant), args.Error(1)
}

func (m *MockGardenPlantRepository) FindByIDs(ctx context.Context, gardenPlantIDs []string) ([]*entity.GardenPlant, error) {
	args := m.Called(ctx, gardenPlantIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenPlant), args.Error(1)
}

func (m *MockGardenPlantRepository) Update(ctx context.Context, gardenPlant *entity.GardenPlant) error {
	args := m.Called(ctx, gardenPlant)
	return args.Error(0)
}

func (m *MockGardenPlantRepository) Delete(ctx context.Context, gardenPlantID string) error {
	args := m.Called(ctx, gardenPlantID)
	return args.Error(0)
}

func (m *MockGardenPlantRepository) CheckPlantSpacing(ctx context.Context, gardenID, locationGeoJSON string, minDistanceM float64) ([]*entity.GardenPlant, error) {
	args := m.Called(ctx, gardenID, locationGeoJSON, minDistanceM)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenPlant), args.Error(1)
}

func (m *MockGardenPlantRepository) FindInZone(ctx context.Context, zoneID string) ([]*entity.GardenPlant, error) {
	args := m.Called(ctx, zoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenPlant), args.Error(1)
}

func (m *MockGardenPlantRepository) ValidatePlantLocation(ctx context.Context, gardenID, locationGeoJSON string) error {
	args := m.Called(ctx, gardenID, locationGeoJSON)
	return args.Error(0)
}

func (m *MockGardenPlantRepository) FindByHealthStatus(ctx context.Context, gardenID string, healthStatus entity.HealthStatus) ([]*entity.GardenPlant, error) {
	args := m.Called(ctx, gardenID, healthStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenPlant), args.Error(1)
}

func (m *MockGardenPlantRepository) FindActivePlants(ctx context.Context, gardenID string) ([]*entity.GardenPlant, error) {
	args := m.Called(ctx, gardenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenPlant), args.Error(1)
}

func (m *MockGardenPlantRepository) BulkCreate(ctx context.Context, gardenPlants []*entity.GardenPlant) error {
	args := m.Called(ctx, gardenPlants)
	return args.Error(0)
}

func (m *MockGardenPlantRepository) CountByGardenID(ctx context.Context, gardenID string) (int, error) {
	args := m.Called(ctx, gardenID)
	return args.Get(0).(int), args.Error(1)
}

// Test PlacePlant
func TestPlantPlacementService_PlacePlant_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	plantID := "plant-123"
	location := `{"type":"Point","coordinates":[-122.4194,37.7749]}`

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: location,
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("ValidatePlantLocation", ctx, gardenID, location).Return(nil)
	mockPlantRepo.On("Create", ctx, mock.AnythingOfType("*entity.GardenPlant")).Return(nil)

	result, err := service.PlacePlant(ctx, gardenPlant)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.GardenPlantID)
	assert.False(t, result.PlantedAt.IsZero())
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
	assert.Equal(t, entity.HealthStatusHealthy, *result.HealthStatus) // Default
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

func TestPlantPlacementService_PlacePlant_WithZone(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	plantID := "plant-123"
	zoneID := "zone-123"
	location := `{"type":"Point","coordinates":[-122.4194,37.7749]}`

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	zoneType := entity.ZoneTypeBed
	existingZone := &entity.GardenZone{
		ZoneID:   zoneID,
		GardenID: gardenID,
		ZoneType: &zoneType,
	}

	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		ZoneID:          &zoneID,
		LocationGeoJSON: location,
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("ValidatePlantLocation", ctx, gardenID, location).Return(nil)
	mockZoneRepo.On("FindByID", ctx, zoneID).Return(existingZone, nil)
	mockPlantRepo.On("Create", ctx, mock.AnythingOfType("*entity.GardenPlant")).Return(nil)

	result, err := service.PlacePlant(ctx, gardenPlant)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
	mockZoneRepo.AssertExpectations(t)
}

func TestPlantPlacementService_PlacePlant_ZoneNotInGarden(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	plantID := "plant-123"
	zoneID := "zone-456"
	location := `{"type":"Point","coordinates":[-122.4194,37.7749]}`

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	zoneType := entity.ZoneTypeBed
	differentGardenZone := &entity.GardenZone{
		ZoneID:   zoneID,
		GardenID: "different-garden",
		ZoneType: &zoneType,
	}

	gardenPlant := &entity.GardenPlant{
		GardenID:        gardenID,
		PlantID:         plantID,
		ZoneID:          &zoneID,
		LocationGeoJSON: location,
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("ValidatePlantLocation", ctx, gardenID, location).Return(nil)
	mockZoneRepo.On("FindByID", ctx, zoneID).Return(differentGardenZone, nil)

	_, err := service.PlacePlant(ctx, gardenPlant)

	assert.Error(t, err)
	assert.IsType(t, &entity.ValidationError{}, err)
	assert.Contains(t, err.Error(), "zone must belong to the same garden")
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
	mockZoneRepo.AssertExpectations(t)
}

// Test GetGardenPlant
func TestPlantPlacementService_GetGardenPlant_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenPlantID := "gp-123"
	healthStatus := entity.HealthStatusHealthy

	expectedPlant := &entity.GardenPlant{
		GardenPlantID:   gardenPlantID,
		GardenID:        "garden-123",
		PlantID:         "plant-123",
		LocationGeoJSON: "{}",
		HealthStatus:    &healthStatus,
	}

	mockPlantRepo.On("FindByID", ctx, gardenPlantID).Return(expectedPlant, nil)

	result, err := service.GetGardenPlant(ctx, gardenPlantID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPlant, result)
	mockPlantRepo.AssertExpectations(t)
}

// Test ListGardenPlants
func TestPlantPlacementService_ListGardenPlants_NoFilters(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	expectedPlants := []*entity.GardenPlant{
		{GardenPlantID: "gp-1", GardenID: gardenID},
		{GardenPlantID: "gp-2", GardenID: gardenID},
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("FindByGardenID", ctx, gardenID).Return(expectedPlants, nil)

	result, err := service.ListGardenPlants(ctx, gardenID, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

func TestPlantPlacementService_ListGardenPlants_FilterByZone(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	zoneID := "zone-123"

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	expectedPlants := []*entity.GardenPlant{
		{GardenPlantID: "gp-1", GardenID: gardenID, ZoneID: &zoneID},
	}

	filter := &GardenPlantFilter{ZoneID: &zoneID}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("FindInZone", ctx, zoneID).Return(expectedPlants, nil)

	result, err := service.ListGardenPlants(ctx, gardenID, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

func TestPlantPlacementService_ListGardenPlants_FilterByHealthStatus(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	healthStatus := entity.HealthStatusStruggling

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	expectedPlants := []*entity.GardenPlant{
		{GardenPlantID: "gp-1", GardenID: gardenID, HealthStatus: &healthStatus},
	}

	filter := &GardenPlantFilter{HealthStatus: &healthStatus}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("FindByHealthStatus", ctx, gardenID, healthStatus).Return(expectedPlants, nil)

	result, err := service.ListGardenPlants(ctx, gardenID, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

func TestPlantPlacementService_ListGardenPlants_FilterActiveOnly(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	expectedPlants := []*entity.GardenPlant{
		{GardenPlantID: "gp-1", GardenID: gardenID},
	}

	filter := &GardenPlantFilter{ActiveOnly: true}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("FindActivePlants", ctx, gardenID).Return(expectedPlants, nil)

	result, err := service.ListGardenPlants(ctx, gardenID, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

// Test UpdatePlantPlacement
func TestPlantPlacementService_UpdatePlantPlacement_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenPlantID := "gp-123"
	gardenID := "garden-123"
	plantID := "plant-123"
	location := `{"type":"Point","coordinates":[-122.4194,37.7749]}`

	createdAt := time.Now().Add(-24 * time.Hour)
	plantedAt := time.Now().Add(-24 * time.Hour)
	healthStatus := entity.HealthStatusHealthy

	existingPlant := &entity.GardenPlant{
		GardenPlantID:   gardenPlantID,
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: location,
		CreatedAt:       createdAt,
		PlantedAt:       plantedAt,
		HealthStatus:    &healthStatus,
	}

	newStatus := entity.HealthStatusThriving
	updatedPlant := &entity.GardenPlant{
		GardenPlantID:   gardenPlantID,
		GardenID:        gardenID,
		PlantID:         plantID,
		LocationGeoJSON: location,
		HealthStatus:    &newStatus,
	}

	mockPlantRepo.On("FindByID", ctx, gardenPlantID).Return(existingPlant, nil)
	mockPlantRepo.On("Update", ctx, mock.AnythingOfType("*entity.GardenPlant")).Return(nil)

	result, err := service.UpdatePlantPlacement(ctx, updatedPlant)

	assert.NoError(t, err)
	assert.Equal(t, newStatus, *result.HealthStatus)
	assert.Equal(t, createdAt, result.CreatedAt) // Preserved
	assert.Equal(t, plantedAt, result.PlantedAt) // Preserved
	assert.True(t, result.UpdatedAt.After(createdAt)) // Updated
	mockPlantRepo.AssertExpectations(t)
}

// Test RemovePlant
func TestPlantPlacementService_RemovePlant_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenPlantID := "gp-123"
	healthStatus := entity.HealthStatusHealthy

	existingPlant := &entity.GardenPlant{
		GardenPlantID:   gardenPlantID,
		GardenID:        "garden-123",
		PlantID:         "plant-123",
		LocationGeoJSON: "{}",
		HealthStatus:    &healthStatus,
	}

	mockPlantRepo.On("FindByID", ctx, gardenPlantID).Return(existingPlant, nil)
	mockPlantRepo.On("Delete", ctx, gardenPlantID).Return(nil)

	err := service.RemovePlant(ctx, gardenPlantID)

	assert.NoError(t, err)
	mockPlantRepo.AssertExpectations(t)
}

// Test CheckPlantSpacing
func TestPlantPlacementService_CheckPlantSpacing_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	location := `{"type":"Point","coordinates":[-122.4194,37.7749]}`
	minSpacing := 0.5

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	nearbyPlants := []*entity.GardenPlant{
		{GardenPlantID: "gp-1", GardenID: gardenID},
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("CheckPlantSpacing", ctx, gardenID, location, minSpacing).Return(nearbyPlants, nil)

	result, err := service.CheckPlantSpacing(ctx, gardenID, location, minSpacing)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

// Test BulkPlacePlants
func TestPlantPlacementService_BulkPlacePlants_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	plantID := "plant-123"
	location := `{"type":"Point","coordinates":[-122.4194,37.7749]}`

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	plants := []*entity.GardenPlant{
		{GardenID: gardenID, PlantID: plantID, LocationGeoJSON: location},
		{GardenID: gardenID, PlantID: plantID, LocationGeoJSON: location},
		{GardenID: gardenID, PlantID: plantID, LocationGeoJSON: location},
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("ValidatePlantLocation", ctx, gardenID, location).Return(nil).Times(3)
	mockPlantRepo.On("BulkCreate", ctx, mock.AnythingOfType("[]*entity.GardenPlant")).Return(nil)

	result, err := service.BulkPlacePlants(ctx, plants)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	for _, plant := range result {
		assert.NotEmpty(t, plant.GardenPlantID)
		assert.Equal(t, entity.HealthStatusHealthy, *plant.HealthStatus)
	}
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

func TestPlantPlacementService_BulkPlacePlants_DifferentGardens(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	location := `{"type":"Point","coordinates":[-122.4194,37.7749]}`

	plants := []*entity.GardenPlant{
		{GardenID: "garden-123", PlantID: "plant-123", LocationGeoJSON: location},
		{GardenID: "garden-456", PlantID: "plant-123", LocationGeoJSON: location},
	}

	_, err := service.BulkPlacePlants(ctx, plants)

	assert.Error(t, err)
	assert.IsType(t, &entity.ValidationError{}, err)
	assert.Contains(t, err.Error(), "all plants must belong to the same garden")
}

// Test UpdatePlantHealth
func TestPlantPlacementService_UpdatePlantHealth_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenPlantID := "gp-123"
	oldStatus := entity.HealthStatusHealthy
	newStatus := entity.HealthStatusStruggling
	notes := "Leaves turning yellow"

	existingPlant := &entity.GardenPlant{
		GardenPlantID:   gardenPlantID,
		GardenID:        "garden-123",
		PlantID:         "plant-123",
		LocationGeoJSON: "{}",
		HealthStatus:    &oldStatus,
	}

	mockPlantRepo.On("FindByID", ctx, gardenPlantID).Return(existingPlant, nil)
	mockPlantRepo.On("Update", ctx, mock.AnythingOfType("*entity.GardenPlant")).Return(nil)

	err := service.UpdatePlantHealth(ctx, gardenPlantID, newStatus, &notes)

	assert.NoError(t, err)
	mockPlantRepo.AssertExpectations(t)
}

// Test GetPlantingStats
func TestPlantPlacementService_GetPlantingStats_Success(t *testing.T) {
	mockPlantRepo := new(MockGardenPlantRepository)
	mockGardenRepo := new(MockGardenRepository)
	mockZoneRepo := new(MockGardenZoneRepository)
	service := NewPlantPlacementService(mockPlantRepo, mockGardenRepo, mockZoneRepo)

	ctx := context.Background()
	gardenID := "garden-123"

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	healthyStatus := entity.HealthStatusHealthy
	activePlants := []*entity.GardenPlant{
		{GardenPlantID: "gp-1", GardenID: gardenID, HealthStatus: &healthyStatus},
		{GardenPlantID: "gp-2", GardenID: gardenID, HealthStatus: &healthyStatus},
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockPlantRepo.On("CountByGardenID", ctx, gardenID).Return(10, nil)
	mockPlantRepo.On("FindActivePlants", ctx, gardenID).Return(activePlants, nil)
	mockPlantRepo.On("FindByHealthStatus", ctx, gardenID, entity.HealthStatusThriving).Return([]*entity.GardenPlant{}, nil)
	mockPlantRepo.On("FindByHealthStatus", ctx, gardenID, entity.HealthStatusHealthy).Return(activePlants, nil)
	mockPlantRepo.On("FindByHealthStatus", ctx, gardenID, entity.HealthStatusStruggling).Return([]*entity.GardenPlant{}, nil)
	mockPlantRepo.On("FindByHealthStatus", ctx, gardenID, entity.HealthStatusDiseased).Return([]*entity.GardenPlant{}, nil)
	mockPlantRepo.On("FindByHealthStatus", ctx, gardenID, entity.HealthStatusDead).Return([]*entity.GardenPlant{}, nil)

	result, err := service.GetPlantingStats(ctx, gardenID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 10, result.TotalPlants)
	assert.Equal(t, 2, result.ActivePlants)
	assert.Equal(t, 2, result.PlantsByHealth[entity.HealthStatusHealthy])
	mockPlantRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}
