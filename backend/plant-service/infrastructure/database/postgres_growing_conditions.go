package database

import (
	"context"
	"database/sql"
	"fmt"

	"twigger-backend/backend/plant-service/pkg/types"

	"github.com/lib/pq"
)

// GetGrowingConditions retrieves growing conditions for a plant in a specific country with localized characteristic values
func (r *PostgresPlantRepository) GetGrowingConditions(ctx context.Context, plantID, countryID, languageID string) (*types.GrowingConditions, error) {
	query := `
		SELECT
			gca.assertion_id,
			gca.source_id,
			gca.confidence,
			gca.hardiness_zones,
			gca.heat_zones,
			gca.sun_requirements,
			gca.shade_tolerance,
			gca.water_needs,
			gca.humidity_preference,
			gca.drought_tolerant,
			gca.soil_types,
			gca.soil_drainage,
			gca.salt_tolerant,
			gca.wind_tolerant,
			gca.flowering_months,
			gca.fruiting_months,
			gca.ph_preference,
			gca.created_at
		FROM growing_conditions_assertions gca
		INNER JOIN country_plants cp ON gca.country_plant_id = cp.country_plant_id
		WHERE cp.plant_id = $1 AND cp.country_id = $2
		ORDER BY gca.confidence DESC
		LIMIT 1
	`

	gc := &types.GrowingConditions{
		PlantID:   plantID,
		CountryID: &countryID,
	}

	var assertionID, sourceID string
	var confidenceStr string
	var hardinessZones, heatZones sql.NullString
	var sunRequirements pq.StringArray
	var shadeTolerance, droughtTolerant, saltTolerant, windTolerant bool
	var waterNeeds, soilDrainage sql.NullString
	var humidityPreference sql.NullFloat64
	var soilTypes pq.StringArray
	var floweringMonths, fruitingMonths pq.Int64Array
	var phPreferenceStr sql.NullString // Composite type ph_range scanned as string

	err := r.db.QueryRowContext(ctx, query, plantID, countryID).Scan(
		&assertionID,
		&sourceID,
		&confidenceStr,
		&hardinessZones,
		&heatZones,
		&sunRequirements,
		&shadeTolerance,
		&waterNeeds,
		&humidityPreference,
		&droughtTolerant,
		&soilTypes,
		&soilDrainage,
		&saltTolerant,
		&windTolerant,
		&floweringMonths,
		&fruitingMonths,
		&phPreferenceStr, // Scan composite type as string
		&gc.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No growing conditions found for this country
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get growing conditions: %w", err)
	}

	// Map confidence
	gc.Confidence = types.ConfidenceLevel(confidenceStr)
	gc.SourceID = &sourceID

	// Parse hardiness zones
	if hardinessZones.Valid && hardinessZones.String != "" {
		// Zones stored as comma-separated string: "5a,5b,6a"
		gc.HardinessZones = parseZoneString(hardinessZones.String)
	}

	// Parse heat zones
	if heatZones.Valid && heatZones.String != "" {
		gc.HeatZones = parseZoneString(heatZones.String)
	}

	// Convert sun requirements
	gc.SunRequirements = make([]types.SunRequirement, 0, len(sunRequirements))
	for _, sr := range sunRequirements {
		gc.SunRequirements = append(gc.SunRequirements, types.SunRequirement(sr))
	}

	gc.ShadeTolerance = shadeTolerance
	gc.DroughtTolerant = droughtTolerant
	gc.SaltTolerant = saltTolerant
	gc.WindTolerant = windTolerant

	// Water needs
	if waterNeeds.Valid {
		wn := types.WaterNeeds(waterNeeds.String)
		gc.WaterNeeds = &wn
	}

	// Humidity
	if humidityPreference.Valid {
		gc.HumidityPreference = &humidityPreference.Float64
	}

	// Soil types
	gc.SoilTypes = make([]string, len(soilTypes))
	copy(gc.SoilTypes, soilTypes)

	// Soil drainage
	if soilDrainage.Valid {
		sd := types.SoilDrainage(soilDrainage.String)
		gc.SoilDrainage = &sd
	}

	// pH preference (parse composite type from string format)
	if phPreferenceStr.Valid && phPreferenceStr.String != "" {
		gc.PHPreference = parsePHRange(phPreferenceStr.String)
	}

	// Flowering months
	gc.FloweringMonths = make([]int, 0, len(floweringMonths))
	for _, m := range floweringMonths {
		gc.FloweringMonths = append(gc.FloweringMonths, int(m))
	}

	// Fruiting months
	gc.FruitingMonths = make([]int, 0, len(fruitingMonths))
	for _, m := range fruitingMonths {
		gc.FruitingMonths = append(gc.FruitingMonths, int(m))
	}

	// Translate characteristic values to user's language
	translator := NewCharacteristicTranslator(r.db)

	// Translate sun requirements
	if len(gc.SunRequirements) > 0 {
		sunReqStrings := make([]string, len(gc.SunRequirements))
		for i, sr := range gc.SunRequirements {
			sunReqStrings[i] = string(sr)
		}
		translated, _ := translator.TranslateArray(ctx, "sun_requirement", sunReqStrings, languageID)
		for i, t := range translated {
			gc.SunRequirements[i] = types.SunRequirement(t)
		}
	}

	// Translate water needs
	if gc.WaterNeeds != nil {
		translated, _ := translator.Translate(ctx, "water_needs", string(*gc.WaterNeeds), languageID)
		wn := types.WaterNeeds(translated)
		gc.WaterNeeds = &wn
	}

	// Translate soil drainage
	if gc.SoilDrainage != nil {
		translated, _ := translator.Translate(ctx, "soil_drainage", string(*gc.SoilDrainage), languageID)
		sd := types.SoilDrainage(translated)
		gc.SoilDrainage = &sd
	}

	return gc, nil
}

