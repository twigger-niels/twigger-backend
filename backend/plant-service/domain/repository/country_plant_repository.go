package repository

import (
	"context"
	"twigger-backend/backend/plant-service/domain/entity"
)

// CountryPlantRepository defines the interface for country-plant relationship data access
type CountryPlantRepository interface {
	// FindByID retrieves a country-plant relationship by its ID
	FindByID(ctx context.Context, countryPlantID string) (*entity.CountryPlant, error)

	// FindByCountry retrieves all plant relationships for a country
	FindByCountry(ctx context.Context, countryID string) ([]*entity.CountryPlant, error)

	// FindByPlant retrieves all country relationships for a plant
	FindByPlant(ctx context.Context, plantID string) ([]*entity.CountryPlant, error)

	// FindByCountryAndPlant retrieves a specific country-plant relationship
	FindByCountryAndPlant(ctx context.Context, countryID, plantID string) (*entity.CountryPlant, error)

	// FindByNativeStatus retrieves all plants with a specific native status in a country
	FindByNativeStatus(ctx context.Context, countryID, nativeStatus string) ([]*entity.CountryPlant, error)

	// FindByLegalStatus retrieves all plants with a specific legal status in a country
	FindByLegalStatus(ctx context.Context, countryID, legalStatus string) ([]*entity.CountryPlant, error)

	// Create creates a new country-plant relationship
	Create(ctx context.Context, countryPlant *entity.CountryPlant) error

	// Update updates an existing country-plant relationship
	Update(ctx context.Context, countryPlant *entity.CountryPlant) error

	// Delete deletes a country-plant relationship by ID
	Delete(ctx context.Context, countryPlantID string) error
}
