package entity

import (
	"fmt"
	"time"
)

// FeatureType represents the type of garden feature
type FeatureType string

const (
	FeatureTypeTree       FeatureType = "tree"
	FeatureTypeShrub      FeatureType = "shrub"
	FeatureTypeBuilding   FeatureType = "building"
	FeatureTypeFence      FeatureType = "fence"
	FeatureTypeWall       FeatureType = "wall"
	FeatureTypeGreenhouse FeatureType = "greenhouse"
	FeatureTypeShed       FeatureType = "shed"
	FeatureTypePond       FeatureType = "pond"
	FeatureTypePath       FeatureType = "path"
)

// GardenFeature represents a feature within a garden (tree, building, structure)
// Used for shade calculations and spatial analysis
type GardenFeature struct {
	FeatureID   string       `json:"feature_id"`
	GardenID    string       `json:"garden_id"`
	FeatureType FeatureType  `json:"feature_type"`
	FeatureName *string      `json:"feature_name,omitempty"`

	// Spatial representation - stored as GeoJSON string
	// Can be Point OR Polygon depending on feature
	GeometryGeoJSON string `json:"geometry"` // GEOMETRY(Geometry, 4326) as GeoJSON - REQUIRED

	// Height for shade calculations
	HeightM *float64 `json:"height_m,omitempty"`

	// For trees specifically
	CanopyDiameterM *float64 `json:"canopy_diameter_m,omitempty"`
	Deciduous       *bool    `json:"deciduous,omitempty"` // true = loses leaves in winter

	CreatedAt time.Time `json:"created_at"`
}

// Validate validates the garden feature entity
func (gf *GardenFeature) Validate() error {
	if gf.GardenID == "" {
		return fmt.Errorf("garden_id is required")
	}

	if !isValidFeatureType(gf.FeatureType) {
		return fmt.Errorf("invalid feature_type: must be tree, shrub, building, fence, wall, greenhouse, shed, pond, or path")
	}

	if gf.GeometryGeoJSON == "" {
		return fmt.Errorf("geometry is required")
	}

	// Validate feature name length if provided
	if gf.FeatureName != nil && len(*gf.FeatureName) > 200 {
		return fmt.Errorf("feature_name must be 200 characters or less")
	}

	// Validate height if provided
	if gf.HeightM != nil && *gf.HeightM < 0 {
		return fmt.Errorf("height_m must be non-negative")
	}

	// Validate canopy diameter if provided
	if gf.CanopyDiameterM != nil && *gf.CanopyDiameterM < 0 {
		return fmt.Errorf("canopy_diameter_m must be non-negative")
	}

	// For trees, deciduous should be set
	if gf.FeatureType == FeatureTypeTree && gf.Deciduous == nil {
		return fmt.Errorf("deciduous is required for tree features")
	}

	return nil
}

// isValidFeatureType checks if feature type value is valid
func isValidFeatureType(featureType FeatureType) bool {
	validTypes := []FeatureType{
		FeatureTypeTree, FeatureTypeShrub, FeatureTypeBuilding,
		FeatureTypeFence, FeatureTypeWall, FeatureTypeGreenhouse,
		FeatureTypeShed, FeatureTypePond, FeatureTypePath,
	}

	for _, valid := range validTypes {
		if featureType == valid {
			return true
		}
	}
	return false
}

// IsShadeProducing returns true if this feature can produce shade
func (gf *GardenFeature) IsShadeProducing() bool {
	return gf.HeightM != nil && *gf.HeightM > 0
}
