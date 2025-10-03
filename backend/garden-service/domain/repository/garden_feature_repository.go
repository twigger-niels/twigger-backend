package repository

import (
	"context"
	"twigger-backend/backend/garden-service/domain/entity"
)

// GardenFeatureRepository defines the interface for garden feature persistence
type GardenFeatureRepository interface {
	// CRUD operations
	Create(ctx context.Context, feature *entity.GardenFeature) error
	FindByID(ctx context.Context, featureID string) (*entity.GardenFeature, error)
	FindByGardenID(ctx context.Context, gardenID string) ([]*entity.GardenFeature, error)
	FindByType(ctx context.Context, gardenID string, featureType entity.FeatureType) ([]*entity.GardenFeature, error)
	Update(ctx context.Context, feature *entity.GardenFeature) error
	Delete(ctx context.Context, featureID string) error

	// Batch operations
	FindByIDs(ctx context.Context, featureIDs []string) ([]*entity.GardenFeature, error)

	// For shade calculations (Part 4 dependency)
	FindFeaturesWithHeight(ctx context.Context, gardenID string) ([]*entity.GardenFeature, error)
	FindTreesInGarden(ctx context.Context, gardenID string) ([]*entity.GardenFeature, error)

	// Statistics
	CountByGardenID(ctx context.Context, gardenID string) (int, error)
}
