package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/backend/garden-service/domain/repository"
)

// ZoneManagementService defines the business logic for managing garden zones
type ZoneManagementService interface {
	// CreateZone creates a new zone with validation that it's within garden bounds
	CreateZone(ctx context.Context, zone *entity.GardenZone) (*entity.GardenZone, error)

	// GetZone retrieves a zone by ID
	GetZone(ctx context.Context, zoneID string) (*entity.GardenZone, error)

	// ListGardenZones retrieves all zones for a garden
	ListGardenZones(ctx context.Context, gardenID string) ([]*entity.GardenZone, error)

	// UpdateZone updates an existing zone with validation
	UpdateZone(ctx context.Context, zone *entity.GardenZone) (*entity.GardenZone, error)

	// DeleteZone deletes a zone
	DeleteZone(ctx context.Context, zoneID string) error

	// CalculateZoneArea calculates the area of a zone in square meters
	CalculateZoneArea(ctx context.Context, zoneID string) (float64, error)

	// GetTotalZoneArea calculates total area of all zones in a garden
	GetTotalZoneArea(ctx context.Context, gardenID string) (float64, error)

	// CheckZoneOverlaps checks if a zone overlaps with existing zones
	CheckZoneOverlaps(ctx context.Context, gardenID, zoneGeometryGeoJSON string, excludeZoneID *string) (bool, error)
}

// zoneManagementService implements ZoneManagementService
type zoneManagementService struct {
	zoneRepo   repository.GardenZoneRepository
	gardenRepo repository.GardenRepository
}

// NewZoneManagementService creates a new zone management service instance
func NewZoneManagementService(
	zoneRepo repository.GardenZoneRepository,
	gardenRepo repository.GardenRepository,
) ZoneManagementService {
	return &zoneManagementService{
		zoneRepo:   zoneRepo,
		gardenRepo: gardenRepo,
	}
}

// CreateZone creates a new zone with validation that it's within garden bounds
func (s *zoneManagementService) CreateZone(ctx context.Context, zone *entity.GardenZone) (*entity.GardenZone, error) {
	// Validate entity
	if err := zone.Validate(); err != nil {
		return nil, entity.NewValidationError("garden_zone", err.Error())
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, zone.GardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	// Validate zone is within garden boundary
	if err := s.zoneRepo.ValidateZoneWithinGarden(ctx, zone.GardenID, zone.GeometryGeoJSON); err != nil {
		return nil, fmt.Errorf("zone must be within garden boundary: %w", err)
	}

	// Check for overlaps with existing zones
	overlaps, err := s.zoneRepo.CheckZoneOverlaps(ctx, zone.GardenID, zone.GeometryGeoJSON, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check zone overlaps: %w", err)
	}

	if overlaps {
		return nil, entity.NewValidationError("garden_zone", "zone overlaps with existing zone")
	}

	// Generate ID if not provided
	if zone.ZoneID == "" {
		zone.ZoneID = uuid.New().String()
	}

	// Set timestamp
	zone.CreatedAt = time.Now()

	// Create zone
	if err := s.zoneRepo.Create(ctx, zone); err != nil {
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}

	return zone, nil
}

// GetZone retrieves a zone by ID
func (s *zoneManagementService) GetZone(ctx context.Context, zoneID string) (*entity.GardenZone, error) {
	if zoneID == "" {
		return nil, entity.NewInvalidInputError("zone_id", "zone ID cannot be empty")
	}

	zone, err := s.zoneRepo.FindByID(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}

	return zone, nil
}

// ListGardenZones retrieves all zones for a garden
func (s *zoneManagementService) ListGardenZones(ctx context.Context, gardenID string) ([]*entity.GardenZone, error) {
	if gardenID == "" {
		return nil, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("garden not found: %w", err)
	}

	zones, err := s.zoneRepo.FindByGardenID(ctx, gardenID)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	return zones, nil
}

// UpdateZone updates an existing zone with validation
func (s *zoneManagementService) UpdateZone(ctx context.Context, zone *entity.GardenZone) (*entity.GardenZone, error) {
	// Validate entity
	if err := zone.Validate(); err != nil {
		return nil, entity.NewValidationError("garden_zone", err.Error())
	}

	// Check if zone exists
	existing, err := s.zoneRepo.FindByID(ctx, zone.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	// Validate zone is within garden boundary if geometry changed
	if zone.GeometryGeoJSON != existing.GeometryGeoJSON {
		if err := s.zoneRepo.ValidateZoneWithinGarden(ctx, zone.GardenID, zone.GeometryGeoJSON); err != nil {
			return nil, fmt.Errorf("zone must be within garden boundary: %w", err)
		}

		// Check for overlaps with other zones (exclude this zone)
		overlaps, err := s.zoneRepo.CheckZoneOverlaps(ctx, zone.GardenID, zone.GeometryGeoJSON, &zone.ZoneID)
		if err != nil {
			return nil, fmt.Errorf("failed to check zone overlaps: %w", err)
		}

		if overlaps {
			return nil, entity.NewValidationError("garden_zone", "zone overlaps with existing zone")
		}
	}

	// Preserve created_at
	zone.CreatedAt = existing.CreatedAt

	if err := s.zoneRepo.Update(ctx, zone); err != nil {
		return nil, fmt.Errorf("failed to update zone: %w", err)
	}

	return zone, nil
}

// DeleteZone deletes a zone
func (s *zoneManagementService) DeleteZone(ctx context.Context, zoneID string) error {
	if zoneID == "" {
		return entity.NewInvalidInputError("zone_id", "zone ID cannot be empty")
	}

	// Check if zone exists
	_, err := s.zoneRepo.FindByID(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("zone not found: %w", err)
	}

	// Delete zone
	if err := s.zoneRepo.Delete(ctx, zoneID); err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	return nil
}

// CalculateZoneArea calculates the area of a zone in square meters
func (s *zoneManagementService) CalculateZoneArea(ctx context.Context, zoneID string) (float64, error) {
	if zoneID == "" {
		return 0, entity.NewInvalidInputError("zone_id", "zone ID cannot be empty")
	}

	area, err := s.zoneRepo.CalculateArea(ctx, zoneID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate area: %w", err)
	}

	return area, nil
}

// GetTotalZoneArea calculates total area of all zones in a garden
func (s *zoneManagementService) GetTotalZoneArea(ctx context.Context, gardenID string) (float64, error) {
	if gardenID == "" {
		return 0, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return 0, fmt.Errorf("garden not found: %w", err)
	}

	totalArea, err := s.zoneRepo.CalculateTotalArea(ctx, gardenID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total area: %w", err)
	}

	return totalArea, nil
}

// CheckZoneOverlaps checks if a zone overlaps with existing zones
func (s *zoneManagementService) CheckZoneOverlaps(ctx context.Context, gardenID, zoneGeometryGeoJSON string, excludeZoneID *string) (bool, error) {
	if gardenID == "" {
		return false, entity.NewInvalidInputError("garden_id", "garden ID cannot be empty")
	}

	if zoneGeometryGeoJSON == "" {
		return false, entity.NewInvalidInputError("zone_geometry", "zone geometry cannot be empty")
	}

	// Check if garden exists
	_, err := s.gardenRepo.FindByID(ctx, gardenID)
	if err != nil {
		return false, fmt.Errorf("garden not found: %w", err)
	}

	overlaps, err := s.zoneRepo.CheckZoneOverlaps(ctx, gardenID, zoneGeometryGeoJSON, excludeZoneID)
	if err != nil {
		return false, fmt.Errorf("failed to check zone overlaps: %w", err)
	}

	return overlaps, nil
}
