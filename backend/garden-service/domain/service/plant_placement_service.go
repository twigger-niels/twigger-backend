package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/backend/garden-service/domain/repository"
)

// PlantPlacementService defines the business logic for managing plant placements
type PlantPlacementService interface {
	// PlacePlant places a plant in a garden with validation
	PlacePlant(ctx context.Context, gardenPlant *entity.GardenPlant) (*entity.GardenPlant, error)

	// GetGardenPlant retrieves a garden plant by ID
	GetGardenPlant(ctx context.Context, gardenPlantID string) (*entity.GardenPlant, error)

	// ListGardenPlants retrieves all plants in a garden with optional filters
	ListGardenPlants(ctx context.Context, gardenID string, filter *GardenPlantFilter) ([]*entity.GardenPlant, error)

	// UpdatePlantPlacement updates a plant placement with validation
	UpdatePlantPlacement(ctx context.Context, gardenPlant *entity.GardenPlant) (*entity.GardenPlant, error)

	// RemovePlant removes a plant from a garden
	RemovePlant(ctx context.Context, gardenPlantID string) error

	// CheckPlantSpacing checks if a plant location maintains minimum spacing
	CheckPlantSpacing(ctx context.Context, gardenID, locationGeoJSON string, minSpacingM float64) ([]*entity.GardenPlant, error)

	// FindPlantsInZone finds all plants within a garden zone
	FindPlantsInZone(ctx context.Context, zoneID string) ([]*entity.GardenPlant, error)

	// BulkPlacePlants places multiple plants at once with validation
	BulkPlacePlants(ctx context.Context, gardenPlants []*entity.GardenPlant) ([]*entity.GardenPlant, error)

	// UpdatePlantHealth updates the health status of a plant
	UpdatePlantHealth(ctx context.Context, gardenPlantID string, healthStatus entity.HealthStatus, notes *string) error

	// GetPlantingStats retrieves planting statistics for a garden
	GetPlantingStats(ctx context.Context, gardenID string) (*PlantingStats, error)
}

// GardenPlantFilter defines filters for listing garden plants
type GardenPlantFilter struct {
	ZoneID       *string             `json:"zone_id,omitempty"`
	HealthStatus *entity.HealthStatus `json:"health_status,omitempty"`
	ActiveOnly   bool                `json:"active_only"`
}

// PlantingStats holds statistics about plants in a garden
type PlantingStats struct {
	TotalPlants    int                         `json:"total_plants"`
	ActivePlants   int                         `json:"active_plants"`
	PlantsByHealth map[entity.HealthStatus]int `json:"plants_by_health"`
}

// plantPlacementService implements PlantPlacementService
type plantPlacementService struct {
	gardenPlantRepo repository.GardenPlantRepository
	gardenRepo      repository.GardenRepository
	zoneRepo        repository.GardenZoneRepository
}

// NewPlantPlacementService creates a new plant placement service instance
func NewPlantPlacementService(
	gardenPlantRepo repository.GardenPlantRepository,
	gardenRepo repository.GardenRepository,
	zoneRepo repository.GardenZoneRepository,
) PlantPlacementService {
	return &plantPlacementService{
		gardenPlantRepo: gardenPlantRepo,
		gardenRepo:      gardenRepo,
		zoneRepo:        zoneRepo,
	}
}

