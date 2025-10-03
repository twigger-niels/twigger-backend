package repository

import (
	"context"
	"twigger-backend/backend/garden-service/domain/entity"
)

// GardenZoneRepository defines the interface for garden zone persistence
type GardenZoneRepository interface {
	// CRUD operations
	Create(ctx context.Context, zone *entity.GardenZone) error
	FindByID(ctx context.Context, zoneID string) (*entity.GardenZone, error)
	FindByGardenID(ctx context.Context, gardenID string) ([]*entity.GardenZone, error)
	Update(ctx context.Context, zone *entity.GardenZone) error
	Delete(ctx context.Context, zoneID string) error

	// Batch operations
	FindByIDs(ctx context.Context, zoneIDs []string) ([]*entity.GardenZone, error)

	// Spatial validation
	ValidateZoneWithinGarden(ctx context.Context, gardenID, zoneGeometryGeoJSON string) error
	CheckZoneOverlaps(ctx context.Context, gardenID, zoneGeometryGeoJSON string, excludeZoneID *string) (bool, error)

	// Area calculations
	CalculateTotalArea(ctx context.Context, gardenID string) (float64, error)
	CalculateZoneArea(ctx context.Context, zoneID string) (float64, error)

	// Statistics
	CountByGardenID(ctx context.Context, gardenID string) (int, error)
}
