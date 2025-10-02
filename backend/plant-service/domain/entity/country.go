package entity

import (
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Country represents a geographical country with climate data
type Country struct {
	CountryID              string         `json:"country_id"`
	CountryCode            string         `json:"country_code"` // ISO 3166-1 alpha-2 (e.g., "US", "GB")
	CountryName            string         `json:"country_name"`
	ClimateSystems         []string       `json:"climate_systems"` // e.g., ["USDA", "RHS"]
	DefaultClimateSystem   *string        `json:"default_climate_system,omitempty"`
	CountryBoundaryGeoJSON *string        `json:"country_boundary,omitempty"` // GeoJSON MultiPolygon
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

// Validate validates the country entity
func (c *Country) Validate() error {
	if c.CountryCode == "" {
		return fmt.Errorf("country_code is required")
	}

	if len(c.CountryCode) != 2 {
		return fmt.Errorf("country_code must be 2 characters (ISO 3166-1 alpha-2)")
	}

	if c.CountryName == "" {
		return fmt.Errorf("country_name is required")
	}

	if len(c.ClimateSystems) == 0 {
		return fmt.Errorf("at least one climate system is required")
	}

	// Validate climate systems
	validSystems := map[string]bool{
		"USDA": true,
		"RHS":  true,
		"AHS":  true, // American Horticultural Society
	}

	for _, system := range c.ClimateSystems {
		if !validSystems[system] {
			return fmt.Errorf("invalid climate system: %s (must be USDA, RHS, or AHS)", system)
		}
	}

	// Validate default climate system if set
	if c.DefaultClimateSystem != nil {
		found := false
		for _, system := range c.ClimateSystems {
			if system == *c.DefaultClimateSystem {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default_climate_system must be one of the climate_systems")
		}
	}

	return nil
}

// ClimateSystemsArray converts climate systems to pq.StringArray for database storage
func (c *Country) ClimateSystemsArray() pq.StringArray {
	return pq.StringArray(c.ClimateSystems)
}
