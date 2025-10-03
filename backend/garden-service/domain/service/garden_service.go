package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/backend/garden-service/domain/repository"
)

// GardenService defines the business logic for garden management
type GardenService interface {
	// CreateGarden creates a new garden with automatic hardiness zone detection
	CreateGarden(ctx context.Context, garden *entity.Garden) (*entity.Garden, error)

	// GetGarden retrieves a garden by ID
	GetGarden(ctx context.Context, gardenID string) (*entity.Garden, error)

	// ListUserGardens retrieves all gardens for a user with pagination
	ListUserGardens(ctx context.Context, userID string, limit, offset int) ([]*entity.Garden, error)

	// UpdateGarden updates an existing garden with validation
	UpdateGarden(ctx context.Context, garden *entity.Garden) (*entity.Garden, error)

	// DeleteGarden deletes a garden and all associated data
	DeleteGarden(ctx context.Context, gardenID string) error

	// CalculateGardenArea calculates the area of a garden in square meters
	CalculateGardenArea(ctx context.Context, gardenID string) (float64, error)

	// DetectClimateZone detects the hardiness zone for a garden based on location
	DetectClimateZone(ctx context.Context, gardenID string) (string, error)

	// FindNearbyGardens finds gardens within a radius of a location
	FindNearbyGardens(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Garden, error)

	// GetGardenStats retrieves statistics for a user's gardens
	GetGardenStats(ctx context.Context, userID string) (*GardenStats, error)

	// ValidateGardenBoundary validates that a garden boundary is valid GeoJSON
	ValidateGardenBoundary(ctx context.Context, boundaryGeoJSON string) error
}

// GardenStats holds statistics about a user's gardens
type GardenStats struct {
	TotalGardens int     `json:"total_gardens"`
	TotalAreaM2  float64 `json:"total_area_m2"`
}

// gardenService implements GardenService
type gardenService struct {
	gardenRepo repository.GardenRepository
}

// NewGardenService creates a new garden service instance
func NewGardenService(gardenRepo repository.GardenRepository) GardenService {
	return &gardenService{
		gardenRepo: gardenRepo,
	}
}

// CreateGarden creates a new garden with automatic hardiness zone detection
func (s *gardenService) CreateGarden(ctx context.Context, garden *entity.Garden) (*entity.Garden, error) {
	// Validate entity
	if err := garden.Validate(); err != nil {
		return nil, entity.NewValidationError("garden", err.Error())
	}

	// Validate boundary if provided
	if garden.BoundaryGeoJSON != nil {
		if err := s.gardenRepo.ValidateBoundary(ctx, *garden.BoundaryGeoJSON); err != nil {
			return nil, fmt.Errorf("invalid garden boundary: %w", err)
		}
	}

	// Generate ID if not provided
	if garden.GardenID == "" {
		garden.GardenID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	garden.CreatedAt = now
	garden.UpdatedAt = now

	// Create garden
	if err := s.gardenRepo.Create(ctx, garden); err != nil {
		return nil, fmt.Errorf("failed to create garden: %w", err)
	}

	// Attempt to detect hardiness zone if location provided
	if garden.LocationGeoJSON != nil || garden.BoundaryGeoJSON != nil {
		zone, err := s.gardenRepo.DetectHardinessZone(ctx, garden.GardenID)
		if err == nil && zone != "" {
			garden.HardinessZone = &zone
			// Update garden with detected zone
			if err := s.gardenRepo.Update(ctx, garden); err != nil {
				// Log error but don't fail creation
				// TODO: Add logging when logger is available
			}
		}
	}

	return garden, nil
}

// GetGarden retrieves a garden by ID
func (s *gardenService) GetGarden(ctx context.Context, gardenID string) (*entity.Garden, error) {
	if gardenID == "" {
		return nil, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	garden, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get garden: %w", err)
	}

	return garden, nil
}

// ListUserGardens retrieves all gardens for a user with pagination
func (s *gardenService) ListUserGardens(ctx context.Context, userID string, limit, offset int) ([]*entity.Garden, error) {
	if userID == "" {
		return nil, entity.NewInvalidInputError("user_id", "user ID cannot be empty")
	}

	if limit <= 0 {
		limit = 10 // Default limit
	}

	if limit > 100 {
		limit = 100 // Maximum limit
	}

	if offset < 0 {
		offset = 0
	}

	gardens, err := s.gardenRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list gardens: %w", err)
	}

	return gardens, nil
}

