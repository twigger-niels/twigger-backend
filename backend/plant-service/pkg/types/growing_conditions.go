package types

import (
	"fmt"
	"time"
)

// GrowingConditions represents the environmental requirements for a plant
// This is a value object that aggregates data from multiple sources
type GrowingConditions struct {
	// Identification
	PlantID       string    `json:"plant_id"`
	CountryID     *string   `json:"country_id,omitempty"`
	SourceID      *string   `json:"source_id,omitempty"`
	Confidence    ConfidenceLevel `json:"confidence"`

	// Climate zones
	HardinessZones []string `json:"hardiness_zones,omitempty"` // e.g., ["5a", "5b", "6a"]
	HeatZones      []string `json:"heat_zones,omitempty"`      // e.g., ["7", "8", "9"]

	// Sun and shade requirements
	SunRequirements []SunRequirement `json:"sun_requirements,omitempty"`
	ShadeTolerance  bool             `json:"shade_tolerance"`

	// Water and humidity
	WaterNeeds         *WaterNeeds `json:"water_needs,omitempty"`
	HumidityPreference *float64    `json:"humidity_preference,omitempty"` // Percentage 0-100
	DroughtTolerant    bool        `json:"drought_tolerant"`

	// Soil requirements
	SoilTypes    []string      `json:"soil_types,omitempty"` // e.g., ["clay", "loam", "sand"]
	SoilDrainage *SoilDrainage `json:"soil_drainage,omitempty"`
	PHPreference *PHRange      `json:"ph_preference,omitempty"`

	// Tolerances
	SaltTolerant bool `json:"salt_tolerant"`
	WindTolerant bool `json:"wind_tolerant"`

	// Temporal aspects (months 1-12)
	FloweringMonths []int `json:"flowering_months,omitempty"` // e.g., [5, 6, 7] for May-July
	FruitingMonths  []int `json:"fruiting_months,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks if the GrowingConditions are valid
func (gc *GrowingConditions) Validate() error {
	if gc == nil {
		return fmt.Errorf("growing conditions cannot be nil")
	}

	// Check required fields
	if gc.PlantID == "" {
		return fmt.Errorf("plant_id is required")
	}

	// Validate confidence level
	if !gc.Confidence.IsValid() {
		return fmt.Errorf("invalid confidence level: %s", gc.Confidence)
	}

	// Validate sun requirements
	for _, sr := range gc.SunRequirements {
		if !sr.IsValid() {
			return fmt.Errorf("invalid sun requirement: %s", sr)
		}
	}

	// Validate water needs
	if gc.WaterNeeds != nil && !gc.WaterNeeds.IsValid() {
		return fmt.Errorf("invalid water needs: %s", *gc.WaterNeeds)
	}

	// Validate soil drainage
	if gc.SoilDrainage != nil && !gc.SoilDrainage.IsValid() {
		return fmt.Errorf("invalid soil drainage: %s", *gc.SoilDrainage)
	}

	// Validate humidity percentage
	if gc.HumidityPreference != nil {
		if *gc.HumidityPreference < 0 || *gc.HumidityPreference > 100 {
			return fmt.Errorf("humidity preference must be between 0 and 100: %.2f", *gc.HumidityPreference)
		}
	}

	// Validate pH preference
	if gc.PHPreference != nil {
		if err := gc.PHPreference.IsValid(); err != nil {
			return fmt.Errorf("invalid pH preference: %w", err)
		}
	}

	// Validate months (1-12)
	for _, month := range gc.FloweringMonths {
		if month < 1 || month > 12 {
			return fmt.Errorf("invalid flowering month: %d", month)
		}
	}
	for _, month := range gc.FruitingMonths {
		if month < 1 || month > 12 {
			return fmt.Errorf("invalid fruiting month: %d", month)
		}
	}

	return nil
}

// HasSunRequirement checks if the plant can tolerate a specific sun requirement
func (gc *GrowingConditions) HasSunRequirement(req SunRequirement) bool {
	for _, sr := range gc.SunRequirements {
		if sr == req {
			return true
		}
	}
	return false
}

// IsFloweringInMonth checks if the plant flowers in a specific month (1-12)
func (gc *GrowingConditions) IsFloweringInMonth(month int) bool {
	for _, m := range gc.FloweringMonths {
		if m == month {
			return true
		}
	}
	return false
}

// IsFruitingInMonth checks if the plant fruits in a specific month (1-12)
func (gc *GrowingConditions) IsFruitingInMonth(month int) bool {
	for _, m := range gc.FruitingMonths {
		if m == month {
			return true
		}
	}
	return false
}

// SupportsHardinessZone checks if the plant can grow in a specific hardiness zone
func (gc *GrowingConditions) SupportsHardinessZone(zone string) bool {
	for _, z := range gc.HardinessZones {
		if z == zone {
			return true
		}
	}
	return false
}

// SupportsHeatZone checks if the plant can tolerate a specific heat zone
func (gc *GrowingConditions) SupportsHeatZone(zone string) bool {
	for _, z := range gc.HeatZones {
		if z == zone {
			return true
		}
	}
	return false
}

// HasSoilType checks if the plant can grow in a specific soil type
func (gc *GrowingConditions) HasSoilType(soilType string) bool {
	for _, st := range gc.SoilTypes {
		if st == soilType {
			return true
		}
	}
	return false
}

// ConfidenceScore returns a numeric confidence score (0-100)
func (gc *GrowingConditions) ConfidenceScore() int {
	switch gc.Confidence {
	case ConfidenceVeryLow:
		return 10
	case ConfidenceLow:
		return 30
	case ConfidenceModerate:
		return 50
	case ConfidenceProbable:
		return 70
	case ConfidenceVeryHigh:
		return 90
	case ConfidenceConfirmed:
		return 100
	default:
		return 0
	}
}
