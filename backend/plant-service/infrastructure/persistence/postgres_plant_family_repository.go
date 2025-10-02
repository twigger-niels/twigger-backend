package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

type postgresPlantFamilyRepository struct {
	db *sql.DB
}

// NewPostgresPlantFamilyRepository creates a new PostgreSQL plant family repository
func NewPostgresPlantFamilyRepository(db *sql.DB) repository.PlantFamilyRepository {
	return &postgresPlantFamilyRepository{db: db}
}

func (r *postgresPlantFamilyRepository) FindByID(ctx context.Context, familyID string) (*entity.PlantFamily, error) {
	query := `
		SELECT family_id, family_name, common_name, created_at
		FROM plant_families
		WHERE family_id = $1
	`

	var family entity.PlantFamily
	err := r.db.QueryRowContext(ctx, query, familyID).Scan(
		&family.FamilyID,
		&family.FamilyName,
		&family.CommonName,
		&family.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plant family not found: %s", familyID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant family: %w", err)
	}

	return &family, nil
}

func (r *postgresPlantFamilyRepository) FindByName(ctx context.Context, familyName string) (*entity.PlantFamily, error) {
	query := `
		SELECT family_id, family_name, common_name, created_at
		FROM plant_families
		WHERE family_name = $1
	`

	var family entity.PlantFamily
	err := r.db.QueryRowContext(ctx, query, familyName).Scan(
		&family.FamilyID,
		&family.FamilyName,
		&family.CommonName,
		&family.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plant family not found: %s", familyName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant family by name: %w", err)
	}

	return &family, nil
}

func (r *postgresPlantFamilyRepository) FindAll(ctx context.Context) ([]*entity.PlantFamily, error) {
	query := `
		SELECT family_id, family_name, common_name, created_at
		FROM plant_families
		ORDER BY family_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plant families: %w", err)
	}
	defer rows.Close()

	return r.scanFamilies(rows)
}

func (r *postgresPlantFamilyRepository) Search(ctx context.Context, query string, limit int) ([]*entity.PlantFamily, error) {
	sqlQuery := `
		SELECT family_id, family_name, common_name, created_at
		FROM plant_families
		WHERE family_name ILIKE $1 OR common_name ILIKE $1
		ORDER BY family_name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search plant families: %w", err)
	}
	defer rows.Close()

	return r.scanFamilies(rows)
}

func (r *postgresPlantFamilyRepository) Create(ctx context.Context, family *entity.PlantFamily) error {
	if err := family.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO plant_families (family_id, family_name, common_name, created_at)
		VALUES ($1, $2, $3, $4)
	`

	family.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		family.FamilyID,
		family.FamilyName,
		family.CommonName,
		family.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create plant family: %w", err)
	}

	return nil
}

func (r *postgresPlantFamilyRepository) Update(ctx context.Context, family *entity.PlantFamily) error {
	if err := family.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE plant_families
		SET family_name = $2, common_name = $3
		WHERE family_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		family.FamilyID,
		family.FamilyName,
		family.CommonName,
	)

	if err != nil {
		return fmt.Errorf("failed to update plant family: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant family not found: %s", family.FamilyID)
	}

	return nil
}

func (r *postgresPlantFamilyRepository) Delete(ctx context.Context, familyID string) error {
	query := `DELETE FROM plant_families WHERE family_id = $1`

	result, err := r.db.ExecContext(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("failed to delete plant family: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant family not found: %s", familyID)
	}

	return nil
}

// Helper method to scan families
func (r *postgresPlantFamilyRepository) scanFamilies(rows *sql.Rows) ([]*entity.PlantFamily, error) {
	var families []*entity.PlantFamily
	for rows.Next() {
		var family entity.PlantFamily
		if err := rows.Scan(
			&family.FamilyID,
			&family.FamilyName,
			&family.CommonName,
			&family.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plant family: %w", err)
		}
		families = append(families, &family)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plant families: %w", err)
	}

	return families, nil
}
