package constants

// ValidClimateSystems contains all supported climate zone systems
var ValidClimateSystems = map[string]bool{
	"USDA": true, // United States Department of Agriculture hardiness zones
	"RHS":  true, // Royal Horticultural Society hardiness zones (UK)
	"AHS":  true, // American Horticultural Society heat zones
}

// Climate system constants
const (
	ClimateSystemUSDA = "USDA"
	ClimateSystemRHS  = "RHS"
	ClimateSystemAHS  = "AHS"
)

// IsValidClimateSystem checks if a climate system is supported
func IsValidClimateSystem(system string) bool {
	return ValidClimateSystems[system]
}

// GetValidClimateSystemsList returns a slice of all valid climate systems
func GetValidClimateSystemsList() []string {
	return []string{ClimateSystemUSDA, ClimateSystemRHS, ClimateSystemAHS}
}
