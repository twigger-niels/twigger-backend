package entity

import (
	"fmt"
	"time"
)

// PlantFamily represents a botanical family
type PlantFamily struct {
	FamilyID   string    `json:"family_id"`
	FamilyName string    `json:"family_name"` // e.g., "Rosaceae"
	CommonName *string   `json:"common_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Validate validates the plant family entity
func (pf *PlantFamily) Validate() error {
	if pf.FamilyName == "" {
		return fmt.Errorf("family_name is required")
	}

	return nil
}
