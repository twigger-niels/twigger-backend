package database

import (
	"context"
	"database/sql"
	"fmt"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/pkg/types"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// GetCompanions retrieves companion relationships for a plant
func (r *PostgresPlantRepository) GetCompanions(ctx context.Context, plantID string, filter *entity.CompanionFilter) ([]*entity.Companion, error) {
	// Build query with filter
	query := `
		SELECT
			cr.relationship_id,
			cr.plant_a_id,
			cr.plant_b_id,
			cr.relationship_type,
			cr.benefits,
			cr.optimal_distance_m,
			cr.max_distance_m,
			cr.created_at
		FROM companion_relationships cr
		WHERE (cr.plant_a_id = $1 OR cr.plant_b_id = $1)
	`

	args := []interface{}{plantID}
	argPos := 2

	// Add filter conditions
	if filter != nil {
		if filter.RelationshipType != nil {
			query += fmt.Sprintf(" AND cr.relationship_type = $%d", argPos)
			args = append(args, string(*filter.RelationshipType))
			argPos++
		}

		if filter.BeneficialOnly {
			query += fmt.Sprintf(" AND cr.relationship_type = $%d", argPos)
			args = append(args, string(types.RelationshipBeneficial))
			argPos++
		}

		if filter.ExcludeNeutral {
			query += fmt.Sprintf(" AND cr.relationship_type != $%d", argPos)
			args = append(args, string(types.RelationshipNeutral))
			argPos++
		}
	}

	query += " ORDER BY cr.relationship_type, cr.created_at"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query companions: %w", err)
	}
	defer rows.Close()

	companions := make([]*entity.Companion, 0)
	for rows.Next() {
		companion := &entity.Companion{}
		var relationshipType string
		var benefits pq.StringArray
		var optimalDistance, maxDistance sql.NullFloat64

		err := rows.Scan(
			&companion.RelationshipID,
			&companion.PlantAID,
			&companion.PlantBID,
			&relationshipType,
			&benefits,
			&optimalDistance,
			&maxDistance,
			&companion.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan companion: %w", err)
		}

		companion.RelationshipType = types.RelationshipType(relationshipType)
		companion.Benefits = make([]string, len(benefits))
		copy(companion.Benefits, benefits)

		if optimalDistance.Valid {
			companion.OptimalDistanceM = &optimalDistance.Float64
		}
		if maxDistance.Valid {
			companion.MaxDistanceM = &maxDistance.Float64
		}

		companions = append(companions, companion)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating companions: %w", err)
	}

	// Optionally load plant details for each companion
	if len(companions) > 0 {
		if err := r.loadCompanionPlants(ctx, companions, plantID); err != nil {
			return nil, fmt.Errorf("failed to load companion plants: %w", err)
		}
	}

	return companions, nil
}

// GetCompanionsByType retrieves companions filtered by relationship type
func (r *PostgresPlantRepository) GetCompanionsByType(ctx context.Context, plantID string, relType types.RelationshipType) ([]*entity.Companion, error) {
	filter := &entity.CompanionFilter{
		PlantID:          &plantID,
		RelationshipType: &relType,
	}
	return r.GetCompanions(ctx, plantID, filter)
}

// CreateCompanionRelationship creates a new companion relationship
func (r *PostgresPlantRepository) CreateCompanionRelationship(ctx context.Context, companion *entity.Companion) error {
	if err := companion.Validate(); err != nil {
		return fmt.Errorf("invalid companion relationship: %w", err)
	}

	// Generate UUID if not provided
	if companion.RelationshipID == "" {
		companion.RelationshipID = uuid.New().String()
	}

	query := `
		INSERT INTO companion_relationships (
			relationship_id, plant_a_id, plant_b_id, relationship_type,
			benefits, optimal_distance_m, max_distance_m, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	`

	var benefits interface{}
	if len(companion.Benefits) > 0 {
		benefits = pq.Array(companion.Benefits)
	}

	var optimalDistance, maxDistance interface{}
	if companion.OptimalDistanceM != nil {
		optimalDistance = *companion.OptimalDistanceM
	}
	if companion.MaxDistanceM != nil {
		maxDistance = *companion.MaxDistanceM
	}

	_, err := r.db.ExecContext(ctx, query,
		companion.RelationshipID,
		companion.PlantAID,
		companion.PlantBID,
		string(companion.RelationshipType),
		benefits,
		optimalDistance,
		maxDistance,
	)
	if err != nil {
		return fmt.Errorf("failed to create companion relationship: %w", err)
	}

	return nil
}

// DeleteCompanionRelationship removes a companion relationship
func (r *PostgresPlantRepository) DeleteCompanionRelationship(ctx context.Context, relationshipID string) error {
	query := `DELETE FROM companion_relationships WHERE relationship_id = $1`

	result, err := r.db.ExecContext(ctx, query, relationshipID)
	if err != nil {
		return fmt.Errorf("failed to delete companion relationship: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrCompanionNotFound
	}

	return nil
}

// loadCompanionPlants loads the plant entities for companion relationships
func (r *PostgresPlantRepository) loadCompanionPlants(ctx context.Context, companions []*entity.Companion, forPlantID string) error {
	// Collect all plant IDs that are NOT the original plant
	plantIDs := make(map[string]bool)
	for _, c := range companions {
		if c.PlantAID != forPlantID {
			plantIDs[c.PlantAID] = true
		}
		if c.PlantBID != forPlantID {
			plantIDs[c.PlantBID] = true
		}
	}

	// Convert map to slice
	ids := make([]string, 0, len(plantIDs))
	for id := range plantIDs {
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil
	}

	// Load all companion plants
	plants, err := r.FindByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("failed to load companion plants: %w", err)
	}

	// Create map for quick lookup
	plantMap := make(map[string]*entity.Plant)
	for _, p := range plants {
		plantMap[p.PlantID] = p
	}

	// Assign plants to companions
	for _, c := range companions {
		if c.PlantAID != forPlantID {
			if plant, ok := plantMap[c.PlantAID]; ok {
				c.PlantA = plant
			}
		}
		if c.PlantBID != forPlantID {
			if plant, ok := plantMap[c.PlantBID]; ok {
				c.PlantB = plant
			}
		}
	}

	return nil
}
