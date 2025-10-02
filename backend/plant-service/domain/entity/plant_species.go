package entity

import (
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/pkg/types"
)

// PlantSpecies represents a botanical species
type PlantSpecies struct {
	SpeciesID   string          `json:"species_id"`
	GenusID     string          `json:"genus_id"`
	SpeciesName string          `json:"species_name"` // e.g., "rugosa"
	PlantType   types.PlantType `json:"plant_type"`
	CreatedAt   time.Time       `json:"created_at"`
}

// Validate validates the plant species entity
func (ps *PlantSpecies) Validate() error {
	if ps.GenusID == "" {
		return fmt.Errorf("genus_id is required")
	}

	if ps.SpeciesName == "" {
		return fmt.Errorf("species_name is required")
	}

	if !ps.PlantType.IsValid() {
		return fmt.Errorf("invalid plant_type: %s", ps.PlantType)
	}

	return nil
}
