package entity

import (
	"fmt"
	"time"
)

// PlantSynonym represents an old/deprecated botanical name
type PlantSynonym struct {
	SynonymID      string     `json:"synonym_id"`
	CurrentPlantID string     `json:"current_plant_id"`
	OldName        string     `json:"old_name"`
	DateDeprecated *time.Time `json:"date_deprecated,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Validate validates the plant synonym entity
func (ps *PlantSynonym) Validate() error {
	if ps.CurrentPlantID == "" {
		return fmt.Errorf("current_plant_id is required")
	}

	if ps.OldName == "" {
		return fmt.Errorf("old_name is required")
	}

	return nil
}
