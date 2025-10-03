package entity

import (
	"fmt"
	"time"
)

// GardenType represents the type of garden
type GardenType string

const (
	GardenTypeOrnamental  GardenType = "ornamental"
	GardenTypeVegetable   GardenType = "vegetable"
	GardenTypeMixed       GardenType = "mixed"
	GardenTypeOrchard     GardenType = "orchard"
	GardenTypeGreenhouse  GardenType = "greenhouse"
)

// Aspect represents compass direction
type Aspect string

const (
	AspectN    Aspect = "N"
	AspectNE   Aspect = "NE"
	AspectE    Aspect = "E"
	AspectSE   Aspect = "SE"
	AspectS    Aspect = "S"
	AspectSW   Aspect = "SW"
	AspectW    Aspect = "W"
	AspectNW   Aspect = "NW"
	AspectFlat Aspect = "flat"
)

// Garden represents a user's garden with spatial boundary
type Garden struct {
	GardenID       string      `json:"garden_id"`
	UserID         string      `json:"user_id"`
	GardenName     string      `json:"garden_name"`

	// Spatial data - stored as GeoJSON strings
	BoundaryGeoJSON *string    `json:"boundary,omitempty"`      // GEOMETRY(Polygon, 4326) as GeoJSON
	LocationGeoJSON *string    `json:"location,omitempty"`      // GEOGRAPHY(Point, 4326) as GeoJSON

	// Environmental data
	ElevationM      *float64   `json:"elevation_m,omitempty"`
	SlopeDegrees    *float64   `json:"slope_degrees,omitempty"`  // 0-90
	Aspect          *Aspect    `json:"aspect,omitempty"`

	// Detected from spatial queries
	HardinessZone   *string    `json:"hardiness_zone,omitempty"` // Detected via ST_Contains with climate_zones

	// Garden metadata
	GardenType      *GardenType `json:"garden_type,omitempty"`

	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Validate validates the garden entity
func (g *Garden) Validate() error {
	if g.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if g.GardenName == "" {
		return fmt.Errorf("garden_name is required")
	}

	if len(g.GardenName) > 200 {
		return fmt.Errorf("garden_name must be 200 characters or less")
	}

	// Validate slope if provided
	if g.SlopeDegrees != nil {
		if *g.SlopeDegrees < 0 || *g.SlopeDegrees > 90 {
			return fmt.Errorf("slope_degrees must be between 0 and 90")
		}
	}

	// Validate aspect if provided
	if g.Aspect != nil {
		if !isValidAspect(*g.Aspect) {
			return fmt.Errorf("invalid aspect: must be N, NE, E, SE, S, SW, W, NW, or flat")
		}
	}

	// Validate garden type if provided
	if g.GardenType != nil {
		if !isValidGardenType(*g.GardenType) {
			return fmt.Errorf("invalid garden_type: must be ornamental, vegetable, mixed, orchard, or greenhouse")
		}
	}

	return nil
}

// isValidAspect checks if aspect value is valid
func isValidAspect(aspect Aspect) bool {
	validAspects := []Aspect{
		AspectN, AspectNE, AspectE, AspectSE,
		AspectS, AspectSW, AspectW, AspectNW, AspectFlat,
	}

	for _, valid := range validAspects {
		if aspect == valid {
			return true
		}
	}
	return false
}

// isValidGardenType checks if garden type value is valid
func isValidGardenType(gardenType GardenType) bool {
	validTypes := []GardenType{
		GardenTypeOrnamental, GardenTypeVegetable, GardenTypeMixed,
		GardenTypeOrchard, GardenTypeGreenhouse,
	}

	for _, valid := range validTypes {
		if gardenType == valid {
			return true
		}
	}
	return false
}
