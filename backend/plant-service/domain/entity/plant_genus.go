package entity

import (
	"fmt"
	"time"
)

// PlantGenus represents a botanical genus
type PlantGenus struct {
	GenusID   string    `json:"genus_id"`
	FamilyID  string    `json:"family_id"`
	GenusName string    `json:"genus_name"` // e.g., "Rosa"
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates the plant genus entity
func (pg *PlantGenus) Validate() error {
	if pg.FamilyID == "" {
		return fmt.Errorf("family_id is required")
	}

	if pg.GenusName == "" {
		return fmt.Errorf("genus_name is required")
	}

	return nil
}
