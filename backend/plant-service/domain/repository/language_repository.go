package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// LanguageRepository defines the interface for language data access
type LanguageRepository interface {
	// FindByID retrieves a language by its ID
	FindByID(ctx context.Context, languageID string) (*entity.Language, error)

	// FindByCode retrieves a language by its ISO code (e.g., "en", "es")
	FindByCode(ctx context.Context, languageCode string) (*entity.Language, error)

	// FindAll retrieves all languages
	FindAll(ctx context.Context) ([]*entity.Language, error)

	// FindActive retrieves all active languages
	FindActive(ctx context.Context) ([]*entity.Language, error)

	// Create creates a new language
	Create(ctx context.Context, language *entity.Language) error

	// Update updates an existing language
	Update(ctx context.Context, language *entity.Language) error

	// Delete deletes a language by ID
	Delete(ctx context.Context, languageID string) error
}
