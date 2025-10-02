package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// PlantProblemRepository defines the interface for plant problem data access
type PlantProblemRepository interface {
	// FindByID retrieves a plant problem by its ID
	FindByID(ctx context.Context, problemID string) (*entity.PlantProblem, error)

	// FindByPlant retrieves all problems for a plant
	FindByPlant(ctx context.Context, plantID string) ([]*entity.PlantProblem, error)

	// FindByType retrieves all problems of a specific type for a plant
	FindByType(ctx context.Context, plantID, problemType string) ([]*entity.PlantProblem, error)

	// FindBySeverity retrieves all problems of a specific severity for a plant
	FindBySeverity(ctx context.Context, plantID, severity string) ([]*entity.PlantProblem, error)

	// Create creates a new plant problem
	Create(ctx context.Context, problem *entity.PlantProblem) error

	// Update updates an existing plant problem
	Update(ctx context.Context, problem *entity.PlantProblem) error

	// Delete deletes a plant problem by ID
	Delete(ctx context.Context, problemID string) error
}