// PlacePlant places a plant in a garden with validation
func (s *plantPlacementService) PlacePlant(ctx context.Context, gardenPlant *entity.GardenPlant) (*entity.GardenPlant, error) {
	// Validate entity
	if err := gardenPlant.Validate(); err != nil {
		return nil, entity.NewValidationError("garden_plant", err.Error())
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenPlant.GardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	// Validate plant location is within garden boundary
	if err := s.gardenPlantRepo.ValidatePlantLocation(ctx, gardenPlant.GardenID, gardenPlant.LocationGeoJSON); err != nil {
		return nil, fmt.Errorf("plant location must be within garden boundary: %w", err)
	}

	// Check if zone_id provided and validate it exists in this garden
	if gardenPlant.ZoneID != nil {
		zone, err := s.zoneRepo.FindByID(ctx, *gardenPlant.ZoneID)
		if err != nil {
			return nil, fmt.Errorf("zone not found: %w", err)
		}
		if zone.GardenID != gardenPlant.GardenID {
			return nil, entity.NewValidationError("garden_plant", "zone must belong to the same garden")
		}
	}

	// Generate ID if not provided
	if gardenPlant.GardenPlantID == "" {
		gardenPlant.GardenPlantID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	gardenPlant.PlantedAt = now
	gardenPlant.CreatedAt = now
	gardenPlant.UpdatedAt = now

	// Set default health status if not provided
	if gardenPlant.HealthStatus == nil {
		healthy := entity.HealthStatusHealthy
		gardenPlant.HealthStatus = &healthy
	}

	// Create garden plant
	if err := s.gardenPlantRepo.Create(ctx, gardenPlant); err != nil {
		return nil, fmt.Errorf("failed to place plant: %w", err)
	}

	return gardenPlant, nil
}

// GetGardenPlant retrieves a garden plant by ID
func (s *plantPlacementService) GetGardenPlant(ctx context.Context, gardenPlantID string) (*entity.GardenPlant, error) {
	if gardenPlantID == "" {
		return nil, entity.NewInvalidInputError("garden_plant_id", "garden plant ID cannot be empty")
	}

	gardenPlant, err := s.gardenPlantRepo.FindByID(ctx, gardenPlantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get garden plant: %w", err)
	}

	return gardenPlant, nil
}

// ListGardenPlants retrieves all plants in a garden with optional filters
func (s *plantPlacementService) ListGardenPlants(ctx context.Context, gardenID string, filter *GardenPlantFilter) ([]*entity.GardenPlant, error) {
	if gardenID == "" {
		return nil, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	// Apply filters
	if filter != nil {
		if filter.ZoneID != nil {
			return s.gardenPlantRepo.FindInZone(ctx, *filter.ZoneID)
		}

		if filter.HealthStatus != nil {
			return s.gardenPlantRepo.FindByHealthStatus(ctx, gardenID, *filter.HealthStatus)
		}

		if filter.ActiveOnly {
			return s.gardenPlantRepo.FindActivePlants(ctx, gardenID)
		}
	}

	// No filters - return all plants
	plants, err := s.gardenPlantRepo.FindByGardenID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("failed to list garden plants: %w", err)
	}

	return plants, nil
}

// UpdatePlantPlacement updates a plant placement with validation
func (s *plantPlacementService) UpdatePlantPlacement(ctx context.Context, gardenPlant *entity.GardenPlant) (*entity.GardenPlant, error) {
	// Validate entity
	if err := gardenPlant.Validate(); err != nil {
		return nil, entity.NewValidationError("garden_plant", err.Error())
	}

	// Check if garden plant exists
	existing, err := s.gardenPlantRepo.FindByID(ctx, gardenPlant.GardenPlantID)
	if err != nil {
		return nil, fmt.Errorf("garden plant not found: %w", err)
	}

	// Validate plant location if changed
	if gardenPlant.LocationGeoJSON != existing.LocationGeoJSON {
		if err := s.gardenPlantRepo.ValidatePlantLocation(ctx, gardenPlant.GardenID, gardenPlant.LocationGeoJSON); err != nil {
			return nil, fmt.Errorf("plant location must be within garden boundary: %w", err)
		}
	}

	// Validate zone if changed
	if gardenPlant.ZoneID != nil && (existing.ZoneID == nil || *gardenPlant.ZoneID != *existing.ZoneID) {
		zone, err := s.zoneRepo.FindByID(ctx, *gardenPlant.ZoneID)
		if err != nil {
			return nil, fmt.Errorf("zone not found: %w", err)
		}
		if zone.GardenID != gardenPlant.GardenID {
			return nil, entity.NewValidationError("garden_plant", "zone must belong to the same garden")
		}
	}

	// Preserve created_at and planted_at
	gardenPlant.CreatedAt = existing.CreatedAt
	gardenPlant.PlantedAt = existing.PlantedAt
	gardenPlant.UpdatedAt = time.Now()

	if err := s.gardenPlantRepo.Update(ctx, gardenPlant); err != nil {
		return nil, fmt.Errorf("failed to update plant placement: %w", err)
	}

	return gardenPlant, nil
}

// RemovePlant removes a plant from a garden
func (s *plantPlacementService) RemovePlant(ctx context.Context, gardenPlantID string) error {
	if gardenPlantID == "" {
		return entity.NewInvalidInputError("garden_plant_id", "garden plant ID cannot be empty")
	}

	// Check if garden plant exists
	_, err := s.gardenPlantRepo.FindByID(ctx, gardenPlantID)
	if err != nil {
		return fmt.Errorf("garden plant not found: %w", err)
	}

	// Delete garden plant
	if err := s.gardenPlantRepo.Delete(ctx, gardenPlantID); err != nil {
		return fmt.Errorf("failed to remove plant: %w", err)
	}

	return nil
}

// CheckPlantSpacing checks if a plant location maintains minimum spacing
func (s *plantPlacementService) CheckPlantSpacing(ctx context.Context, gardenID, locationGeoJSON string, minSpacingM float64) ([]*entity.GardenPlant, error) {
	if gardenID == "" {
		return nil, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	if locationGeoJSON == "" {
		return nil, entity.NewInvalidInputError("location", "location GeoJSON cannot be empty")
	}

	if minSpacingM <= 0 {
		return nil, entity.NewInvalidInputError("min_spacing", "minimum spacing must be greater than 0")
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	// Find nearby plants
	nearbyPlants, err := s.gardenPlantRepo.CheckPlantSpacing(ctx, gardenID, locationGeoJSON, minSpacingM)
	if err != nil {
		return nil, fmt.Errorf("failed to check plant spacing: %w", err)
	}

	return nearbyPlants, nil
}

// FindPlantsInZone finds all plants within a garden zone
func (s *plantPlacementService) FindPlantsInZone(ctx context.Context, zoneID string) ([]*entity.GardenPlant, error) {
	if zoneID == "" {
		return nil, entity.NewInvalidInputError("zone_id", "zone ID cannot be empty")
	}

	// Check if zone exists
	_, err := s.zoneRepo.FindByID(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	plants, err := s.gardenPlantRepo.FindInZone(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to find plants in zone: %w", err)
	}

	return plants, nil
}

// BulkPlacePlants places multiple plants at once with validation
func (s *plantPlacementService) BulkPlacePlants(ctx context.Context, gardenPlants []*entity.GardenPlant) ([]*entity.GardenPlant, error) {
	if len(gardenPlants) == 0 {
		return nil, entity.NewInvalidInputError("garden_plants", "garden plants list cannot be empty")
	}

	// Validate all plants belong to the same garden
	gardenID := gardenPlants[0].GardenID
	for _, gp := range gardenPlants {
		if gp.GardenID != gardenID {
			return nil, entity.NewValidationError("garden_plants", "all plants must belong to the same garden")
		}
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	// Validate each plant
	for i, gp := range gardenPlants {
		if err := gp.Validate(); err != nil {
			return nil, fmt.Errorf("plant at index %d is invalid: %w", i, err)
		}

		// Validate location
		if err := s.gardenPlantRepo.ValidatePlantLocation(ctx, gp.GardenID, gp.LocationGeoJSON); err != nil {
			return nil, fmt.Errorf("plant at index %d has invalid location: %w", i, err)
		}

		// Generate ID if not provided
		if gp.GardenPlantID == "" {
			gp.GardenPlantID = uuid.New().String()
		}

		// Set timestamps
		now := time.Now()
		gp.PlantedAt = now
		gp.CreatedAt = now
		gp.UpdatedAt = now

		// Set default health status
		if gp.HealthStatus == nil {
			healthy := entity.HealthStatusHealthy
			gp.HealthStatus = &healthy
		}
	}

	// Bulk create
	if err := s.gardenPlantRepo.BulkCreate(ctx, gardenPlants); err != nil {
		return nil, fmt.Errorf("failed to bulk place plants: %w", err)
	}

	return gardenPlants, nil
}

// UpdatePlantHealth updates the health status of a plant
func (s *plantPlacementService) UpdatePlantHealth(ctx context.Context, gardenPlantID string, healthStatus entity.HealthStatus, notes *string) error {
	if gardenPlantID == "" {
		return entity.NewInvalidInputError("garden_plant_id", "garden plant ID cannot be empty")
	}

	// Check if garden plant exists
	gardenPlant, err := s.gardenPlantRepo.FindByID(ctx, gardenPlantID)
	if err != nil {
		return fmt.Errorf("garden plant not found: %w", err)
	}

	// Update health status
	gardenPlant.HealthStatus = &healthStatus
	if notes != nil {
		gardenPlant.Notes = notes
	}
	gardenPlant.UpdatedAt = time.Now()

	if err := s.gardenPlantRepo.Update(ctx, gardenPlant); err != nil {
		return fmt.Errorf("failed to update plant health: %w", err)
	}

	return nil
}

// GetPlantingStats retrieves planting statistics for a garden
func (s *plantPlacementService) GetPlantingStats(ctx context.Context, gardenID string) (*PlantingStats, error) {
	if gardenID == "" {
		return nil, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	// Get total count
	totalCount, err := s.gardenPlantRepo.CountByGardenID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("failed to count plants: %w", err)
	}

	// Get active plants
	activePlants, err := s.gardenPlantRepo.FindActivePlants(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active plants: %w", err)
	}

	// Count plants by health status
	plantsByHealth := make(map[entity.HealthStatus]int)
	for _, status := range []entity.HealthStatus{
		entity.HealthStatusThriving,
		entity.HealthStatusHealthy,
		entity.HealthStatusStruggling,
		entity.HealthStatusDiseased,
		entity.HealthStatusDead,
	} {
		plants, err := s.gardenPlantRepo.FindByHealthStatus(ctx, gardenID, status)
		if err != nil {
			return nil, fmt.Errorf("failed to count plants by health status: %w", err)
		}
		plantsByHealth[status] = len(plants)
	}

	return &PlantingStats{
		TotalPlants:    totalCount,
		ActivePlants:   len(activePlants),
		PlantsByHealth: plantsByHealth,
	}, nil
}
