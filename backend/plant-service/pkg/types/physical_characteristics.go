package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// PhysicalCharacteristics represents the physical attributes of a plant
type PhysicalCharacteristics struct {
	// Identification
	PlantID  string  `json:"plant_id"`
	SourceID *string `json:"source_id,omitempty"`

	// Core measurements
	MatureHeight *SizeRange  `json:"mature_height,omitempty"` // in meters
	MatureSpread *SizeRange  `json:"mature_spread,omitempty"` // in meters
	GrowthRate   *GrowthRate `json:"growth_rate,omitempty"`

	// Flexible traits stored as JSONB in database
	Traits map[string]interface{} `json:"traits,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
}

// Common trait keys used in the Traits map
const (
	TraitLeafColor      = "leaf_color"       // string or []string
	TraitLeafShape      = "leaf_shape"       // string
	TraitLeafTexture    = "leaf_texture"     // string
	TraitFlowerColor    = "flower_color"     // string or []string
	TraitFlowerShape    = "flower_shape"     // string
	TraitFlowerFragrant = "flower_fragrant"  // bool
	TraitFruitType      = "fruit_type"       // string
	TraitFruitColor     = "fruit_color"      // string or []string
	TraitFruitEdible    = "fruit_edible"     // bool
	TraitBarkTexture    = "bark_texture"     // string
	TraitBarkColor      = "bark_color"       // string
	TraitRootType       = "root_type"        // string: "taproot", "fibrous", "rhizome", "bulb", "tuber"
	TraitRootDepth      = "root_depth_m"     // float64
	TraitEvergreen      = "evergreen"        // bool
	TraitDeciduous      = "deciduous"        // bool
	TraitThorny         = "thorny"           // bool
	TraitFragrant       = "fragrant"         // bool
	TraitToxic          = "toxic"            // bool
	TraitToxicParts     = "toxic_parts"      // []string: ["leaves", "berries", "roots"]
	TraitWildlifeValue  = "wildlife_value"   // []string: ["birds", "bees", "butterflies"]
)

// Validate checks if the PhysicalCharacteristics are valid
func (pc *PhysicalCharacteristics) Validate() error {
	if pc == nil {
		return fmt.Errorf("physical characteristics cannot be nil")
	}

	// Check required fields
	if pc.PlantID == "" {
		return fmt.Errorf("plant_id is required")
	}

	// Validate size ranges
	if pc.MatureHeight != nil {
		if err := pc.MatureHeight.IsValid(); err != nil {
			return fmt.Errorf("invalid mature height: %w", err)
		}
	}

	if pc.MatureSpread != nil {
		if err := pc.MatureSpread.IsValid(); err != nil {
			return fmt.Errorf("invalid mature spread: %w", err)
		}
	}

	// Validate growth rate
	if pc.GrowthRate != nil && !pc.GrowthRate.IsValid() {
		return fmt.Errorf("invalid growth rate: %s", *pc.GrowthRate)
	}

	return nil
}

// GetStringTrait retrieves a string trait
func (pc *PhysicalCharacteristics) GetStringTrait(key string) (string, bool) {
	if pc.Traits == nil {
		return "", false
	}
	val, ok := pc.Traits[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetStringArrayTrait retrieves a string array trait
func (pc *PhysicalCharacteristics) GetStringArrayTrait(key string) ([]string, bool) {
	if pc.Traits == nil {
		return nil, false
	}
	val, ok := pc.Traits[key]
	if !ok {
		return nil, false
	}

	// Handle both []string and []interface{} from JSON
	switch v := val.(type) {
	case []string:
		return v, true
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result, true
	}
	return nil, false
}

// GetBoolTrait retrieves a boolean trait
func (pc *PhysicalCharacteristics) GetBoolTrait(key string) (bool, bool) {
	if pc.Traits == nil {
		return false, false
	}
	val, ok := pc.Traits[key]
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// GetFloat64Trait retrieves a float64 trait
func (pc *PhysicalCharacteristics) GetFloat64Trait(key string) (float64, bool) {
	if pc.Traits == nil {
		return 0, false
	}
	val, ok := pc.Traits[key]
	if !ok {
		return 0, false
	}

	// Handle both float64 and int from JSON
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	}
	return 0, false
}

// SetTrait sets a trait value
func (pc *PhysicalCharacteristics) SetTrait(key string, value interface{}) {
	if pc.Traits == nil {
		pc.Traits = make(map[string]interface{})
	}
	pc.Traits[key] = value
}

// IsEvergreen checks if the plant is evergreen
func (pc *PhysicalCharacteristics) IsEvergreen() bool {
	val, ok := pc.GetBoolTrait(TraitEvergreen)
	return ok && val
}

// IsDeciduous checks if the plant is deciduous
func (pc *PhysicalCharacteristics) IsDeciduous() bool {
	val, ok := pc.GetBoolTrait(TraitDeciduous)
	return ok && val
}

// IsToxic checks if the plant is toxic
func (pc *PhysicalCharacteristics) IsToxic() bool {
	val, ok := pc.GetBoolTrait(TraitToxic)
	return ok && val
}

// GetToxicParts returns which parts of the plant are toxic
func (pc *PhysicalCharacteristics) GetToxicParts() []string {
	parts, _ := pc.GetStringArrayTrait(TraitToxicParts)
	return parts
}

// GetWildlifeValue returns what wildlife the plant attracts
func (pc *PhysicalCharacteristics) GetWildlifeValue() []string {
	wildlife, _ := pc.GetStringArrayTrait(TraitWildlifeValue)
	return wildlife
}

// HasWildlifeValue checks if the plant attracts specific wildlife
func (pc *PhysicalCharacteristics) HasWildlifeValue(wildlife string) bool {
	values := pc.GetWildlifeValue()
	for _, v := range values {
		if v == wildlife {
			return true
		}
	}
	return false
}

// TraitsToJSON converts traits map to JSON string for database storage
func (pc *PhysicalCharacteristics) TraitsToJSON() (string, error) {
	if pc.Traits == nil || len(pc.Traits) == 0 {
		return "{}", nil
	}
	bytes, err := json.Marshal(pc.Traits)
	if err != nil {
		return "", fmt.Errorf("failed to marshal traits: %w", err)
	}
	return string(bytes), nil
}

// TraitsFromJSON parses traits from JSON string
func (pc *PhysicalCharacteristics) TraitsFromJSON(jsonStr string) error {
	if jsonStr == "" || jsonStr == "{}" {
		pc.Traits = make(map[string]interface{})
		return nil
	}

	if err := json.Unmarshal([]byte(jsonStr), &pc.Traits); err != nil {
		return fmt.Errorf("failed to unmarshal traits: %w", err)
	}
	return nil
}
