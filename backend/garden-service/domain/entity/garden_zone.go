package entity

import (
	"fmt"
	"time"
)

// ZoneType represents the type of zone within a garden
type ZoneType string

const (
	ZoneTypeBed       ZoneType = "bed"
	ZoneTypeBorder    ZoneType = "border"
	ZoneTypeLawn      ZoneType = "lawn"
	ZoneTypePath      ZoneType = "path"
	ZoneTypeWater     ZoneType = "water"
	ZoneTypeStructure ZoneType = "structure"
	ZoneTypeCompost   ZoneType = "compost"
)

// IrrigationType represents irrigation method
type IrrigationType string

const (
	IrrigationNone      IrrigationType = "none"
	IrrigationDrip      IrrigationType = "drip"
	IrrigationSprinkler IrrigationType = "sprinkler"
	IrrigationSoaker    IrrigationType = "soaker"
	IrrigationManual    IrrigationType = "manual"
)

// GardenZone represents a zone/bed within a garden
type GardenZone struct {
	ZoneID      string  `json:"zone_id"`
	GardenID    string  `json:"garden_id"`
	ZoneName    *string `json:"zone_name,omitempty"`
	ZoneType    *ZoneType `json:"zone_type,omitempty"`

	// Spatial representation - stored as GeoJSON string
	GeometryGeoJSON string   `json:"geometry"`           // GEOMETRY(Polygon, 4326) as GeoJSON - REQUIRED
	AreaM2          *float64 `json:"area_m2,omitempty"`  // Generated column from ST_Area

	// Zone characteristics
	SoilType       *string          `json:"soil_type,omitempty"`
	SoilAmended    bool             `json:"soil_amended"`
	IrrigationType *IrrigationType  `json:"irrigation_type,omitempty"`

	// Sun exposure (hours)
	SunHoursSummer *int `json:"sun_hours_summer,omitempty"` // 0-24
	SunHoursWinter *int `json:"sun_hours_winter,omitempty"` // 0-24

	CreatedAt time.Time `json:"created_at"`
}

// Validate validates the garden zone entity
func (gz *GardenZone) Validate() error {
	if gz.GardenID == "" {
		return fmt.Errorf("garden_id is required")
	}

	if gz.GeometryGeoJSON == "" {
		return fmt.Errorf("geometry is required")
	}

	// Validate zone name length if provided
	if gz.ZoneName != nil && len(*gz.ZoneName) > 100 {
		return fmt.Errorf("zone_name must be 100 characters or less")
	}

	// Validate zone type if provided
	if gz.ZoneType != nil {
		if !isValidZoneType(*gz.ZoneType) {
			return fmt.Errorf("invalid zone_type: must be bed, border, lawn, path, water, structure, or compost")
		}
	}

	// Validate irrigation type if provided
	if gz.IrrigationType != nil {
		if !isValidIrrigationType(*gz.IrrigationType) {
			return fmt.Errorf("invalid irrigation_type: must be none, drip, sprinkler, soaker, or manual")
		}
	}

	// Validate sun hours if provided
	if gz.SunHoursSummer != nil {
		if *gz.SunHoursSummer < 0 || *gz.SunHoursSummer > 24 {
			return fmt.Errorf("sun_hours_summer must be between 0 and 24")
		}
	}

	if gz.SunHoursWinter != nil {
		if *gz.SunHoursWinter < 0 || *gz.SunHoursWinter > 24 {
			return fmt.Errorf("sun_hours_winter must be between 0 and 24")
		}
	}

	return nil
}

// isValidZoneType checks if zone type value is valid
func isValidZoneType(zoneType ZoneType) bool {
	validTypes := []ZoneType{
		ZoneTypeBed, ZoneTypeBorder, ZoneTypeLawn, ZoneTypePath,
		ZoneTypeWater, ZoneTypeStructure, ZoneTypeCompost,
	}

	for _, valid := range validTypes {
		if zoneType == valid {
			return true
		}
	}
	return false
}

// isValidIrrigationType checks if irrigation type value is valid
func isValidIrrigationType(irrigationType IrrigationType) bool {
	validTypes := []IrrigationType{
		IrrigationNone, IrrigationDrip, IrrigationSprinkler,
		IrrigationSoaker, IrrigationManual,
	}

	for _, valid := range validTypes {
		if irrigationType == valid {
			return true
		}
	}
	return false
}
