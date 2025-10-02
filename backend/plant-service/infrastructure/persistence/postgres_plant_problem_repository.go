package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"twigger-backend/backend/plant-service/domain/entity"
	"twigger-backend/backend/plant-service/domain/repository"
)

type postgresPlantProblemRepository struct {
	db *sql.DB
}

// NewPostgresPlantProblemRepository creates a new PostgreSQL plant problem repository
func NewPostgresPlantProblemRepository(db *sql.DB) repository.PlantProblemRepository {
	return &postgresPlantProblemRepository{db: db}
}

func (r *postgresPlantProblemRepository) FindByID(ctx context.Context, problemID string) (*entity.PlantProblem, error) {
	query := `
		SELECT problem_id, plant_id, problem_type, severity, created_at, updated_at
		FROM plant_problems
		WHERE problem_id = $1
	`

	var problem entity.PlantProblem
	err := r.db.QueryRowContext(ctx, query, problemID).Scan(
		&problem.ProblemID,
		&problem.PlantID,
		&problem.ProblemType,
		&problem.Severity,
		&problem.CreatedAt,
		&problem.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plant problem not found: %s", problemID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plant problem: %w", err)
	}

	return &problem, nil
}

func (r *postgresPlantProblemRepository) FindByPlant(ctx context.Context, plantID string) ([]*entity.PlantProblem, error) {
	query := `
		SELECT problem_id, plant_id, problem_type, severity, created_at, updated_at
		FROM plant_problems
		WHERE plant_id = $1
		ORDER BY severity DESC, problem_type
	`

	rows, err := r.db.QueryContext(ctx, query, plantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query problems by plant: %w", err)
	}
	defer rows.Close()

	return r.scanProblems(rows)
}

func (r *postgresPlantProblemRepository) FindByType(ctx context.Context, plantID, problemType string) ([]*entity.PlantProblem, error) {
	query := `
		SELECT problem_id, plant_id, problem_type, severity, created_at, updated_at
		FROM plant_problems
		WHERE plant_id = $1 AND problem_type = $2
		ORDER BY severity DESC
	`

	rows, err := r.db.QueryContext(ctx, query, plantID, problemType)
	if err != nil {
		return nil, fmt.Errorf("failed to query problems by type: %w", err)
	}
	defer rows.Close()

	return r.scanProblems(rows)
}

func (r *postgresPlantProblemRepository) FindBySeverity(ctx context.Context, plantID, severity string) ([]*entity.PlantProblem, error) {
	query := `
		SELECT problem_id, plant_id, problem_type, severity, created_at, updated_at
		FROM plant_problems
		WHERE plant_id = $1 AND severity = $2
		ORDER BY problem_type
	`

	rows, err := r.db.QueryContext(ctx, query, plantID, severity)
	if err != nil {
		return nil, fmt.Errorf("failed to query problems by severity: %w", err)
	}
	defer rows.Close()

	return r.scanProblems(rows)
}

func (r *postgresPlantProblemRepository) Create(ctx context.Context, problem *entity.PlantProblem) error {
	if err := problem.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO plant_problems (problem_id, plant_id, problem_type, severity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	problem.CreatedAt = now
	problem.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		problem.ProblemID,
		problem.PlantID,
		problem.ProblemType,
		problem.Severity,
		problem.CreatedAt,
		problem.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create plant problem: %w", err)
	}

	return nil
}

func (r *postgresPlantProblemRepository) Update(ctx context.Context, problem *entity.PlantProblem) error {
	if err := problem.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE plant_problems
		SET plant_id = $2, problem_type = $3, severity = $4, updated_at = $5
		WHERE problem_id = $1
	`

	problem.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		problem.ProblemID,
		problem.PlantID,
		problem.ProblemType,
		problem.Severity,
		problem.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update plant problem: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant problem not found: %s", problem.ProblemID)
	}

	return nil
}

func (r *postgresPlantProblemRepository) Delete(ctx context.Context, problemID string) error {
	query := `DELETE FROM plant_problems WHERE problem_id = $1`

	result, err := r.db.ExecContext(ctx, query, problemID)
	if err != nil {
		return fmt.Errorf("failed to delete plant problem: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plant problem not found: %s", problemID)
	}

	return nil
}

// Helper method to scan problems
func (r *postgresPlantProblemRepository) scanProblems(rows *sql.Rows) ([]*entity.PlantProblem, error) {
	var problems []*entity.PlantProblem
	for rows.Next() {
		var problem entity.PlantProblem
		if err := rows.Scan(
			&problem.ProblemID,
			&problem.PlantID,
			&problem.ProblemType,
			&problem.Severity,
			&problem.CreatedAt,
			&problem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plant problem: %w", err)
		}
		problems = append(problems, &problem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plant problems: %w", err)
	}

	return problems, nil
}
