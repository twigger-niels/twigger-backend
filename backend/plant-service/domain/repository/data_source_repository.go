package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// DataSourceRepository defines the interface for data source access
type DataSourceRepository interface {
	// FindByID retrieves a data source by its ID
	FindByID(ctx context.Context, sourceID string) (*entity.DataSource, error)

	// FindAll retrieves all data sources
	FindAll(ctx context.Context) ([]*entity.DataSource, error)

	// FindByType retrieves all data sources of a specific type
	FindByType(ctx context.Context, sourceType string) ([]*entity.DataSource, error)

	// FindVerified retrieves all verified data sources
	FindVerified(ctx context.Context) ([]*entity.DataSource, error)

	// Create creates a new data source
	Create(ctx context.Context, source *entity.DataSource) error

	// Update updates an existing data source
	Update(ctx context.Context, source *entity.DataSource) error

	// Delete deletes a data source by ID
	Delete(ctx context.Context, sourceID string) error
}
