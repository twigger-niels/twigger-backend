package entity

import (
	"fmt"
	"time"
)

// Cultivar represents a cultivated variety of a plant
type Cultivar struct {
	CultivarID            string     `json:"cultivar_id"`
	SpeciesID             string     `json:"species_id"`
	CultivarName          string     `json:"cultivar_name"`
	TradeName             *string    `json:"trade_name,omitempty"`
	PatentNumber          *string    `json:"patent_number,omitempty"`
	PatentExpiry          *time.Time `json:"patent_expiry,omitempty"`
	PropagationRestricted bool       `json:"propagation_restricted"`
	CreatedAt             time.Time  `json:"created_at"`
}

// Validate validates the cultivar entity
func (c *Cultivar) Validate() error {
	if c.SpeciesID == "" {
		return fmt.Errorf("species_id is required")
	}

	if c.CultivarName == "" {
		return fmt.Errorf("cultivar_name is required")
	}

	return nil
}
