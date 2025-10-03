package database

import (
	"encoding/json"
	"fmt"
)

// GeoJSONGeometry represents a basic GeoJSON geometry structure
type GeoJSONGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// ValidateGeoJSON validates a GeoJSON string before passing to ST_GeomFromGeoJSON
// Checks for:
// - Valid JSON structure
// - Correct geometry type
// - Coordinates exist
func ValidateGeoJSON(geojsonStr string) error {
	if geojsonStr == "" {
		return fmt.Errorf("geojson cannot be empty")
	}

	// Parse as JSON first
	var geom GeoJSONGeometry
	if err := json.Unmarshal([]byte(geojsonStr), &geom); err != nil {
		return fmt.Errorf("invalid geojson: not valid JSON: %w", err)
	}

	// Validate type field exists
	if geom.Type == "" {
		return fmt.Errorf("invalid geojson: 'type' field is required")
	}

	// Validate type is a known geometry type
	validTypes := map[string]bool{
		"Point":              true,
		"MultiPoint":         true,
		"LineString":         true,
		"MultiLineString":    true,
		"Polygon":            true,
		"MultiPolygon":       true,
		"GeometryCollection": true,
	}

	if !validTypes[geom.Type] {
		return fmt.Errorf("invalid geojson: unknown type '%s'", geom.Type)
	}

	// Validate coordinates exist (not applicable for GeometryCollection)
	if geom.Type != "GeometryCollection" && len(geom.Coordinates) == 0 {
		return fmt.Errorf("invalid geojson: 'coordinates' field is required for type '%s'", geom.Type)
	}

	return nil
}

// ValidatePolygonClosure validates that a Polygon geometry has closed rings
// (first and last coordinates must be identical)
func ValidatePolygonClosure(geojsonStr string) error {
	var geom struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	}

	if err := json.Unmarshal([]byte(geojsonStr), &geom); err != nil {
		return fmt.Errorf("invalid polygon geojson: %w", err)
	}

	if geom.Type != "Polygon" && geom.Type != "MultiPolygon" {
		return nil // Not a polygon, skip closure check
	}

	// For Polygon: coordinates is array of rings
	// For MultiPolygon: coordinates is array of polygons (array of rings)
	for _, ring := range geom.Coordinates {
		if len(ring) < 4 {
			return fmt.Errorf("polygon ring must have at least 4 points")
		}

		// Check if first and last points are identical
		first := ring[0]
		last := ring[len(ring)-1]

		if len(first) < 2 || len(last) < 2 {
			return fmt.Errorf("polygon coordinates must have at least 2 dimensions")
		}

		if first[0] != last[0] || first[1] != last[1] {
			return fmt.Errorf("polygon ring is not closed: first point %v != last point %v", first, last)
		}
	}

	return nil
}
