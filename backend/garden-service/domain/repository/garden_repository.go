package repository

import (
	"context"
	"twigger-backend/backend/garden-service/domain/entity"
)

// GardenRepository defines the interface for garden persistence
type GardenRepository interface {
	// CRUD operations
	Create(ctx context.Context, garden *entity.Garden) error
	FindByID(ctx context.Context, gardenID string) (*entity.Garden, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Garden, error)
	Update(ctx context.Context, garden *entity.Garden) error
	Delete(ctx context.Context, gardenID string) error

	// Batch operations
	FindByIDs(ctx context.Context, gardenIDs []string) ([]*entity.Garden, error)

	// Spatial queries
	FindByLocation(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Garden, error)
	CalculateArea(ctx context.Context, gardenID string) (float64, error)
	GetCenterPoint(ctx context.Context, gardenID string) (lat, lng float64, err error)

	// Climate zone detection
	DetectHardinessZone(ctx context.Context, gardenID string) (string, error)

	// Validation
	ValidateBoundary(ctx context.Context, boundaryGeoJSON string) error
	CheckBoundaryValid(ctx context.Context, gardenID string) (bool, error)

	// Statistics
	CountByUserID(ctx context.Context, userID string) (int, error)
	GetTotalArea(ctx context.Context, userID string) (float64, error)
}
