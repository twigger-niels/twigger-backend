package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"twigger-backend/backend/garden-service/domain/entity"
)

// MockGardenZoneRepository is a mock implementation of repository.GardenZoneRepository
type MockGardenZoneRepository struct {
	mock.Mock
}

func (m *MockGardenZoneRepository) Create(ctx context.Context, zone *entity.GardenZone) error {
	args := m.Called(ctx, zone)
	return args.Error(0)
}

func (m *MockGardenZoneRepository) FindByID(ctx context.Context, zoneID string) (*entity.GardenZone, error) {
	args := m.Called(ctx, zoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GardenZone), args.Error(1)
}

func (m *MockGardenZoneRepository) FindByGardenID(ctx context.Context, gardenID string) ([]*entity.GardenZone, error) {
	args := m.Called(ctx, gardenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenZone), args.Error(1)
}

func (m *MockGardenZoneRepository) Update(ctx context.Context, zone *entity.GardenZone) error {
	args := m.Called(ctx, zone)
	return args.Error(0)
}

func (m *MockGardenZoneRepository) Delete(ctx context.Context, zoneID string) error {
	args := m.Called(ctx, zoneID)
	return args.Error(0)
}

func (m *MockGardenZoneRepository) ValidateZoneWithinGarden(ctx context.Context, gardenID, zoneGeometryGeoJSON string) error {
	args := m.Called(ctx, gardenID, zoneGeometryGeoJSON)
	return args.Error(0)
}