// GetPhysicalCharacteristics retrieves physical characteristics for a plant with localized growth rate
func (r *PostgresPlantRepository) GetPhysicalCharacteristics(ctx context.Context, plantID, languageID string) (*types.PhysicalCharacteristics, error) {
	query := `
		SELECT
			pc.characteristic_id,
			pc.source_id,
			pc.mature_height,
			pc.mature_spread,
			pc.growth_rate,
			pc.traits,
			pc.created_at
		FROM physical_characteristics pc
		WHERE pc.plant_id = $1
		ORDER BY pc.created_at DESC
		LIMIT 1
	`

	pc := &types.PhysicalCharacteristics{
		PlantID: plantID,
	}

	var characteristicID string
	var sourceID sql.NullString
	var matureHeightStr sql.NullString // Composite type size_range
	var matureSpreadStr sql.NullString // Composite type size_range
	var growthRate sql.NullString
	var traitsJSON string

	// Scan composite types as strings
	err := r.db.QueryRowContext(ctx, query, plantID).Scan(
		&characteristicID,
		&sourceID,
		&matureHeightStr, // Composite type scanned as string
		&matureSpreadStr, // Composite type scanned as string
		&growthRate,
		&traitsJSON,
		&pc.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No physical characteristics found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get physical characteristics: %w", err)
	}

	if sourceID.Valid {
		pc.SourceID = &sourceID.String
	}

	// Parse mature height (composite type)
	if matureHeightStr.Valid && matureHeightStr.String != "" {
		pc.MatureHeight = parseSizeRange(matureHeightStr.String)
	}

	// Parse mature spread (composite type)
	if matureSpreadStr.Valid && matureSpreadStr.String != "" {
		pc.MatureSpread = parseSizeRange(matureSpreadStr.String)
	}

	// Growth rate
	if growthRate.Valid {
		gr := types.GrowthRate(growthRate.String)
		pc.GrowthRate = &gr
	}

	// Parse traits from JSONB
	if traitsJSON != "" {
		if err := pc.TraitsFromJSON(traitsJSON); err != nil {
			return nil, fmt.Errorf("failed to parse traits: %w", err)
		}
	}

	// Translate growth rate to user's language
	if pc.GrowthRate != nil {
		translator := NewCharacteristicTranslator(r.db)
		translated, _ := translator.Translate(ctx, "growth_rate", string(*pc.GrowthRate), languageID)
		gr := types.GrowthRate(translated)
		pc.GrowthRate = &gr
	}

	return pc, nil
}

// Helper function to parse zone strings like "5a,5b,6a"
func parseZoneString(zones string) []string {
	if zones == "" {
		return []string{}
	}
	// Split by comma and trim spaces
	var result []string
	if err := pq.Array(&result).Scan(zones); err != nil {
		return []string{}
	}
	return result
}

// Helper function to convert sql.NullFloat64 to *float64
func nullFloatToPtr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

// Helper function to convert sql.NullString to *string
func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// parsePHRange parses a PostgreSQL composite type ph_range from string format
// Expected format: "(min,max,optimal)" e.g., "(6.0,7.5,6.8)"
func parsePHRange(s string) *types.PHRange {
	if s == "" || s == "(,,)" {
		return nil
	}

	// Remove parentheses
	s = s[1 : len(s)-1]

	// Split by comma
	parts := make([]string, 0, 3)
	var current string
	for _, char := range s {
		if char == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	parts = append(parts, current)

	if len(parts) != 3 {
		return nil
	}

	result := &types.PHRange{}

	// Parse min pH
	if parts[0] != "" {
		var minPH float64
		if _, err := fmt.Sscanf(parts[0], "%f", &minPH); err == nil {
			result.MinPH = &minPH
		}
	}

	// Parse max pH
	if parts[1] != "" {
		var maxPH float64
		if _, err := fmt.Sscanf(parts[1], "%f", &maxPH); err == nil {
			result.MaxPH = &maxPH
		}
	}

	// Parse optimal pH
	if parts[2] != "" {
		var optimalPH float64
		if _, err := fmt.Sscanf(parts[2], "%f", &optimalPH); err == nil {
			result.OptimalPH = &optimalPH
		}
	}

	return result
}

// parseSizeRange parses a PostgreSQL composite type size_range from string format
// Expected format: "(min,typical,max)" e.g., "(0.5,1.0,1.5)"
func parseSizeRange(s string) *types.SizeRange {
	if s == "" || s == "(,,)" {
		return nil
	}

	// Remove parentheses
	s = s[1 : len(s)-1]

	// Split by comma
	parts := make([]string, 0, 3)
	var current string
	for _, char := range s {
		if char == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	parts = append(parts, current)

	if len(parts) != 3 {
		return nil
	}

	result := &types.SizeRange{}

	// Parse min size
	if parts[0] != "" {
		var minM float64
		if _, err := fmt.Sscanf(parts[0], "%f", &minM); err == nil {
			result.MinM = &minM
		}
	}

	// Parse typical size
	if parts[1] != "" {
		var typicalM float64
		if _, err := fmt.Sscanf(parts[1], "%f", &typicalM); err == nil {
			result.TypicalM = &typicalM
		}
	}

	// Parse max size
	if parts[2] != "" {
		var maxM float64
		if _, err := fmt.Sscanf(parts[2], "%f", &maxM); err == nil {
			result.MaxM = &maxM
		}
	}

	return result
}
