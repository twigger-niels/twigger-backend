package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

// Search performs full-text search on plants with filtering and pagination
func (r *PostgresPlantRepository) Search(ctx context.Context, query string, filter *repository.SearchFilter) (*repository.SearchResult, error) {
	if filter == nil {
		filter = repository.DefaultSearchFilter()
	}

	// Build the WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argPos := 1

	// Full-text search on botanical name
	if query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"to_tsvector('english', COALESCE(p.full_botanical_name, '')) @@ plainto_tsquery('english', $%d)",
			argPos,
		))
		args = append(args, query)
		argPos++
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

	// Build WHERE clause
	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
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
		if query != "" {
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

	// Count total results
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM plants p
		INNER JOIN plant_species ps ON p.species_id = ps.species_id
		INNER JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		INNER JOIN plant_families pf ON pg.family_id = pf.family_id
		%s
	`, whereSQL)

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Build main query with pagination
	searchQuery := fmt.Sprintf(`
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
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereSQL, orderBySQL, argPos, argPos+1)

	args = append(args, filter.Limit, filter.Offset)

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

	// Load common names for all plants
	for _, plant := range plants {
		if err := r.loadCommonNames(ctx, plant); err != nil {
			return nil, fmt.Errorf("failed to load common names: %w", err)
		}
	}

	return &repository.SearchResult{
		Plants:  plants,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: int64(filter.Offset+filter.Limit) < total,
		Query:   query,
	}, nil
}

// FindByFamily finds all plants in a family
func (r *PostgresPlantRepository) FindByFamily(ctx context.Context, familyName string, limit, offset int) ([]*entity.Plant, error) {
	filter := repository.DefaultSearchFilter()
	filter.FamilyName = &familyName
	filter.Limit = limit
	filter.Offset = offset
	filter.SortBy = repository.SortByGenusName

	result, err := r.Search(ctx, "", filter)
	if err != nil {
		return nil, err
	}

	return result.Plants, nil
}

// FindByGenus finds all plants in a genus
func (r *PostgresPlantRepository) FindByGenus(ctx context.Context, genusName string, limit, offset int) ([]*entity.Plant, error) {
	filter := repository.DefaultSearchFilter()
	filter.GenusName = &genusName
	filter.Limit = limit
	filter.Offset = offset
	filter.SortBy = repository.SortByBotanicalName

	result, err := r.Search(ctx, "", filter)
	if err != nil {
		return nil, err
	}

	return result.Plants, nil
}

// FindBySpecies finds plants by genus and species name
func (r *PostgresPlantRepository) FindBySpecies(ctx context.Context, genusName, speciesName string) ([]*entity.Plant, error) {
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

	// Load common names
	for _, plant := range plants {
		if err := r.loadCommonNames(ctx, plant); err != nil {
			return nil, fmt.Errorf("failed to load common names: %w", err)
		}
	}

	return plants, nil
}

// FindByGrowingConditions finds plants matching specific growing conditions
func (r *PostgresPlantRepository) FindByGrowingConditions(ctx context.Context, filter *repository.GrowingConditionsFilter) ([]*entity.Plant, error) {
	if filter == nil {
		filter = repository.DefaultGrowingConditionsFilter()
	}

	// This would join with growing_conditions_assertions and country_plants tables
	// For now, return empty results as it requires complex filtering
	// TODO: Implement after testing basic functionality
	return []*entity.Plant{}, nil
}
