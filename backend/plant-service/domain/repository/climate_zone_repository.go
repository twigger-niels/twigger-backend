package repository

import (
	"context"

	"twigger-backend/backend/plant-service/domain/entity"
)

// ClimateZoneRepository defines operations for climate zone data access
type ClimateZoneRepository interface {
	// FindByID retrieves a climate zone by its UUID
	FindByID(ctx context.Context, zoneID string) (*entity.ClimateZone, error)

	// FindByCountry retrieves all climate zones for a specific country
	FindByCountry(ctx context.Context, countryID string) ([]*entity.ClimateZone, error)

	// FindByCountryAndSystem retrieves zones for a country and climate system
	FindByCountryAndSystem(ctx context.Context, countryID, zoneSystem string) ([]*entity.ClimateZone, error)

	// FindByCode retrieves a zone by country, system, and code
	FindByCode(ctx context.Context, countryID, zoneSystem, zoneCode string) (*entity.ClimateZone, error)

	// FindByPoint retrieves the climate zone containing a specific geographic point
	FindByPoint(ctx context.Context, latitude, longitude float64, zoneSystem string) (*entity.ClimateZone, error)

	// Create creates a new climate zone
	Create(ctx context.Context, zone *entity.ClimateZone) error

	// Update updates an existing climate zone
	Update(ctx context.Context, zone *entity.ClimateZone) error

	// Delete deletes a climate zone by ID
	Delete(ctx context.Context, zoneID string) error
}
