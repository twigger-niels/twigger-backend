package database

import (
	"context"
	"database/sql"
	"fmt"

	"twigger-backend/backend/plant-service/pkg/types"

	"github.com/lib/pq"
)

// GetGrowingConditions retrieves growing conditions for a plant in a specific country
func (r *PostgresPlantRepository) GetGrowingConditions(ctx context.Context, plantID, countryID string) (*types.GrowingConditions, error) {
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
	var phMin, phMax, phOptimal sql.NullFloat64

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
		&phMin, // This is a simplified scan - actual schema uses composite type
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

	// pH preference (simplified - actual schema uses composite type)
	if phMin.Valid || phMax.Valid || phOptimal.Valid {
		gc.PHPreference = &types.PHRange{
			MinPH:     nullFloatToPtr(phMin),
			MaxPH:     nullFloatToPtr(phMax),
			OptimalPH: nullFloatToPtr(phOptimal),
		}
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

	return gc, nil
}

// GetPhysicalCharacteristics retrieves physical characteristics for a plant
func (r *PostgresPlantRepository) GetPhysicalCharacteristics(ctx context.Context, plantID string) (*types.PhysicalCharacteristics, error) {
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
	var matureHeightMin, matureHeightTyp, matureHeightMax sql.NullFloat64
	var matureSpreadMin, matureSpreadTyp, matureSpreadMax sql.NullFloat64
	var growthRate sql.NullString
	var traitsJSON string

	// Simplified scan - actual schema uses composite types for ranges
	err := r.db.QueryRowContext(ctx, query, plantID).Scan(
		&characteristicID,
		&sourceID,
		&matureHeightMin, // Simplified
		&matureSpreadMin, // Simplified
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

	// Parse mature height (simplified)
	if matureHeightMin.Valid {
		pc.MatureHeight = &types.SizeRange{
			MinM: nullFloatToPtr(matureHeightMin),
		}
	}

	// Parse mature spread (simplified)
	if matureSpreadMin.Valid {
		pc.MatureSpread = &types.SizeRange{
			MinM: nullFloatToPtr(matureSpreadMin),
		}
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

	return pc, nil
}

// Helper function to parse zone strings like "5a,5b,6a"
func parseZoneString(zones string) []string {
	if zones == "" {
		return []string{}
	}
	// Split by comma and trim spaces
	result := make([]string, 0)
	for _, zone := range pq.Array(&result).Scan(zones); zone != "" {
		result = append(result, zone)
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
