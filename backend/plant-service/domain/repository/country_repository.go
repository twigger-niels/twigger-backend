package repository

import (
	"context"

	"twigger-backend/backend/plant-service/domain/entity"
)

// CountryRepository defines operations for country data access
type CountryRepository interface {
	// FindByID retrieves a country by its UUID
	FindByID(ctx context.Context, countryID string) (*entity.Country, error)

	// FindByCode retrieves a country by its ISO 3166-1 alpha-2 code (e.g., "US", "GB")
	FindByCode(ctx context.Context, countryCode string) (*entity.Country, error)

	// FindAll retrieves all countries
	FindAll(ctx context.Context) ([]*entity.Country, error)

	// FindByClimateSystem retrieves all countries that support a specific climate system
	FindByClimateSystem(ctx context.Context, climateSystem string) ([]*entity.Country, error)

	// FindByPoint retrieves the country containing a specific geographic point (lat, lng)
	FindByPoint(ctx context.Context, latitude, longitude float64) (*entity.Country, error)

	// Create creates a new country
	Create(ctx context.Context, country *entity.Country) error

	// Update updates an existing country
	Update(ctx context.Context, country *entity.Country) error

	// Delete deletes a country by ID
	Delete(ctx context.Context, countryID string) error
}
