package entity

import (
	"fmt"
	"time"
)

// CountryPlant represents country-specific plant information
type CountryPlant struct {
	CountryPlantID     string    `json:"country_plant_id"`
	CountryID          string    `json:"country_id"`
	PlantID            string    `json:"plant_id"`
	NativeStatus       *string   `json:"native_status,omitempty"`       // native, endemic, naturalized, etc.
	LegalStatus        *string   `json:"legal_status,omitempty"`        // prohibited, restricted, unrestricted, protected
	NativeRangeGeoJSON *string   `json:"native_range,omitempty"`        // GeoJSON MultiPolygon
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Validate validates the country plant entity
func (cp *CountryPlant) Validate() error {
	if cp.CountryID == "" {
		return fmt.Errorf("country_id is required")
	}

	if cp.PlantID == "" {
		return fmt.Errorf("plant_id is required")
	}

	// Validate native status if provided
	if cp.NativeStatus != nil {
		validStatuses := map[string]bool{
			"native":          true,
			"endemic":         true,
			"naturalized":     true,
			"introduced":      true,
			"invasive":        true,
			"cultivated_only": true,
		}

		if !validStatuses[*cp.NativeStatus] {
			return fmt.Errorf("invalid native_status: %s", *cp.NativeStatus)
		}
	}

	// Validate legal status if provided
	if cp.LegalStatus != nil {
		validStatuses := map[string]bool{
			"prohibited":    true,
			"restricted":    true,
			"unrestricted":  true,
			"protected":     true,
		}

		if !validStatuses[*cp.LegalStatus] {
			return fmt.Errorf("invalid legal_status: %s", *cp.LegalStatus)
		}
	}

	return nil
}
