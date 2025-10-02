package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

type postgresCultivarRepository struct {
	db *sql.DB
}

// NewPostgresCultivarRepository creates a new PostgreSQL cultivar repository
func NewPostgresCultivarRepository(db *sql.DB) repository.CultivarRepository {
	return &postgresCultivarRepository{db: db}
}

func (r *postgresCultivarRepository) FindByID(ctx context.Context, cultivarID string) (*entity.Cultivar, error) {
	query := `
		SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
		       patent_expiry, propagation_restricted, created_at
		FROM cultivars
		WHERE cultivar_id = $1
	`

	var cultivar entity.Cultivar
	err := r.db.QueryRowContext(ctx, query, cultivarID).Scan(
		&cultivar.CultivarID,
		&cultivar.SpeciesID,
		&cultivar.CultivarName,
		&cultivar.TradeName,
		&cultivar.PatentNumber,
		&cultivar.PatentExpiry,
		&cultivar.PropagationRestricted,
		&cultivar.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cultivar not found: %s", cultivarID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find cultivar: %w", err)
	}

	return &cultivar, nil
}

func (r *postgresCultivarRepository) FindBySpecies(ctx context.Context, speciesID string) ([]*entity.Cultivar, error) {
	query := `
		SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
		       patent_expiry, propagation_restricted, created_at
		FROM cultivars
		WHERE species_id = $1
		ORDER BY cultivar_name
	`

	rows, err := r.db.QueryContext(ctx, query, speciesID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cultivars by species: %w", err)
	}
	defer rows.Close()

	return r.scanCultivars(rows)
}

func (r *postgresCultivarRepository) FindByPatent(ctx context.Context, patentNumber string) (*entity.Cultivar, error) {
	query := `
		SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
		       patent_expiry, propagation_restricted, created_at
		FROM cultivars
		WHERE patent_number = $1
	`

	var cultivar entity.Cultivar
	err := r.db.QueryRowContext(ctx, query, patentNumber).Scan(
		&cultivar.CultivarID,
		&cultivar.SpeciesID,
		&cultivar.CultivarName,
		&cultivar.TradeName,
		&cultivar.PatentNumber,
		&cultivar.PatentExpiry,
		&cultivar.PropagationRestricted,
		&cultivar.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cultivar not found with patent: %s", patentNumber)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find cultivar by patent: %w", err)
	}

	return &cultivar, nil
}

func (r *postgresCultivarRepository) FindByTradeName(ctx context.Context, tradeName string) ([]*entity.Cultivar, error) {
	query := `
		SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
		       patent_expiry, propagation_restricted, created_at
		FROM cultivars
		WHERE trade_name = $1
		ORDER BY cultivar_name
	`

	rows, err := r.db.QueryContext(ctx, query, tradeName)
	if err != nil {
		return nil, fmt.Errorf("failed to query cultivars by trade name: %w", err)
	}
	defer rows.Close()

	return r.scanCultivars(rows)
}

func (r *postgresCultivarRepository) FindRestricted(ctx context.Context) ([]*entity.Cultivar, error) {
	query := `
		SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
		       patent_expiry, propagation_restricted, created_at
		FROM cultivars
		WHERE propagation_restricted = true
		ORDER BY cultivar_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query restricted cultivars: %w", err)
	}
	defer rows.Close()

	return r.scanCultivars(rows)
}

func (r *postgresCultivarRepository) Search(ctx context.Context, query string, limit int) ([]*entity.Cultivar, error) {
	sqlQuery := `
		SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
		       patent_expiry, propagation_restricted, created_at
		FROM cultivars
		WHERE cultivar_name ILIKE $1 OR trade_name ILIKE $1
		ORDER BY cultivar_name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search cultivars: %w", err)
	}
	defer rows.Close()

	return r.scanCultivars(rows)
}

func (r *postgresCultivarRepository) Create(ctx context.Context, cultivar *entity.Cultivar) error {
	if err := cultivar.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO cultivars (cultivar_id, species_id, cultivar_name, trade_name,
		                       patent_number, patent_expiry, propagation_restricted, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	cultivar.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		cultivar.CultivarID,
		cultivar.SpeciesID,
		cultivar.CultivarName,
		cultivar.TradeName,
		cultivar.PatentNumber,
		cultivar.PatentExpiry,
		cultivar.PropagationRestricted,
		cultivar.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create cultivar: %w", err)
	}

	return nil
}

func (r *postgresCultivarRepository) Update(ctx context.Context, cultivar *entity.Cultivar) error {
	if err := cultivar.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE cultivars
		SET species_id = $2, cultivar_name = $3, trade_name = $4,
		    patent_number = $5, patent_expiry = $6, propagation_restricted = $7
		WHERE cultivar_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		cultivar.CultivarID,
		cultivar.SpeciesID,
		cultivar.CultivarName,
		cultivar.TradeName,
		cultivar.PatentNumber,
		cultivar.PatentExpiry,
		cultivar.PropagationRestricted,
	)

	if err != nil {
		return fmt.Errorf("failed to update cultivar: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("cultivar not found: %s", cultivar.CultivarID)
	}

	return nil
}

func (r *postgresCultivarRepository) Delete(ctx context.Context, cultivarID string) error {
	query := `DELETE FROM cultivars WHERE cultivar_id = $1`

	result, err := r.db.ExecContext(ctx, query, cultivarID)
	if err != nil {
		return fmt.Errorf("failed to delete cultivar: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("cultivar not found: %s", cultivarID)
	}

	return nil
}

// Helper method to scan cultivars
func (r *postgresCultivarRepository) scanCultivars(rows *sql.Rows) ([]*entity.Cultivar, error) {
	var cultivars []*entity.Cultivar
	for rows.Next() {
		var cultivar entity.Cultivar
		if err := rows.Scan(
			&cultivar.CultivarID,
			&cultivar.SpeciesID,
			&cultivar.CultivarName,
			&cultivar.TradeName,
			&cultivar.PatentNumber,
			&cultivar.PatentExpiry,
			&cultivar.PropagationRestricted,
			&cultivar.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cultivar: %w", err)
		}
		cultivars = append(cultivars, &cultivar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cultivars: %w", err)
	}

	return cultivars, nil
}
