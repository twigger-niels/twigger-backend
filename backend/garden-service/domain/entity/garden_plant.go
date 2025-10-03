package entity

import (
	"fmt"
	"time"
)

// HealthStatus represents the health status of a planted specimen
type HealthStatus string

const (
	HealthStatusThriving   HealthStatus = "thriving"
	HealthStatusHealthy    HealthStatus = "healthy"
	HealthStatusStruggling HealthStatus = "struggling"
	HealthStatusDiseased   HealthStatus = "diseased"
	HealthStatusDead       HealthStatus = "dead"
)

// GardenPlant represents a planted specimen in a garden
type GardenPlant struct {
	GardenPlantID string  `json:"garden_plant_id"`
	GardenID      string  `json:"garden_id"`
	ZoneID        *string `json:"zone_id,omitempty"` // Optional - plant may not be in a zone
	PlantID       string  `json:"plant_id"`          // FK to plants table

	// Location - stored as GeoJSON string
	LocationGeoJSON string `json:"location"` // GEOMETRY(Point, 4326) as GeoJSON - REQUIRED

	// Timing
	PlantedDate *time.Time `json:"planted_date,omitempty"`
	RemovedDate *time.Time `json:"removed_date,omitempty"` // NULL if still active

	// Plant specifics
	Quantity    int     `json:"quantity"`           // Default 1
	PlantSource *string `json:"plant_source,omitempty"` // 'seed', 'cutting', 'nursery', etc.

	// Health tracking
	HealthStatus *HealthStatus `json:"health_status,omitempty"`

	// Notes and observations
	Notes *string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the garden plant entity
func (gp *GardenPlant) Validate() error {
	if gp.GardenID == "" {
		return fmt.Errorf("garden_id is required")
	}

	if gp.PlantID == "" {
		return fmt.Errorf("plant_id is required")
	}

	if gp.LocationGeoJSON == "" {
		return fmt.Errorf("location is required")
	}

	// Validate quantity
	if gp.Quantity < 1 {
		return fmt.Errorf("quantity must be at least 1")
	}

	// Validate health status if provided
	if gp.HealthStatus != nil {
		if !isValidHealthStatus(*gp.HealthStatus) {
			return fmt.Errorf("invalid health_status: must be thriving, healthy, struggling, diseased, or dead")
		}
	}

	// Validate dates
	if gp.PlantedDate != nil && gp.RemovedDate != nil {
		if gp.RemovedDate.Before(*gp.PlantedDate) {
			return fmt.Errorf("removed_date cannot be before planted_date")
		}
	}

	return nil
}

// isValidHealthStatus checks if health status value is valid
func isValidHealthStatus(status HealthStatus) bool {
	validStatuses := []HealthStatus{
		HealthStatusThriving,
		HealthStatusHealthy,
		HealthStatusStruggling,
		HealthStatusDiseased,
		HealthStatusDead,
	}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// IsActive returns true if the plant is still in the garden (not removed)
func (gp *GardenPlant) IsActive() bool {
	return gp.RemovedDate == nil
}

// IsHealthy returns true if the plant is in good health
func (gp *GardenPlant) IsHealthy() bool {
	if gp.HealthStatus == nil {
		return true // Assume healthy if no status set
	}

	return *gp.HealthStatus == HealthStatusThriving || *gp.HealthStatus == HealthStatusHealthy
}
