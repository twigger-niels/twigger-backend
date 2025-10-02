package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// CultivarRepository defines the interface for cultivar data access
type CultivarRepository interface {
	// FindByID retrieves a cultivar by its ID
	FindByID(ctx context.Context, cultivarID string) (*entity.Cultivar, error)

	// FindBySpecies retrieves all cultivars for a species
	FindBySpecies(ctx context.Context, speciesID string) ([]*entity.Cultivar, error)

	// FindByPatent retrieves a cultivar by patent number
	FindByPatent(ctx context.Context, patentNumber string) (*entity.Cultivar, error)

	// FindByTradeName retrieves cultivars by trade name
	FindByTradeName(ctx context.Context, tradeName string) ([]*entity.Cultivar, error)

	// FindRestricted retrieves all cultivars with propagation restrictions
	FindRestricted(ctx context.Context) ([]*entity.Cultivar, error)

	// Search searches for cultivars by name
	Search(ctx context.Context, query string, limit int) ([]*entity.Cultivar, error)

	// Create creates a new cultivar
	Create(ctx context.Context, cultivar *entity.Cultivar) error

	// Update updates an existing cultivar
	Update(ctx context.Context, cultivar *entity.Cultivar) error

	// Delete deletes a cultivar by ID
	Delete(ctx context.Context, cultivarID string) error
}
