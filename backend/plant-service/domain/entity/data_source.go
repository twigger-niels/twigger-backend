package entity

import (
	"fmt"
	"time"
)

// DataSource represents a source of plant data
type DataSource struct {
	SourceID          string     `json:"source_id"`
	SourceName        string     `json:"source_name"`
	SourceType        string     `json:"source_type"` // botanical_garden, university, etc.
	ReliabilityScore  int        `json:"reliability_score"` // 1-5
	WebsiteURL        *string    `json:"website_url,omitempty"`
	LastVerified      *time.Time `json:"last_verified,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// Validate validates the data source entity
func (ds *DataSource) Validate() error {
	if ds.SourceName == "" {
		return fmt.Errorf("source_name is required")
	}

	if ds.SourceType == "" {
		return fmt.Errorf("source_type is required")
	}

	validTypes := map[string]bool{
		"botanical_garden":   true,
		"university":         true,
		"government_db":      true,
		"commercial_nursery": true,
		"book":               true,
		"website":            true,
		"expert":             true,
		"observation":        true,
	}

	if !validTypes[ds.SourceType] {
		return fmt.Errorf("invalid source_type: %s", ds.SourceType)
	}

	if ds.ReliabilityScore < 1 || ds.ReliabilityScore > 5 {
		return fmt.Errorf("reliability_score must be between 1 and 5")
	}

	return nil
}
