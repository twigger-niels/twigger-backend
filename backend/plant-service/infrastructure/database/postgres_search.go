package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
	"twigger-backend/backend/plant-service/pkg/types"

	"github.com/lib/pq"
)

// Search performs full-text search on plants with filtering and pagination
func (r *PostgresPlantRepository) Search(ctx context.Context, query string, filter *repository.SearchFilter, languageID string, countryID *string) (*repository.SearchResult, error) {
	if filter == nil {
		filter = repository.DefaultSearchFilter()
	}

	// Build the WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argPos := 1

	// Full-text search on botanical name AND common names
	// Use CTE to search across both botanical and common names
	searchInCommonNames := false
	if query != "" {
		searchInCommonNames = true
		// We'll use a CTE for this - see below
		argPos++ // Reserve parameter for query text
	}

	// Filter by family
	if filter.FamilyName != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(pf.family_name) = LOWER($%d)", argPos))
		args = append(args, *filter.FamilyName)
		argPos++
	}

	// Filter by genus
	if filter.GenusName != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(pg.genus_name) = LOWER($%d)", argPos))
		args = append(args, *filter.GenusName)
		argPos++
	}

	// Filter by species
	if filter.SpeciesName != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(ps.species_name) = LOWER($%d)", argPos))
		args = append(args, *filter.SpeciesName)
		argPos++
	}

	// Filter by plant type
	if filter.PlantType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("ps.plant_type = $%d", argPos))
		args = append(args, string(*filter.PlantType))
		argPos++
	}

	// Physical characteristic filters
	if filter.MinHeight != nil {
		// Height is stored as size_range composite type (min_m, typical_m, max_m)
		// User wants plants that can grow AT LEAST this tall - check if max_m >= MinHeight
		whereClauses = append(whereClauses, fmt.Sprintf("(pc.mature_height).max_m >= $%d", argPos))
		args = append(args, *filter.MinHeight)
		argPos++
	}

	if filter.MaxHeight != nil {
		// User wants plants that fit in this space - check if typical_m <= MaxHeight
		whereClauses = append(whereClauses, fmt.Sprintf("(pc.mature_height).typical_m <= $%d", argPos))
		args = append(args, *filter.MaxHeight)
		argPos++
	}

	if filter.GrowthRate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("pc.growth_rate = $%d", argPos))
		args = append(args, string(*filter.GrowthRate))
		argPos++
	}

	// Trait filters using JSONB
	if filter.Evergreen != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(pc.traits->>'evergreen')::boolean = $%d", argPos))
		args = append(args, *filter.Evergreen)
		argPos++
	}

	if filter.Deciduous != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(pc.traits->>'deciduous')::boolean = $%d", argPos))
		args = append(args, *filter.Deciduous)
		argPos++
	}

	if filter.Toxic != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(pc.traits->>'toxic')::boolean = $%d", argPos))
		args = append(args, *filter.Toxic)
		argPos++
	}

	// Add cursor filter if provided (for pagination)
	if filter.Cursor != nil && *filter.Cursor != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.plant_id > $%d", argPos))
		args = append(args, *filter.Cursor)
		argPos++
	}

	// Build WHERE clause
	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Build CTE for searching both botanical and common names if query provided
	cteSQL := ""
	countCTESQL := ""
	if searchInCommonNames {
		cteSQL = fmt.Sprintf(`
		WITH name_matches AS (
			SELECT DISTINCT p.plant_id
			FROM plants p
			LEFT JOIN plant_common_names pcn ON p.plant_id = pcn.plant_id
			WHERE (
				to_tsvector('english', COALESCE(p.full_botanical_name, '')) @@ plainto_tsquery('english', $1)
				OR to_tsvector('english', COALESCE(pcn.common_name, '')) @@ plainto_tsquery('english', $1)
			)
			%s
		)`,
		// Add language filter if we want to restrict common names by language
		// For now we search across all languages for better discoverability
		"")

		countCTESQL = cteSQL

		// Inject CTE into WHERE clause
		if whereSQL == "" {
			whereSQL = "WHERE p.plant_id IN (SELECT plant_id FROM name_matches)"
		} else {
			whereSQL += " AND p.plant_id IN (SELECT plant_id FROM name_matches)"
		}

		// Prepend query parameter to args
		args = append([]interface{}{query}, args...)
	}

	// Build ORDER BY clause
	orderBySQL := ""
	switch filter.SortBy {
	case repository.SortByBotanicalName:
		orderBySQL = "ORDER BY p.full_botanical_name"
	case repository.SortByFamilyName:
		orderBySQL = "ORDER BY pf.family_name, pg.genus_name, ps.species_name"
	case repository.SortByGenusName:
		orderBySQL = "ORDER BY pg.genus_name, ps.species_name"
	case repository.SortByCreatedAt:
		orderBySQL = "ORDER BY p.created_at"
	case repository.SortByRelevance:
		if query != "" && searchInCommonNames {
			// When searching common names, we can't easily compute relevance
			// across the CTE boundary, so fall back to botanical name sorting
			// TODO: Enhance with materialized relevance scores
			orderBySQL = "ORDER BY p.full_botanical_name"
		} else if query != "" {
			orderBySQL = "ORDER BY ts_rank(to_tsvector('english', COALESCE(p.full_botanical_name, '')), plainto_tsquery('english', $1))"
		} else {
			orderBySQL = "ORDER BY p.created_at"
		}
	default:
		orderBySQL = "ORDER BY p.created_at"
	}

	if filter.SortOrder == repository.SortAsc {
		orderBySQL += " ASC"
	} else {
		orderBySQL += " DESC"
	}

	// Always add plant_id to ORDER BY for consistent cursor pagination
	orderBySQL += ", p.plant_id ASC"

	// Count total results
	countQuery := fmt.Sprintf(`
		%s
		SELECT COUNT(*)
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		LEFT JOIN physical_characteristics pc ON p.plant_id = pc.plant_id
		%s
	`, countCTESQL, whereSQL)

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Build main query with pagination
	searchQuery := fmt.Sprintf(`
		%s
		SELECT
			p.plant_id,
			p.species_id,
			p.cultivar_id,
			p.full_botanical_name,
			ps.plant_type,
			ps.species_name,
			pg.genus_name,
			pf.family_name,
			c.cultivar_name,
			p.created_at
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		LEFT JOIN cultivars c ON p.cultivar_id = c.cultivar_id
		LEFT JOIN physical_characteristics pc ON p.plant_id = pc.plant_id
		%s
		%s
		LIMIT $%d
	`, cteSQL, whereSQL, orderBySQL, argPos)

	args = append(args, filter.Limit+1) // Fetch one extra to determine if there are more results

	rows, err := r.db.QueryContext(ctx, searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search plants: %w", err)
	}
	defer rows.Close()

	plants := make([]*entity.Plant, 0, filter.Limit)
	for rows.Next() {
		plant := &entity.Plant{}
		var cultivarID, cultivarName sql.NullString

		err := rows.Scan(
			&plant.PlantID,
			&plant.SpeciesID,
			&cultivarID,
			&plant.FullBotanicalName,
			&plant.PlantType,
			&plant.SpeciesName,
			&plant.GenusName,
			&plant.FamilyName,
			&cultivarName,
			&plant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plant: %w", err)
		}

		if cultivarID.Valid {
			plant.CultivarID = &cultivarID.String
		}
		if cultivarName.Valid {
			plant.CultivarName = &cultivarName.String
		}

		plants = append(plants, plant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	// Check if we have more results (we fetched limit+1)
	hasMore := len(plants) > filter.Limit
	var nextCursor *string
	if hasMore {
		// Remove the extra plant and use its ID as the next cursor
		lastPlant := plants[len(plants)-1]
		plants = plants[:filter.Limit]
		nextCursor = &lastPlant.PlantID
	}

	// Load common names for all plants (batch loading to avoid N+1 query)
	if err := r.loadCommonNamesForMultiplePlants(ctx, plants, languageID, countryID); err != nil {
		return nil, fmt.Errorf("failed to load common names: %w", err)
	}

	return &repository.SearchResult{
		Plants:     plants,
		Total:      total,
		Limit:      filter.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Query:      query,
	}, nil
}

// FindByFamily finds all plants in a family
func (r *PostgresPlantRepository) FindByFamily(ctx context.Context, familyName string, languageID string, countryID *string, limit, offset int) ([]*entity.Plant, error) {
	filter := repository.DefaultSearchFilter()
	filter.FamilyName = &familyName
	filter.Limit = limit
	// Note: offset parameter ignored - cursor-based pagination should be used via Search() directly
	filter.SortBy = repository.SortByGenusName

	result, err := r.Search(ctx, "", filter, languageID, countryID)
	if err != nil {
		return nil, err
	}

	return result.Plants, nil
}

// FindByGenus finds all plants in a genus
func (r *PostgresPlantRepository) FindByGenus(ctx context.Context, genusName string, languageID string, countryID *string, limit, offset int) ([]*entity.Plant, error) {
	filter := repository.DefaultSearchFilter()
	filter.GenusName = &genusName
	filter.Limit = limit
	// Note: offset parameter ignored - cursor-based pagination should be used via Search() directly
	filter.SortBy = repository.SortByBotanicalName

	result, err := r.Search(ctx, "", filter, languageID, countryID)
	if err != nil {
		return nil, err
	}

	return result.Plants, nil
}

// FindBySpecies finds plants by genus and species name
func (r *PostgresPlantRepository) FindBySpecies(ctx context.Context, genusName, speciesName, languageID string, countryID *string) ([]*entity.Plant, error) {
	query := `
		SELECT
			p.plant_id,
			p.species_id,
			p.cultivar_id,
			p.full_botanical_name,
			ps.plant_type,
			ps.species_name,
			pg.genus_name,
			pf.family_name,
			c.cultivar_name,
			p.created_at
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		LEFT JOIN cultivars c ON p.cultivar_id = c.cultivar_id
		WHERE LOWER(pg.genus_name) = LOWER($1) AND LOWER(ps.species_name) = LOWER($2)
		ORDER BY p.full_botanical_name
	`

	rows, err := r.db.QueryContext(ctx, query, genusName, speciesName)
	if err != nil {
		return nil, fmt.Errorf("failed to query plants by species: %w", err)
	}
	defer rows.Close()

	plants := make([]*entity.Plant, 0)
	for rows.Next() {
		plant := &entity.Plant{}
		var cultivarID, cultivarName sql.NullString

		err := rows.Scan(
			&plant.PlantID,
			&plant.SpeciesID,
			&cultivarID,
			&plant.FullBotanicalName,
			&plant.PlantType,
			&plant.SpeciesName,
			&plant.GenusName,
			&plant.FamilyName,
			&cultivarName,
			&plant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plant: %w", err)
		}

		if cultivarID.Valid {
			plant.CultivarID = &cultivarID.String
		}
		if cultivarName.Valid {
			plant.CultivarName = &cultivarName.String
		}

		plants = append(plants, plant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plants: %w", err)
	}

	// Load common names (batch loading to avoid N+1 query)
	if err := r.loadCommonNamesForMultiplePlants(ctx, plants, languageID, countryID); err != nil {
		return nil, fmt.Errorf("failed to load common names: %w", err)
	}

	return plants, nil
}

// FindByGrowingConditions finds plants matching specific growing conditions
func (r *PostgresPlantRepository) FindByGrowingConditions(ctx context.Context, filter *repository.GrowingConditionsFilter) ([]*entity.Plant, error) {
	if filter == nil {
		filter = repository.DefaultGrowingConditionsFilter()
	}

	// Build the query with dynamic WHERE conditions
	query := `
		SELECT DISTINCT
			p.plant_id,
			p.species_id,
			p.cultivar_id,
			p.full_botanical_name,
			ps.plant_type,
			ps.species_name,
			pg.genus_name,
			pf.family_name,
			p.created_at
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		INNER JOIN country_plants cp ON p.plant_id = cp.plant_id
		INNER JOIN growing_conditions_assertions gca ON cp.country_plant_id = gca.country_plant_id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	// Hardiness zone filter (use regex for exact zone match with comma boundaries)
	if filter.HardinessZone != nil {
		query += fmt.Sprintf(" AND gca.hardiness_zones ~ $%d", argCount)
		args = append(args, "(^|,)"+*filter.HardinessZone+"(,|$)")
		argCount++
	}

	// Heat zone filter (use regex for exact zone match with comma boundaries)
	if filter.HeatZone != nil {
		query += fmt.Sprintf(" AND gca.heat_zones ~ $%d", argCount)
		args = append(args, "(^|,)"+*filter.HeatZone+"(,|$)")
		argCount++
	}

	// Sun requirements filter (any match)
	if len(filter.SunRequirements) > 0 {
		sunReqs := make([]string, len(filter.SunRequirements))
		for i, sr := range filter.SunRequirements {
			sunReqs[i] = string(sr)
		}
		query += fmt.Sprintf(" AND gca.sun_requirements && $%d", argCount)
		args = append(args, pq.Array(sunReqs))
		argCount++
	}

	// Water needs filter
	if filter.WaterNeeds != nil {
		query += fmt.Sprintf(" AND gca.water_needs = $%d", argCount)
		args = append(args, string(*filter.WaterNeeds))
		argCount++
	}

	// Soil drainage filter
	if filter.SoilDrainage != nil {
		query += fmt.Sprintf(" AND gca.soil_drainage = $%d", argCount)
		args = append(args, string(*filter.SoilDrainage))
		argCount++
	}

	// Drought tolerant filter
	if filter.DroughtTolerant != nil {
		query += fmt.Sprintf(" AND gca.drought_tolerant = $%d", argCount)
		args = append(args, *filter.DroughtTolerant)
		argCount++
	}

	// Salt tolerant filter
	if filter.SaltTolerant != nil {
		query += fmt.Sprintf(" AND gca.salt_tolerant = $%d", argCount)
		args = append(args, *filter.SaltTolerant)
		argCount++
	}

	// Wind tolerant filter
	if filter.WindTolerant != nil {
		query += fmt.Sprintf(" AND gca.wind_tolerant = $%d", argCount)
		args = append(args, *filter.WindTolerant)
		argCount++
	}

	// pH range filter
	if filter.MinPH != nil || filter.MaxPH != nil {
		if filter.MinPH != nil {
			query += fmt.Sprintf(" AND (gca.ph_preference).max_ph >= $%d", argCount)
			args = append(args, *filter.MinPH)
			argCount++
		}
		if filter.MaxPH != nil {
			query += fmt.Sprintf(" AND (gca.ph_preference).min_ph <= $%d", argCount)
			args = append(args, *filter.MaxPH)
			argCount++
		}
	}

	// Flowering month filter
	if filter.FloweringMonth != nil {
		query += fmt.Sprintf(" AND $%d = ANY(gca.flowering_months)", argCount)
		args = append(args, *filter.FloweringMonth)
		argCount++
	}

	// Fruiting month filter
	if filter.FruitingMonth != nil {
		query += fmt.Sprintf(" AND $%d = ANY(gca.fruiting_months)", argCount)
		args = append(args, *filter.FruitingMonth)
		argCount++
	}

	// Confidence level filter
	if filter.MinConfidence != nil {
		query += fmt.Sprintf(" AND gca.confidence >= $%d", argCount)
		args = append(args, string(*filter.MinConfidence))
		argCount++
	}

	// Cursor-based pagination (before ORDER BY)
	if filter.Cursor != nil {
		query += fmt.Sprintf(" AND p.plant_id > $%d", argCount)
		args = append(args, *filter.Cursor)
		argCount++
	}

	// Order by plant_id for consistent pagination
	query += " ORDER BY p.plant_id"

	// Limit
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query plants by growing conditions: %w", err)
	}
	defer rows.Close()

	plants := []*entity.Plant{}
	for rows.Next() {
		plant := &entity.Plant{}
		var cultivarID sql.NullString
		var plantType string

		err := rows.Scan(
			&plant.PlantID,
			&plant.SpeciesID,
			&cultivarID,
			&plant.FullBotanicalName,
			&plantType,
			&plant.SpeciesName,
			&plant.GenusName,
			&plant.FamilyName,
			&plant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plant: %w", err)
		}

		if cultivarID.Valid {
			plant.CultivarID = &cultivarID.String
		}
		plant.PlantType = types.PlantType(plantType)

		plants = append(plants, plant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plant rows: %w", err)
	}

	return plants, nil
}