// UpdateGarden updates an existing garden with validation
func (s *gardenService) UpdateGarden(ctx context.Context, garden *entity.Garden) (*entity.Garden, error) {
	// Validate entity
	if err := garden.Validate(); err != nil {
		return nil, entity.NewValidationError("garden", err.Error())
	}

	// Check if garden exists
	existing, err := s.gardenRepo.FindByID(ctx, garden.GardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	// Validate boundary if changed
	if garden.BoundaryGeoJSON != nil &&
	   (existing.BoundaryGeoJSON == nil || *garden.BoundaryGeoJSON != *existing.BoundaryGeoJSON) {
		if err := s.gardenRepo.ValidateBoundary(ctx, *garden.BoundaryGeoJSON); err != nil {
			return nil, fmt.Errorf("invalid garden boundary: %w", err)
		}

		// Re-detect hardiness zone if boundary changed
		zone, err := s.gardenRepo.DetectHardinessZone(ctx, garden.GardenID)
		if err == nil && zone != "" {
			garden.HardinessZone = &zone
		}
	}

	// Preserve created_at, update updated_at
	garden.CreatedAt = existing.CreatedAt
	garden.UpdatedAt = time.Now()

	if err := s.gardenRepo.Update(ctx, garden); err != nil {
		return nil, fmt.Errorf("failed to update garden: %w", err)
	}

	return garden, nil
}

// DeleteGarden deletes a garden and all associated data
func (s *gardenService) DeleteGarden(ctx context.Context, gardenID string) error {
	if gardenID == "" {
		return entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return fmt.Errorf("garden not found: %w", err)
	}

	// Delete garden (CASCADE will delete zones, features, plants)
	if err := s.gardenRepo.Delete(ctx, gardenID); err != nil {
		return fmt.Errorf("failed to delete garden: %w", err)
	}

	return nil
}

// CalculateGardenArea calculates the area of a garden in square meters
func (s *gardenService) CalculateGardenArea(ctx context.Context, gardenID string) (float64, error) {
	if gardenID == "" {
		return 0, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	area, err := s.gardenRepo.CalculateArea(ctx, gardenID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate area: %w", err)
	}

	return area, nil
}

// DetectClimateZone detects the hardiness zone for a garden based on location
func (s *gardenService) DetectClimateZone(ctx context.Context, gardenID string) (string, error) {
	if gardenID == "" {
		return "", entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	zone, err := s.gardenRepo.DetectHardinessZone(ctx, gardenID)
	if err != nil {
		return "", fmt.Errorf("failed to detect climate zone: %w", err)
	}

	if zone == "" {
		return "", entity.NewNotFoundError("climate_zone", "no climate zone found for garden location")
	}

	return zone, nil
}

// FindNearbyGardens finds gardens within a radius of a location
func (s *gardenService) FindNearbyGardens(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Garden, error) {
	// Validate coordinates
	if lat < -90 || lat > 90 {
		return nil, entity.NewInvalidInputError("latitude", "latitude must be between -90 and 90")
	}

	if lng < -180 || lng > 180 {
		return nil, entity.NewInvalidInputError("longitude", "longitude must be between -180 and 180")
	}

	if radiusKm <= 0 {
		return nil, entity.NewInvalidInputError("radius", "radius must be greater than 0")
	}

	if radiusKm > 100 {
		radiusKm = 100 // Cap at 100km for performance
	}

	gardens, err := s.gardenRepo.FindByLocation(ctx, lat, lng, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby gardens: %w", err)
	}

	return gardens, nil
}

// GetGardenStats retrieves statistics for a user's gardens
func (s *gardenService) GetGardenStats(ctx context.Context, userID string) (*GardenStats, error) {
	if userID == "" {
		return nil, entity.NewInvalidInputError("user_id", "user ID cannot be empty")
	}

	count, err := s.gardenRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count gardens: %w", err)
	}

	totalArea, err := s.gardenRepo.GetTotalArea(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total area: %w", err)
	}

	return &GardenStats{
		TotalGardens: count,
		TotalAreaM2:  totalArea,
	}, nil
}

// ValidateGardenBoundary validates that a garden boundary is valid GeoJSON
func (s *gardenService) ValidateGardenBoundary(ctx context.Context, boundaryGeoJSON string) error {
	if boundaryGeoJSON == "" {
		return entity.NewInvalidInputError("boundary", "boundary GeoJSON cannot be empty")
	}

	if err := s.gardenRepo.ValidateBoundary(ctx, boundaryGeoJSON); err != nil {
		return fmt.Errorf("invalid boundary: %w", err)
	}

	return nil
}
