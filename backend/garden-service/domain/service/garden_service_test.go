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

// MockGardenRepository is a mock implementation of repository.GardenRepository
type MockGardenRepository struct {
	mock.Mock
}

func (m *MockGardenRepository) Create(ctx context.Context, garden *entity.Garden) error {
	args := m.Called(ctx, garden)
	return args.Error(0)
}

func (m *MockGardenRepository) FindByID(ctx context.Context, gardenID string) (*entity.Garden, error) {
	args := m.Called(ctx, gardenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Garden), args.Error(1)
}

func (m *MockGardenRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Garden, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Garden), args.Error(1)
}

func (m *MockGardenRepository) FindByIDs(ctx context.Context, gardenIDs []string) ([]*entity.Garden, error) {
	args := m.Called(ctx, gardenIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Garden), args.Error(1)
}

func (m *MockGardenRepository) Update(ctx context.Context, garden *entity.Garden) error {
	args := m.Called(ctx, garden)
	return args.Error(0)
}

func (m *MockGardenRepository) Delete(ctx context.Context, gardenID string) error {
	args := m.Called(ctx, gardenID)
	return args.Error(0)
}

func (m *MockGardenRepository) FindByLocation(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Garden, error) {
	args := m.Called(ctx, lat, lng, radiusKm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Garden), args.Error(1)
}

