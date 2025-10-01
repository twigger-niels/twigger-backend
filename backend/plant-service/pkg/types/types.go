package types

// PlantType represents the classification of a plant
type PlantType string

const (
	PlantTypeTree      PlantType = "tree"
	PlantTypeShrub     PlantType = "shrub"
	PlantTypePerennial PlantType = "perennial"
	PlantTypeAnnual    PlantType = "annual"
	PlantTypeBiennial  PlantType = "biennial"
	PlantTypeBulb      PlantType = "bulb"
	PlantTypeGrass     PlantType = "grass"
	PlantTypeFern      PlantType = "fern"
	PlantTypeClimber   PlantType = "climber"
	PlantTypeAquatic   PlantType = "aquatic"
	PlantTypeSucculent PlantType = "succulent"
	PlantTypePalm      PlantType = "palm"
	PlantTypeBamboo    PlantType = "bamboo"
	PlantTypeOrchid    PlantType = "orchid"
	PlantTypeVine      PlantType = "vine"
)

// IsValid checks if the PlantType is valid
func (pt PlantType) IsValid() bool {
	switch pt {
	case PlantTypeTree, PlantTypeShrub, PlantTypePerennial, PlantTypeAnnual,
		PlantTypeBiennial, PlantTypeBulb, PlantTypeGrass, PlantTypeFern,
		PlantTypeClimber, PlantTypeAquatic, PlantTypeSucculent, PlantTypePalm,
		PlantTypeBamboo, PlantTypeOrchid, PlantTypeVine:
		return true
	}
	return false
}

// ConfidenceLevel represents confidence in data accuracy
type ConfidenceLevel string

const (
	ConfidenceVeryLow  ConfidenceLevel = "very_low"   // 0-20%
	ConfidenceLow      ConfidenceLevel = "low"        // 20-40%
	ConfidenceModerate ConfidenceLevel = "moderate"   // 40-60%
	ConfidenceProbable ConfidenceLevel = "probable"   // 60-80%
	ConfidenceVeryHigh ConfidenceLevel = "very_high"  // 80-95%
	ConfidenceConfirmed ConfidenceLevel = "confirmed" // 95-100%
)

// IsValid checks if the ConfidenceLevel is valid
func (cl ConfidenceLevel) IsValid() bool {
	switch cl {
	case ConfidenceVeryLow, ConfidenceLow, ConfidenceModerate,
		ConfidenceProbable, ConfidenceVeryHigh, ConfidenceConfirmed:
		return true
	}
	return false
}

// Season represents seasons of the year
type Season string

const (
	SeasonSpring  Season = "spring"
	SeasonSummer  Season = "summer"
	SeasonAutumn  Season = "autumn"
	SeasonWinter  Season = "winter"
	SeasonAllYear Season = "all_year"
)

// IsValid checks if the Season is valid
func (s Season) IsValid() bool {
	switch s {
	case SeasonSpring, SeasonSummer, SeasonAutumn, SeasonWinter, SeasonAllYear:
		return true
	}
	return false
}

// SunRequirement represents sunlight requirements
type SunRequirement string

const (
	SunFullSun         SunRequirement = "full_sun"
	SunPartialSun      SunRequirement = "partial_sun"
	SunPartialShade    SunRequirement = "partial_shade"
	SunFullShade       SunRequirement = "full_shade"
	SunMorningSun      SunRequirement = "morning_sun"
	SunAfternoonShade  SunRequirement = "afternoon_shade"
	SunDappledShade    SunRequirement = "dappled_shade"
)

// IsValid checks if the SunRequirement is valid
func (sr SunRequirement) IsValid() bool {
	switch sr {
	case SunFullSun, SunPartialSun, SunPartialShade, SunFullShade,
		SunMorningSun, SunAfternoonShade, SunDappledShade:
		return true
	}
	return false
}

// WaterNeeds represents water requirements
type WaterNeeds string

const (
	WaterVeryDry  WaterNeeds = "very_dry"
	WaterDry      WaterNeeds = "dry"
	WaterModerate WaterNeeds = "moderate"
	WaterMoist    WaterNeeds = "moist"
	WaterWet      WaterNeeds = "wet"
	WaterAquatic  WaterNeeds = "aquatic"
	WaterBog      WaterNeeds = "bog"
)

// IsValid checks if the WaterNeeds is valid
func (wn WaterNeeds) IsValid() bool {
	switch wn {
	case WaterVeryDry, WaterDry, WaterModerate, WaterMoist,
		WaterWet, WaterAquatic, WaterBog:
		return true
	}
	return false
}

// SoilDrainage represents soil drainage characteristics
type SoilDrainage string

const (
	DrainageVeryWell SoilDrainage = "very_well_drained"
	DrainageWell     SoilDrainage = "well_drained"
	DrainageModerate SoilDrainage = "moderate_drainage"
	DrainagePoor     SoilDrainage = "poorly_drained"
	DrainageWaterlog SoilDrainage = "waterlogged"
)

// IsValid checks if the SoilDrainage is valid
func (sd SoilDrainage) IsValid() bool {
	switch sd {
	case DrainageVeryWell, DrainageWell, DrainageModerate,
		DrainagePoor, DrainageWaterlog:
		return true
	}
	return false
}

// GrowthRate represents how fast a plant grows
type GrowthRate string

const (
	GrowthVerySlow GrowthRate = "very_slow"
	GrowthSlow     GrowthRate = "slow"
	GrowthModerate GrowthRate = "moderate"
	GrowthFast     GrowthRate = "fast"
	GrowthVeryFast GrowthRate = "very_fast"
)

// IsValid checks if the GrowthRate is valid
func (gr GrowthRate) IsValid() bool {
	switch gr {
	case GrowthVerySlow, GrowthSlow, GrowthModerate, GrowthFast, GrowthVeryFast:
		return true
	}
	return false
}

// RelationshipType represents companion plant relationship
type RelationshipType string

const (
	RelationshipBeneficial   RelationshipType = "beneficial"
	RelationshipAntagonistic RelationshipType = "antagonistic"
	RelationshipNeutral      RelationshipType = "neutral"
)

// IsValid checks if the RelationshipType is valid
func (rt RelationshipType) IsValid() bool {
	switch rt {
	case RelationshipBeneficial, RelationshipAntagonistic, RelationshipNeutral:
		return true
	}
	return false
}

// NativeStatus represents regional plant status
type NativeStatus string

const (
	NativeStatusNative         NativeStatus = "native"
	NativeStatusEndemic        NativeStatus = "endemic"
	NativeStatusNaturalized    NativeStatus = "naturalized"
	NativeStatusIntroduced     NativeStatus = "introduced"
	NativeStatusInvasive       NativeStatus = "invasive"
	NativeStatusCultivatedOnly NativeStatus = "cultivated_only"
)

// IsValid checks if the NativeStatus is valid
func (ns NativeStatus) IsValid() bool {
	switch ns {
	case NativeStatusNative, NativeStatusEndemic, NativeStatusNaturalized,
		NativeStatusIntroduced, NativeStatusInvasive, NativeStatusCultivatedOnly:
		return true
	}
	return false
}

// LegalStatus represents legal restrictions
type LegalStatus string

const (
	LegalProhibited   LegalStatus = "prohibited"
	LegalRestricted   LegalStatus = "restricted"
	LegalUnrestricted LegalStatus = "unrestricted"
	LegalProtected    LegalStatus = "protected"
)

// IsValid checks if the LegalStatus is valid
func (ls LegalStatus) IsValid() bool {
	switch ls {
	case LegalProhibited, LegalRestricted, LegalUnrestricted, LegalProtected:
		return true
	}
	return false
}
