package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// PlantProblemRepository defines the interface for plant problem data access
type PlantProblemRepository interface {
	// FindByID retrieves a plant problem by its ID
	FindByID(ctx context.Context, problemID string) (*entity.PlantProblem, error)

	// FindByPlant retrieves problems for a plant with pagination
	FindByPlant(ctx context.Context, plantID string, limit, offset int) ([]*entity.PlantProblem, error)

	// FindByType retrieves problems of a specific type for a plant with pagination
	FindByType(ctx context.Context, plantID, problemType string, limit, offset int) ([]*entity.PlantProblem, error)

	// FindBySeverity retrieves problems of a specific severity for a plant with pagination
	FindBySeverity(ctx context.Context, plantID, severity string, limit, offset int) ([]*entity.PlantProblem, error)

	// Create creates a new plant problem
	Create(ctx context.Context, problem *entity.PlantProblem) error

	// Update updates an existing plant problem
	Update(ctx context.Context, problem *entity.PlantProblem) error

	// Delete deletes a plant problem by ID
	Delete(ctx context.Context, problemID string) error
}
