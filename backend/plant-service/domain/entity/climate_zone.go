package entity

import (
	"fmt"
	"time"

	"twigger-backend/backend/shared/constants"
)

// ClimateZone represents a climatic zone with spatial boundaries
type ClimateZone struct {
	ZoneID           string    `json:"zone_id"`
	CountryID        string    `json:"country_id"`
	ZoneSystem       string    `json:"zone_system"`        // e.g., "USDA", "RHS", "AHS"
	ZoneCode         string    `json:"zone_code"`          // e.g., "7a", "10b"
	ZoneGeometryJSON *string   `json:"zone_geometry,omitempty"` // GeoJSON MultiPolygon
	MinTempC         *float64  `json:"min_temp_c,omitempty"`
	MaxTempC         *float64  `json:"max_temp_c,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Validate validates the climate zone entity
func (cz *ClimateZone) Validate() error {
	if cz.CountryID == "" {
		return fmt.Errorf("country_id is required")
	}

	if cz.ZoneSystem == "" {
		return fmt.Errorf("zone_system is required")
	}

	if cz.ZoneCode == "" {
		return fmt.Errorf("zone_code is required")
	}

	// Validate zone system using shared constants
	if !constants.IsValidClimateSystem(cz.ZoneSystem) {
		return fmt.Errorf("invalid zone_system: %s (must be USDA, RHS, or AHS)", cz.ZoneSystem)
	}

	// Validate temperature range if both provided
	if cz.MinTempC != nil && cz.MaxTempC != nil {
		if *cz.MinTempC > *cz.MaxTempC {
			return fmt.Errorf("min_temp_c cannot be greater than max_temp_c")
		}
	}

	return nil
}
