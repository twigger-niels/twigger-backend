package repository

import (
	"context"
	"twigger-backend/backend/garden-service/domain/entity"
)

// GardenPlantRepository defines the interface for garden plant placement persistence
type GardenPlantRepository interface {
	// CRUD operations
	Create(ctx context.Context, gardenPlant *entity.GardenPlant) error
	FindByID(ctx context.Context, gardenPlantID string) (*entity.GardenPlant, error)
	FindByGardenID(ctx context.Context, gardenID string, includeRemoved bool) ([]*entity.GardenPlant, error)
	FindByZoneID(ctx context.Context, zoneID string, includeRemoved bool) ([]*entity.GardenPlant, error)
	FindByPlantID(ctx context.Context, plantID string) ([]*entity.GardenPlant, error)
	Update(ctx context.Context, gardenPlant *entity.GardenPlant) error
	Delete(ctx context.Context, gardenPlantID string) error

	// Batch operations
	FindByIDs(ctx context.Context, gardenPlantIDs []string) ([]*entity.GardenPlant, error)
	BulkCreate(ctx context.Context, gardenPlants []*entity.GardenPlant) error

	// Spatial queries
	CheckPlantSpacing(ctx context.Context, gardenID, locationGeoJSON string, minDistanceM float64) ([]*entity.GardenPlant, error)
	FindInZone(ctx context.Context, zoneID string) ([]*entity.GardenPlant, error)
	ValidatePlantLocation(ctx context.Context, gardenID, locationGeoJSON string) error

	// Health and status queries
	FindByHealthStatus(ctx context.Context, gardenID string, status entity.HealthStatus) ([]*entity.GardenPlant, error)
	FindActiveInGarden(ctx context.Context, gardenID string) ([]*entity.GardenPlant, error)

	// Statistics
	CountByGardenID(ctx context.Context, gardenID string, includeRemoved bool) (int, error)
	CountByPlantID(ctx context.Context, plantID string) (int, error)
}