func (m *MockGardenRepository) CalculateArea(ctx context.Context, gardenID string) (float64, error) {
	args := m.Called(ctx, gardenID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockGardenRepository) GetCenterPoint(ctx context.Context, gardenID string) (float64, float64, error) {
	args := m.Called(ctx, gardenID)
	return args.Get(0).(float64), args.Get(1).(float64), args.Error(2)
}

func (m *MockGardenRepository) DetectHardinessZone(ctx context.Context, gardenID string) (string, error) {
	args := m.Called(ctx, gardenID)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGardenRepository) ValidateBoundary(ctx context.Context, boundaryGeoJSON string) error {
	args := m.Called(ctx, boundaryGeoJSON)
	return args.Error(0)
}

func (m *MockGardenRepository) CheckBoundaryValid(ctx context.Context, boundaryGeoJSON string) (bool, error) {
	args := m.Called(ctx, boundaryGeoJSON)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockGardenRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockGardenRepository) GetTotalArea(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

// Test CreateGarden
func TestGardenService_CreateGarden_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	validBoundary := `{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.4,37.7],[-122.4,37.8],[-122.5,37.8],[-122.5,37.7]]]}`
	gardenName := "My Garden"
	garden := &entity.Garden{
		UserID:          "user-123",
		GardenName:      gardenName,
		BoundaryGeoJSON: &validBoundary,
	}

	mockRepo.On("ValidateBoundary", ctx, validBoundary).Return(nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Garden")).Return(nil)
	mockRepo.On("DetectHardinessZone", ctx, mock.AnythingOfType("string")).Return("9a", nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.Garden")).Return(nil)

	result, err := service.CreateGarden(ctx, garden)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.GardenID)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestGardenService_CreateGarden_ValidationError(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	invalidGarden := &entity.Garden{
		UserID:     "user-123",
		GardenName: "", // Empty name - invalid
	}

	_, err := service.CreateGarden(ctx, invalidGarden)

	assert.Error(t, err)
	assert.IsType(t, &entity.ValidationError{}, err)
}

func TestGardenService_CreateGarden_InvalidBoundary(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	invalidBoundary := `{"type":"Polygon","coordinates":[]}`
	gardenName := "My Garden"
	garden := &entity.Garden{
		UserID:          "user-123",
		GardenName:      gardenName,
		BoundaryGeoJSON: &invalidBoundary,
	}

	mockRepo.On("ValidateBoundary", ctx, invalidBoundary).Return(errors.New("invalid boundary"))

	_, err := service.CreateGarden(ctx, garden)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid garden boundary")
	mockRepo.AssertExpectations(t)
}

// Test GetGarden
func TestGardenService_GetGarden_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	expectedGarden := &entity.Garden{
		GardenID:   gardenID,
		UserID:     "user-123",
		GardenName: "My Garden",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockRepo.On("FindByID", ctx, gardenID).Return(expectedGarden, nil)

	result, err := service.GetGarden(ctx, gardenID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGarden, result)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_GetGarden_EmptyID(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()

	_, err := service.GetGarden(ctx, "")

	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)
}

func TestGardenService_GetGarden_NotFound(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	gardenID := "non-existent"

	mockRepo.On("FindByID", ctx, gardenID).Return(nil, entity.NewNotFoundError("garden", gardenID))

	_, err := service.GetGarden(ctx, gardenID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// Test ListUserGardens
func TestGardenService_ListUserGardens_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	userID := "user-123"
	expectedGardens := []*entity.Garden{
		{GardenID: "garden-1", UserID: userID, GardenName: "Garden 1"},
		{GardenID: "garden-2", UserID: userID, GardenName: "Garden 2"},
	}

	mockRepo.On("FindByUserID", ctx, userID, 10, 0).Return(expectedGardens, nil)

	result, err := service.ListUserGardens(ctx, userID, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_ListUserGardens_DefaultLimit(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("FindByUserID", ctx, userID, 10, 0).Return([]*entity.Garden{}, nil)

	_, err := service.ListUserGardens(ctx, userID, 0, 0) // Zero limit

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_ListUserGardens_MaxLimit(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("FindByUserID", ctx, userID, 100, 0).Return([]*entity.Garden{}, nil)

	_, err := service.ListUserGardens(ctx, userID, 200, 0) // Over max limit

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test UpdateGarden
func TestGardenService_UpdateGarden_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	createdAt := time.Now().Add(-24 * time.Hour)
	existingGarden := &entity.Garden{
		GardenID:   "garden-123",
		UserID:     "user-123",
		GardenName: "Old Name",
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}

	updatedGarden := &entity.Garden{
		GardenID:   "garden-123",
		UserID:     "user-123",
		GardenName: "New Name",
	}

	mockRepo.On("FindByID", ctx, "garden-123").Return(existingGarden, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.Garden")).Return(nil)

	result, err := service.UpdateGarden(ctx, updatedGarden)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.GardenName)
	assert.Equal(t, createdAt, result.CreatedAt) // Preserved
	assert.True(t, result.UpdatedAt.After(createdAt)) // Updated
	mockRepo.AssertExpectations(t)
}

func TestGardenService_UpdateGarden_NotFound(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	updatedGarden := &entity.Garden{
		GardenID:   "non-existent",
		UserID:     "user-123",
		GardenName: "Name",
	}

	mockRepo.On("FindByID", ctx, "non-existent").Return(nil, entity.NewNotFoundError("garden", "non-existent"))

	_, err := service.UpdateGarden(ctx, updatedGarden)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// Test DeleteGarden
func TestGardenService_DeleteGarden_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	existingGarden := &entity.Garden{
		GardenID:   gardenID,
		UserID:     "user-123",
		GardenName: "Garden",
	}

	mockRepo.On("FindByID", ctx, gardenID).Return(existingGarden, nil)
	mockRepo.On("Delete", ctx, gardenID).Return(nil)

	err := service.DeleteGarden(ctx, gardenID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_DeleteGarden_EmptyID(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()

	err := service.DeleteGarden(ctx, "")

	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)
}

// Test CalculateGardenArea
func TestGardenService_CalculateGardenArea_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	expectedArea := 1500.0

	mockRepo.On("CalculateArea", ctx, gardenID).Return(expectedArea, nil)

	result, err := service.CalculateGardenArea(ctx, gardenID)

	assert.NoError(t, err)
	assert.Equal(t, expectedArea, result)
	mockRepo.AssertExpectations(t)
}

// Test DetectClimateZone
func TestGardenService_DetectClimateZone_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	gardenID := "garden-123"
	expectedZone := "9a"

	mockRepo.On("DetectHardinessZone", ctx, gardenID).Return(expectedZone, nil)

	result, err := service.DetectClimateZone(ctx, gardenID)

	assert.NoError(t, err)
	assert.Equal(t, expectedZone, result)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_DetectClimateZone_NotFound(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	gardenID := "garden-123"

	mockRepo.On("DetectHardinessZone", ctx, gardenID).Return("", nil)

	_, err := service.DetectClimateZone(ctx, gardenID)

	assert.Error(t, err)
	assert.IsType(t, &entity.NotFoundError{}, err)
	mockRepo.AssertExpectations(t)
}

// Test FindNearbyGardens
func TestGardenService_FindNearbyGardens_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	lat, lng, radiusKm := 37.7749, -122.4194, 5.0
	expectedGardens := []*entity.Garden{
		{GardenID: "garden-1", GardenName: "Nearby Garden 1"},
		{GardenID: "garden-2", GardenName: "Nearby Garden 2"},
	}

	mockRepo.On("FindByLocation", ctx, lat, lng, radiusKm).Return(expectedGardens, nil)

	result, err := service.FindNearbyGardens(ctx, lat, lng, radiusKm)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_FindNearbyGardens_InvalidCoordinates(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()

	// Invalid latitude
	_, err := service.FindNearbyGardens(ctx, 95.0, -122.0, 5.0)
	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)

	// Invalid longitude
	_, err = service.FindNearbyGardens(ctx, 37.0, -200.0, 5.0)
	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)

	// Invalid radius
	_, err = service.FindNearbyGardens(ctx, 37.0, -122.0, -5.0)
	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)
}

func TestGardenService_FindNearbyGardens_RadiusCapped(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	lat, lng := 37.7749, -122.4194

	// Radius capped at 100km
	mockRepo.On("FindByLocation", ctx, lat, lng, 100.0).Return([]*entity.Garden{}, nil)

	_, err := service.FindNearbyGardens(ctx, lat, lng, 200.0)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test GetGardenStats
func TestGardenService_GetGardenStats_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	userID := "user-123"

	mockRepo.On("CountByUserID", ctx, userID).Return(5, nil)
	mockRepo.On("GetTotalArea", ctx, userID).Return(7500.0, nil)

	result, err := service.GetGardenStats(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 5, result.TotalGardens)
	assert.Equal(t, 7500.0, result.TotalAreaM2)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_GetGardenStats_EmptyUserID(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()

	_, err := service.GetGardenStats(ctx, "")

	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)
}

// Test ValidateGardenBoundary
func TestGardenService_ValidateGardenBoundary_Success(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()
	validBoundary := `{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.4,37.7],[-122.4,37.8],[-122.5,37.8],[-122.5,37.7]]]}`

	mockRepo.On("ValidateBoundary", ctx, validBoundary).Return(nil)

	err := service.ValidateGardenBoundary(ctx, validBoundary)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGardenService_ValidateGardenBoundary_EmptyBoundary(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := NewGardenService(mockRepo)

	ctx := context.Background()

	err := service.ValidateGardenBoundary(ctx, "")

	assert.Error(t, err)
	assert.IsType(t, &entity.InvalidInputError{}, err)
}
