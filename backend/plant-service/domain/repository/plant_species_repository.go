package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// PlantSpeciesRepository defines the interface for plant species data access
type PlantSpeciesRepository interface {
	// FindByID retrieves a plant species by its ID
	FindByID(ctx context.Context, speciesID string) (*entity.PlantSpecies, error)

	// FindByGenus retrieves all species in a genus
	FindByGenus(ctx context.Context, genusID string) ([]*entity.PlantSpecies, error)

	// FindByType retrieves all species of a specific plant type
	FindByType(ctx context.Context, plantType string) ([]*entity.PlantSpecies, error)

	// FindAll retrieves all plant species
	FindAll(ctx context.Context) ([]*entity.PlantSpecies, error)

	// Search searches for plant species by name
	Search(ctx context.Context, query string, limit int) ([]*entity.PlantSpecies, error)

	// Create creates a new plant species
	Create(ctx context.Context, species *entity.PlantSpecies) error

	// Update updates an existing plant species
	Update(ctx context.Context, species *entity.PlantSpecies) error

	// Delete deletes a plant species by ID
	Delete(ctx context.Context, speciesID string) error
}
