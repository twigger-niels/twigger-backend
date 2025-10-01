package entity

import (
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/pkg/types"
)

// Companion represents a companion planting relationship between two plants
type Companion struct {
	// Identification
	RelationshipID string `json:"relationship_id"`
	PlantAID       string `json:"plant_a_id"`
	PlantBID       string `json:"plant_b_id"`

	// Relationship details
	RelationshipType types.RelationshipType `json:"relationship_type"`

	// Benefits (for beneficial relationships)
	Benefits []string `json:"benefits,omitempty"` // e.g., ["pest_control", "nitrogen_fixation", "pollinator_attraction"]

	// Spacing recommendations (in meters)
	OptimalDistanceM *float64 `json:"optimal_distance_m,omitempty"`
	MaxDistanceM     *float64 `json:"max_distance_m,omitempty"`

	// Optional: Plant details (populated when needed)
	PlantA *Plant `json:"plant_a,omitempty"`
	PlantB *Plant `json:"plant_b,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks if the Companion relationship is valid
func (c *Companion) Validate() error {
	if c == nil {
		return fmt.Errorf("companion relationship cannot be nil")
	}

	// Check required fields
	if c.RelationshipID == "" {
		return fmt.Errorf("relationship_id is required")
	}
	if c.PlantAID == "" {
		return fmt.Errorf("plant_a_id is required")
	}
	if c.PlantBID == "" {
		return fmt.Errorf("plant_b_id is required")
	}

	// Cannot be companion with self
	if c.PlantAID == c.PlantBID {
		return fmt.Errorf("plant cannot be companion with itself")
	}

	// Validate relationship type
	if !c.RelationshipType.IsValid() {
		return fmt.Errorf("invalid relationship type: %s", c.RelationshipType)
	}

	// Validate distances
	if c.OptimalDistanceM != nil && *c.OptimalDistanceM < 0 {
		return fmt.Errorf("optimal distance cannot be negative: %.2f", *c.OptimalDistanceM)
	}
	if c.MaxDistanceM != nil && *c.MaxDistanceM < 0 {
		return fmt.Errorf("max distance cannot be negative: %.2f", *c.MaxDistanceM)
	}
	if c.OptimalDistanceM != nil && c.MaxDistanceM != nil {
		if *c.OptimalDistanceM > *c.MaxDistanceM {
			return fmt.Errorf("optimal distance (%.2f) cannot exceed max distance (%.2f)",
				*c.OptimalDistanceM, *c.MaxDistanceM)
		}
	}

	return nil
}

// IsBeneficial checks if the relationship is beneficial
func (c *Companion) IsBeneficial() bool {
	return c.RelationshipType == types.RelationshipBeneficial
}

// IsAntagonistic checks if the relationship is antagonistic
func (c *Companion) IsAntagonistic() bool {
	return c.RelationshipType == types.RelationshipAntagonistic
}

// IsNeutral checks if the relationship is neutral
func (c *Companion) IsNeutral() bool {
	return c.RelationshipType == types.RelationshipNeutral
}

// HasBenefit checks if the relationship provides a specific benefit
func (c *Companion) HasBenefit(benefit string) bool {
	for _, b := range c.Benefits {
		if b == benefit {
			return true
		}
	}
	return false
}

// GetOtherPlantID returns the ID of the other plant in the relationship
func (c *Companion) GetOtherPlantID(plantID string) (string, error) {
	if plantID == c.PlantAID {
		return c.PlantBID, nil
	}
	if plantID == c.PlantBID {
		return c.PlantAID, nil
	}
	return "", fmt.Errorf("plant ID %s is not part of this relationship", plantID)
}

// GetOtherPlant returns the other plant entity in the relationship
func (c *Companion) GetOtherPlant(plantID string) (*Plant, error) {
	if plantID == c.PlantAID && c.PlantB != nil {
		return c.PlantB, nil
	}
	if plantID == c.PlantBID && c.PlantA != nil {
		return c.PlantA, nil
	}
	return nil, fmt.Errorf("other plant not loaded or plant ID %s is not part of this relationship", plantID)
}

// IsWithinOptimalDistance checks if a distance is within optimal range
func (c *Companion) IsWithinOptimalDistance(distanceM float64) bool {
	if c.OptimalDistanceM == nil {
		return true // No optimal distance specified
	}
	return distanceM <= *c.OptimalDistanceM
}

// IsWithinMaxDistance checks if a distance is within maximum range
func (c *Companion) IsWithinMaxDistance(distanceM float64) bool {
	if c.MaxDistanceM == nil {
		return true // No max distance specified
	}
	return distanceM <= *c.MaxDistanceM
}

// GetRecommendedDistanceM returns the recommended planting distance
// Returns optimal if set, otherwise max, otherwise nil
func (c *Companion) GetRecommendedDistanceM() *float64 {
	if c.OptimalDistanceM != nil {
		return c.OptimalDistanceM
	}
	return c.MaxDistanceM
}

// CompanionFilter represents filter criteria for companion queries
type CompanionFilter struct {
	PlantID          *string
	RelationshipType *types.RelationshipType
	BeneficialOnly   bool
	ExcludeNeutral   bool
}

// Matches checks if a companion matches the filter criteria
func (cf *CompanionFilter) Matches(c *Companion) bool {
	if cf == nil {
		return true
	}

	// Filter by plant ID
	if cf.PlantID != nil {
		if c.PlantAID != *cf.PlantID && c.PlantBID != *cf.PlantID {
			return false
		}
	}

	// Filter by relationship type
	if cf.RelationshipType != nil && c.RelationshipType != *cf.RelationshipType {
		return false
	}

	// Filter beneficial only
	if cf.BeneficialOnly && !c.IsBeneficial() {
		return false
	}

	// Exclude neutral
	if cf.ExcludeNeutral && c.IsNeutral() {
		return false
	}

	return true
}

// Common benefit types
const (
	BenefitPestControl           = "pest_control"
	BenefitNitrogenFixation      = "nitrogen_fixation"
	BenefitPollinatorAttraction  = "pollinator_attraction"
	BenefitShadeProvision        = "shade_provision"
	BenefitWindProtection        = "wind_protection"
	BenefitGroundCover           = "ground_cover"
	BenefitWeedSuppression       = "weed_suppression"
	BenefitMoistureRetention     = "moisture_retention"
	BenefitSoilImprovement       = "soil_improvement"
	BenefitDiseaseResistance     = "disease_resistance"
)
