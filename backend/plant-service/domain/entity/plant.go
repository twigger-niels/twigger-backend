package entity

import (
	"fmt"
	"strings"
	"time"

	"twigger-backend/backend/plant-service/pkg/types"
)

// Plant represents a plant entity (species or cultivar)
type Plant struct {
	// Core identification
	PlantID     string  `json:"plant_id"`
	SpeciesID   string  `json:"species_id"`
	CultivarID  *string `json:"cultivar_id,omitempty"`

	// Taxonomy information
	FamilyName      string           `json:"family_name"`
	GenusName       string           `json:"genus_name"`
	SpeciesName     string           `json:"species_name"`
	CultivarName    *string          `json:"cultivar_name,omitempty"`
	CommonNames     []string         `json:"common_names,omitempty"`
	PlantType       types.PlantType  `json:"plant_type"`

	// Full botanical name (generated from taxonomy)
	FullBotanicalName string `json:"full_botanical_name"`

	// Aggregated data from related tables
	GrowingConditions      *types.GrowingConditions      `json:"growing_conditions,omitempty"`
	PhysicalCharacteristics *types.PhysicalCharacteristics `json:"physical_characteristics,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks if the Plant entity is valid
func (p *Plant) Validate() error {
	if p == nil {
		return fmt.Errorf("plant cannot be nil")
	}

	// Check required fields
	if p.PlantID == "" {
		return fmt.Errorf("plant_id is required")
	}
	if p.SpeciesID == "" {
		return fmt.Errorf("species_id is required")
	}
	if p.FamilyName == "" {
		return fmt.Errorf("family_name is required")
	}
	if p.GenusName == "" {
		return fmt.Errorf("genus_name is required")
	}
	if p.SpeciesName == "" {
		return fmt.Errorf("species_name is required")
	}

	// Validate plant type
	if !p.PlantType.IsValid() {
		return fmt.Errorf("invalid plant type: %s", p.PlantType)
	}

	// Validate nested value objects
	if p.GrowingConditions != nil {
		if err := p.GrowingConditions.Validate(); err != nil {
			return fmt.Errorf("invalid growing conditions: %w", err)
		}
	}

	if p.PhysicalCharacteristics != nil {
		if err := p.PhysicalCharacteristics.Validate(); err != nil {
			return fmt.Errorf("invalid physical characteristics: %w", err)
		}
	}

	return nil
}

// GenerateBotanicalName creates the full botanical name from taxonomy
func (p *Plant) GenerateBotanicalName() string {
	parts := []string{p.GenusName, p.SpeciesName}

	if p.CultivarName != nil && *p.CultivarName != "" {
		parts = append(parts, fmt.Sprintf("'%s'", *p.CultivarName))
	}

	return strings.Join(parts, " ")
}

// UpdateBotanicalName updates the FullBotanicalName field
func (p *Plant) UpdateBotanicalName() {
	p.FullBotanicalName = p.GenerateBotanicalName()
}

// IsCultivar checks if this is a cultivar (has cultivar_id)
func (p *Plant) IsCultivar() bool {
	return p.CultivarID != nil && *p.CultivarID != ""
}

// GetDisplayName returns the best display name (common name or botanical)
func (p *Plant) GetDisplayName() string {
	if len(p.CommonNames) > 0 && p.CommonNames[0] != "" {
		return p.CommonNames[0]
	}
	return p.FullBotanicalName
}

// HasCommonName checks if the plant has a specific common name
func (p *Plant) HasCommonName(name string) bool {
	searchName := strings.ToLower(strings.TrimSpace(name))
	for _, cn := range p.CommonNames {
		if strings.ToLower(strings.TrimSpace(cn)) == searchName {
			return true
		}
	}
	return false
}

// SearchScore calculates a relevance score for search results
// Higher score means better match
func (p *Plant) SearchScore(query string) int {
	query = strings.ToLower(strings.TrimSpace(query))
	score := 0

	// Exact botanical name match (highest priority)
	if strings.ToLower(p.FullBotanicalName) == query {
		score += 100
	} else if strings.Contains(strings.ToLower(p.FullBotanicalName), query) {
		score += 50
	}

	// Genus match
	if strings.ToLower(p.GenusName) == query {
		score += 80
	} else if strings.HasPrefix(strings.ToLower(p.GenusName), query) {
		score += 40
	}

	// Species match
	if strings.ToLower(p.SpeciesName) == query {
		score += 70
	} else if strings.HasPrefix(strings.ToLower(p.SpeciesName), query) {
		score += 35
	}

	// Common name matches
	for i, cn := range p.CommonNames {
		cnLower := strings.ToLower(cn)
		if cnLower == query {
			// First common name gets higher score
			if i == 0 {
				score += 90
			} else {
				score += 60
			}
		} else if strings.Contains(cnLower, query) {
			score += 30
		}
	}

	// Family match (lower priority)
	if strings.ToLower(p.FamilyName) == query {
		score += 20
	}

	return score
}

// CanGrowInConditions checks if the plant can grow in given conditions
func (p *Plant) CanGrowInConditions(hardinessZone string, sunReq types.SunRequirement) bool {
	if p.GrowingConditions == nil {
		return false
	}

	// Check hardiness zone
	hasZone := false
	if len(p.GrowingConditions.HardinessZones) == 0 {
		// No zones specified, assume it can grow anywhere
		hasZone = true
	} else {
		hasZone = p.GrowingConditions.SupportsHardinessZone(hardinessZone)
	}

	// Check sun requirements
	hasSun := false
	if len(p.GrowingConditions.SunRequirements) == 0 {
		// No sun requirements specified
		hasSun = true
	} else {
		hasSun = p.GrowingConditions.HasSunRequirement(sunReq)
	}

	return hasZone && hasSun
}

// GetMatureHeightRange returns the typical mature height in meters
func (p *Plant) GetMatureHeightRange() *types.SizeRange {
	if p.PhysicalCharacteristics == nil {
		return nil
	}
	return p.PhysicalCharacteristics.MatureHeight
}

// GetMatureSpreadRange returns the typical mature spread in meters
func (p *Plant) GetMatureSpreadRange() *types.SizeRange {
	if p.PhysicalCharacteristics == nil {
		return nil
	}
	return p.PhysicalCharacteristics.MatureSpread
}

// IsFloweringInMonth checks if plant flowers in a specific month
func (p *Plant) IsFloweringInMonth(month int) bool {
	if p.GrowingConditions == nil {
		return false
	}
	return p.GrowingConditions.IsFloweringInMonth(month)
}

// IsToxic checks if the plant is toxic
func (p *Plant) IsToxic() bool {
	if p.PhysicalCharacteristics == nil {
		return false
	}
	return p.PhysicalCharacteristics.IsToxic()
}

// AttractsWildlife checks if the plant attracts specific wildlife
func (p *Plant) AttractsWildlife(wildlife string) bool {
	if p.PhysicalCharacteristics == nil {
		return false
	}
	return p.PhysicalCharacteristics.HasWildlifeValue(wildlife)
}