func (m *MockGardenZoneRepository) CheckZoneOverlaps(ctx context.Context, gardenID, zoneGeometryGeoJSON string, excludeZoneID *string) (bool, error) {
	args := m.Called(ctx, gardenID, zoneGeometryGeoJSON, excludeZoneID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockGardenZoneRepository) CalculateTotalArea(ctx context.Context, gardenID string) (float64, error) {
	args := m.Called(ctx, gardenID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockGardenZoneRepository) CalculateZoneArea(ctx context.Context, zoneID string) (float64, error) {
	args := m.Called(ctx, zoneID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockGardenZoneRepository) CountByGardenID(ctx context.Context, gardenID string) (int, error) {
	args := m.Called(ctx, gardenID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockGardenZoneRepository) FindByIDs(ctx context.Context, zoneIDs []string) ([]*entity.GardenZone, error) {
	args := m.Called(ctx, zoneIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GardenZone), args.Error(1)
}

// Test CreateZone
func TestZoneManagementService_CreateZone_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	zoneGeometry := `{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.49,37.7],[-122.49,37.71],[-122.5,37.71],[-122.5,37.7]]]}`
	zoneType := entity.ZoneTypeBed

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: zoneGeometry,
		ZoneType:        &zoneType,
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockZoneRepo.On("ValidateZoneWithinGarden", ctx, gardenID, zoneGeometry).Return(nil)
	mockZoneRepo.On("CheckZoneOverlaps", ctx, gardenID, zoneGeometry, (*string)(nil)).Return(false, nil)
	mockZoneRepo.On("Create", ctx, mock.AnythingOfType("*entity.GardenZone")).Return(nil)

	result, err := service.CreateZone(ctx, zone)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ZoneID)
	assert.False(t, result.CreatedAt.IsZero())
	mockZoneRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

func TestZoneManagementService_CreateZone_GardenNotFound(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	gardenID := "non-existent"
	zoneType := entity.ZoneTypeBed

	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: "{}",
		ZoneType:        &zoneType,
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(nil, entity.NewNotFoundError("garden", gardenID))

	_, err := service.CreateZone(ctx, zone)

	assert.Error(t, err)
	mockGardenRepo.AssertExpectations(t)
}

func TestZoneManagementService_CreateZone_NotWithinBoundary(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	zoneGeometry := `{"type":"Polygon","coordinates":[]}`
	zoneType := entity.ZoneTypeBed

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: zoneGeometry,
		ZoneType:        &zoneType,
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockZoneRepo.On("ValidateZoneWithinGarden", ctx, gardenID, zoneGeometry).Return(errors.New("zone outside boundary"))

	_, err := service.CreateZone(ctx, zone)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zone must be within garden boundary")
	mockZoneRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

func TestZoneManagementService_CreateZone_Overlaps(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	zoneGeometry := `{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.49,37.7],[-122.49,37.71],[-122.5,37.71],[-122.5,37.7]]]}`
	zoneType := entity.ZoneTypeBed

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	zone := &entity.GardenZone{
		GardenID:        gardenID,
		GeometryGeoJSON: zoneGeometry,
		ZoneType:        &zoneType,
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockZoneRepo.On("ValidateZoneWithinGarden", ctx, gardenID, zoneGeometry).Return(nil)
	mockZoneRepo.On("CheckZoneOverlaps", ctx, gardenID, zoneGeometry, (*string)(nil)).Return(true, nil)

	_, err := service.CreateZone(ctx, zone)

	assert.Error(t, err)
	assert.IsType(t, &entity.ValidationError{}, err)
	assert.Contains(t, err.Error(), "overlaps")
	mockZoneRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

// Test GetZone
func TestZoneManagementService_GetZone_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	zoneID := "zone-123"
	zoneType := entity.ZoneTypeBed
	expectedZone := &entity.GardenZone{
		ZoneID:          zoneID,
		GardenID:        "garden-123",
		GeometryGeoJSON: "{}",
		ZoneType:        &zoneType,
	}

	mockZoneRepo.On("FindByID", ctx, zoneID).Return(expectedZone, nil)

	result, err := service.GetZone(ctx, zoneID)

	assert.NoError(t, err)
	assert.Equal(t, expectedZone, result)
	mockZoneRepo.AssertExpectations(t)
}

func TestZoneManagementService_GetZone_EmptyID(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()

	_, err := service.GetZone(ctx, "")

	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)
}

// Test ListGardenZones
func TestZoneManagementService_ListGardenZones_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	gardenID := "garden-123"

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	zoneType := entity.ZoneTypeBed
	expectedZones := []*entity.GardenZone{
		{ZoneID: "zone-1", GardenID: gardenID, ZoneType: &zoneType},
		{ZoneID: "zone-2", GardenID: gardenID, ZoneType: &zoneType},
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockZoneRepo.On("FindByGardenID", ctx, gardenID).Return(expectedZones, nil)

	result, err := service.ListGardenZones(ctx, gardenID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockZoneRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

// Test UpdateZone
func TestZoneManagementService_UpdateZone_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	zoneID := "zone-123"
	gardenID := "garden-123"
	oldGeometry := `{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.49,37.7],[-122.49,37.71],[-122.5,37.71],[-122.5,37.7]]]}`
	newGeometry := `{"type":"Polygon","coordinates":[[[-122.51,37.71],[-122.50,37.71],[-122.50,37.72],[-122.51,37.72],[-122.51,37.71]]]}`

	createdAt := time.Now().Add(-24 * time.Hour)
	zoneType := entity.ZoneTypeBed

	existingZone := &entity.GardenZone{
		ZoneID:          zoneID,
		GardenID:        gardenID,
		GeometryGeoJSON: oldGeometry,
		ZoneType:        &zoneType,
		CreatedAt:       createdAt,
	}

	updatedZone := &entity.GardenZone{
		ZoneID:          zoneID,
		GardenID:        gardenID,
		GeometryGeoJSON: newGeometry,
		ZoneType:        &zoneType,
	}

	mockZoneRepo.On("FindByID", ctx, zoneID).Return(existingZone, nil)
	mockZoneRepo.On("ValidateZoneWithinGarden", ctx, gardenID, newGeometry).Return(nil)
	mockZoneRepo.On("CheckZoneOverlaps", ctx, gardenID, newGeometry, &zoneID).Return(false, nil)
	mockZoneRepo.On("Update", ctx, mock.AnythingOfType("*entity.GardenZone")).Return(nil)

	result, err := service.UpdateZone(ctx, updatedZone)

	assert.NoError(t, err)
	assert.Equal(t, createdAt, result.CreatedAt) // Preserved
	mockZoneRepo.AssertExpectations(t)
}

// Test DeleteZone
func TestZoneManagementService_DeleteZone_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	zoneID := "zone-123"
	zoneType := entity.ZoneTypeBed

	existingZone := &entity.GardenZone{
		ZoneID:   zoneID,
		GardenID: "garden-123",
		ZoneType: &zoneType,
	}

	mockZoneRepo.On("FindByID", ctx, zoneID).Return(existingZone, nil)
	mockZoneRepo.On("Delete", ctx, zoneID).Return(nil)

	err := service.DeleteZone(ctx, zoneID)

	assert.NoError(t, err)
	mockZoneRepo.AssertExpectations(t)
}

// Test CalculateZoneArea
func TestZoneManagementService_CalculateZoneArea_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	zoneID := "zone-123"
	expectedArea := 250.5

	mockZoneRepo.On("CalculateArea", ctx, zoneID).Return(expectedArea, nil)

	result, err := service.CalculateZoneArea(ctx, zoneID)

	assert.NoError(t, err)
	assert.Equal(t, expectedArea, result)
	mockZoneRepo.AssertExpectations(t)
}

// Test GetTotalZoneArea
func TestZoneManagementService_GetTotalZoneArea_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	expectedTotalArea := 750.0

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockZoneRepo.On("CalculateTotalArea", ctx, gardenID).Return(expectedTotalArea, nil)

	result, err := service.GetTotalZoneArea(ctx, gardenID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotalArea, result)
	mockZoneRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}

// Test CheckZoneOverlaps
func TestZoneManagementService_CheckZoneOverlaps_Success(t *testing.T) {
	mockZoneRepo := new(MockGardenZoneRepository)
	mockGardenRepo := new(MockGardenRepository)
	service := NewZoneManagementService(mockZoneRepo, mockGardenRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	zoneGeometry := `{"type":"Polygon","coordinates":[]}`

	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		GardenName: "Test Garden",
	}

	mockGardenRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockZoneRepo.On("CheckZoneOverlaps", ctx, gardenID, zoneGeometry, (*string)(nil)).Return(true, nil)

	result, err := service.CheckZoneOverlaps(ctx, gardenID, zoneGeometry, nil)

	assert.NoError(t, err)
	assert.True(t, result)
	mockZoneRepo.AssertExpectations(t)
	mockGardenRepo.AssertExpectations(t)
}
