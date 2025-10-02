package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// PlantSynonymRepository defines the interface for plant synonym data access
type PlantSynonymRepository interface {
	// FindByID retrieves a plant synonym by its ID
	FindByID(ctx context.Context, synonymID string) (*entity.PlantSynonym, error)

	// FindByCurrentPlant retrieves all synonyms for a current plant
	FindByCurrentPlant(ctx context.Context, currentPlantID string) ([]*entity.PlantSynonym, error)

	// FindByOldName retrieves synonyms by old name
	FindByOldName(ctx context.Context, oldName string) ([]*entity.PlantSynonym, error)

	// Create creates a new plant synonym
	Create(ctx context.Context, synonym *entity.PlantSynonym) error

	// Update updates an existing plant synonym
	Update(ctx context.Context, synonym *entity.PlantSynonym) error

	// Delete deletes a plant synonym by ID
	Delete(ctx context.Context, synonymID string) error
}
