package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

type postgresPlantGenusRepository struct {
	db *sql.DB
}

// NewPostgresPlantGenusRepository creates a new PostgreSQL plant genus repository
func NewPostgresPlantGenusRepository(db *sql.DB) repository.PlantGenusRepository {
	return &postgresPlantGenusRepository{db: db}
}

func (r *postgresPlantGenusRepository) FindByID(ctx context.Context, genusID string) (*entity.PlantGenus, error) {
	query := `
		SELECT genus_id, family_id, genus_name, created_at
		FROM plant_genera
		WHERE genus_id = $1
	`

	var genus entity.PlantGenus
	err := r.db.QueryRowContext(ctx, query, genusID).Scan(
		&genus.GenusID,
		&genus.FamilyID,
		&genus.GenusName,
		&genus.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plant genus not found: %s", genusID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant genus: %w", err)
	}

	return &genus, nil
}

func (r *postgresPlantGenusRepository) FindByName(ctx context.Context, genusName string) (*entity.PlantGenus, error) {
	query := `
		SELECT genus_id, family_id, genus_name, created_at
		FROM plant_genera
		WHERE genus_name = $1
	`

	var genus entity.PlantGenus
	err := r.db.QueryRowContext(ctx, query, genusName).Scan(
		&genus.GenusID,
		&genus.FamilyID,
		&genus.GenusName,
		&genus.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plant genus not found: %s", genusName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant genus by name: %w", err)
	}

	return &genus, nil
}

func (r *postgresPlantGenusRepository) FindByFamily(ctx context.Context, familyID string) ([]*entity.PlantGenus, error) {
	query := `
		SELECT genus_id, family_id, genus_name, created_at
		FROM plant_genera
		WHERE family_id = $1
		ORDER BY genus_name
	`

	rows, err := r.db.QueryContext(ctx, query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query genera by family: %w", err)
	}
	defer rows.Close()

	return r.scanGenera(rows)
}

func (r *postgresPlantGenusRepository) FindAll(ctx context.Context) ([]*entity.PlantGenus, error) {
	query := `
		SELECT genus_id, family_id, genus_name, created_at
		FROM plant_genera
		ORDER BY genus_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plant genera: %w", err)
	}
	defer rows.Close()

	return r.scanGenera(rows)
}

func (r *postgresPlantGenusRepository) Search(ctx context.Context, query string, limit int) ([]*entity.PlantGenus, error) {
	sqlQuery := `
		SELECT genus_id, family_id, genus_name, created_at
		FROM plant_genera
		WHERE genus_name ILIKE $1
		ORDER BY genus_name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search plant genera: %w", err)
	}
	defer rows.Close()

	return r.scanGenera(rows)
}

func (r *postgresPlantGenusRepository) Create(ctx context.Context, genus *entity.PlantGenus) error {
	if err := genus.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO plant_genera (genus_id, family_id, genus_name, created_at)
		VALUES ($1, $2, $3, $4)
	`

	genus.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		genus.GenusID,
		genus.FamilyID,
		genus.GenusName,
		genus.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create plant genus: %w", err)
	}

	return nil
}

func (r *postgresPlantGenusRepository) Update(ctx context.Context, genus *entity.PlantGenus) error {
	if err := genus.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE plant_genera
		SET family_id = $2, genus_name = $3
		WHERE genus_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		genus.GenusID,
		genus.FamilyID,
		genus.GenusName,
	)

	if err != nil {
		return fmt.Errorf("failed to update plant genus: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant genus not found: %s", genus.GenusID)
	}

	return nil
}

func (r *postgresPlantGenusRepository) Delete(ctx context.Context, genusID string) error {
	query := `DELETE FROM plant_genera WHERE genus_id = $1`

	result, err := r.db.ExecContext(ctx, query, genusID)
	if err != nil {
		return fmt.Errorf("failed to delete plant genus: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant genus not found: %s", genusID)
	}

	return nil
}

// Helper method to scan genera
func (r *postgresPlantGenusRepository) scanGenera(rows *sql.Rows) ([]*entity.PlantGenus, error) {
	var genera []*entity.PlantGenus
	for rows.Next() {
		var genus entity.PlantGenus
		if err := rows.Scan(
			&genus.GenusID,
			&genus.FamilyID,
			&genus.GenusName,
			&genus.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plant genus: %w", err)
		}
		genera = append(genera, &genus)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plant genera: %w", err)
	}

	return genera, nil
}
