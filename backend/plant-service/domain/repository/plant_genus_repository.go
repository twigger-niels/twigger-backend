package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// PlantGenusRepository defines the interface for plant genus data access
type PlantGenusRepository interface {
	// FindByID retrieves a plant genus by its ID
	FindByID(ctx context.Context, genusID string) (*entity.PlantGenus, error)

	// FindByName retrieves a plant genus by its name
	FindByName(ctx context.Context, genusName string) (*entity.PlantGenus, error)

	// FindByFamily retrieves all genera in a family
	FindByFamily(ctx context.Context, familyID string) ([]*entity.PlantGenus, error)

	// FindAll retrieves all plant genera
	FindAll(ctx context.Context) ([]*entity.PlantGenus, error)

	// Search searches for plant genera by name
	Search(ctx context.Context, query string, limit int) ([]*entity.PlantGenus, error)

	// Create creates a new plant genus
	Create(ctx context.Context, genus *entity.PlantGenus) error

	// Update updates an existing plant genus
	Update(ctx context.Context, genus *entity.PlantGenus) error

	// Delete deletes a plant genus by ID
	Delete(ctx context.Context, genusID string) error
}
