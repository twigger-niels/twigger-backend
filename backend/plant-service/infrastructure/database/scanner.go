package database

import (
	"database/sql"
	"fmt"
)

// Scanner is a generic interface for types that can be scanned from database rows
type Scanner[T any] interface {
	Scan(row *sql.Rows) (*T, error)
}

// ScanRows is a generic utility to scan multiple rows into a slice
// This reduces code duplication across repository scan methods
func ScanRows[T any](rows *sql.Rows, scanFunc func(*sql.Rows) (*T, error)) ([]*T, error) {
	var results []*T

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// ScanRow is a generic utility to scan a single row
func ScanRow[T any](row *sql.Row, scanFunc func(*sql.Row) (*T, error)) (*T, error) {
	return scanFunc(row)
}

// RowScanner is a generic row scanner that uses a scan function
type RowScanner[T any] struct {
	scanFunc func(*sql.Rows) (*T, error)
}

// NewRowScanner creates a new generic row scanner
func NewRowScanner[T any](scanFunc func(*sql.Rows) (*T, error)) *RowScanner[T] {
	return &RowScanner[T]{scanFunc: scanFunc}
}

// ScanAll scans all rows into a slice
func (s *RowScanner[T]) ScanAll(rows *sql.Rows) ([]*T, error) {
	return ScanRows(rows, s.scanFunc)
}

// ScanOne scans a single row
func (s *RowScanner[T]) ScanOne(rows *sql.Rows) (*T, error) {
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	return s.scanFunc(rows)
}

// --- Example usage patterns ---

// Example 1: Using ScanRows directly
// func (r *PostgresLanguageRepository) scanLanguages(rows *sql.Rows) ([]*entity.Language, error) {
//     return ScanRows(rows, func(row *sql.Rows) (*entity.Language, error) {
//         var lang entity.Language
//         err := row.Scan(&lang.LanguageID, &lang.LanguageCode, &lang.LanguageName, ...)
//         return &lang, err
//     })
// }

// Example 2: Using RowScanner with reusable scan function
// type PostgresLanguageRepository struct {
//     db      *sql.DB
//     scanner *RowScanner[entity.Language]
// }
//
// func NewPostgresLanguageRepository(db *sql.DB) *PostgresLanguageRepository {
//     scanner := NewRowScanner(func(row *sql.Rows) (*entity.Language, error) {
//         var lang entity.Language
//         err := row.Scan(&lang.LanguageID, &lang.LanguageCode, ...)
//         return &lang, err
//     })
//     return &PostgresLanguageRepository{db: db, scanner: scanner}
// }
//
// func (r *PostgresLanguageRepository) FindAll(ctx context.Context) ([]*entity.Language, error) {
//     rows, err := r.db.QueryContext(ctx, "SELECT ...")
//     if err != nil {
//         return nil, err
//     }
//     defer rows.Close()
//
//     return r.scanner.ScanAll(rows)
// }

// --- Advanced generic scanning utilities ---

// ScanIntoMap scans rows into a map keyed by a specified field
func ScanIntoMap[K comparable, V any](
	rows *sql.Rows,
	scanFunc func(*sql.Rows) (*V, error),
	keyFunc func(*V) K,
) (map[K]*V, error) {
	result := make(map[K]*V)

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		key := keyFunc(item)
		result[key] = item
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// ScanWithTransform scans rows and applies a transform function to each result
func ScanWithTransform[T any, R any](
	rows *sql.Rows,
	scanFunc func(*sql.Rows) (*T, error),
	transformFunc func(*T) (*R, error),
) ([]*R, error) {
	var results []*R

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		transformed, err := transformFunc(item)
		if err != nil {
			return nil, fmt.Errorf("failed to transform item: %w", err)
		}

		results = append(results, transformed)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// ScanWithFilter scans rows and filters results based on a predicate
func ScanWithFilter[T any](
	rows *sql.Rows,
	scanFunc func(*sql.Rows) (*T, error),
	filterFunc func(*T) bool,
) ([]*T, error) {
	var results []*T

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if filterFunc(item) {
			results = append(results, item)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GroupBy groups scanned rows by a key function
func GroupBy[K comparable, V any](
	rows *sql.Rows,
	scanFunc func(*sql.Rows) (*V, error),
	keyFunc func(*V) K,
) (map[K][]*V, error) {
	result := make(map[K][]*V)

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		key := keyFunc(item)
		result[key] = append(result[key], item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// Example: Group plants by species
// grouped, err := GroupBy(
//     rows,
//     scanPlant,
//     func(p *entity.Plant) string { return p.SpeciesID },
// )
