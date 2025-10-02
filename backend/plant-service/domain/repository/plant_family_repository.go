package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// PlantFamilyRepository defines the interface for plant family data access
type PlantFamilyRepository interface {
	// FindByID retrieves a plant family by its ID
	FindByID(ctx context.Context, familyID string) (*entity.PlantFamily, error)

	// FindByName retrieves a plant family by its name
	FindByName(ctx context.Context, familyName string) (*entity.PlantFamily, error)

	// FindAll retrieves all plant families
	FindAll(ctx context.Context) ([]*entity.PlantFamily, error)

	// Search searches for plant families by name
	Search(ctx context.Context, query string, limit int) ([]*entity.PlantFamily, error)

	// Create creates a new plant family
	Create(ctx context.Context, family *entity.PlantFamily) error

	// Update updates an existing plant family
	Update(ctx context.Context, family *entity.PlantFamily) error

	// Delete deletes a plant family by ID
	Delete(ctx context.Context, familyID string) error
}
