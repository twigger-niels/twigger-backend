package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// PreparedStatementManager manages prepared statements for frequently-called queries
type PreparedStatementManager struct {
	db         *sql.DB
	statements map[string]*sql.Stmt
	mu         sync.RWMutex
}

// NewPreparedStatementManager creates a new prepared statement manager
func NewPreparedStatementManager(db *sql.DB) *PreparedStatementManager {
	return &PreparedStatementManager{
		db:         db,
		statements: make(map[string]*sql.Stmt),
	}
}

// Statement names for commonly used queries
const (
	// Plant queries
	StmtFindPlantByID          = "find_plant_by_id"
	StmtFindPlantByBotanical   = "find_plant_by_botanical"
	StmtCountPlants            = "count_plants"

	// Country queries
	StmtFindCountryByID        = "find_country_by_id"
	StmtFindCountryByCode      = "find_country_by_code"

	// Climate zone queries
	StmtFindClimateZoneByID    = "find_climate_zone_by_id"

	// Language queries
	StmtFindLanguageByID       = "find_language_by_id"
	StmtFindLanguageByCode     = "find_language_by_code"

	// Data source queries
	StmtFindDataSourceByID     = "find_data_source_by_id"
)

// PrepareStatements prepares all frequently-used statements
func (psm *PreparedStatementManager) PrepareStatements(ctx context.Context) error {
	statements := map[string]string{
		// Plant statements
		StmtFindPlantByID: `
			SELECT plant_id, species_id, full_botanical_name, created_at, updated_at
			FROM plants
			WHERE plant_id = $1
		`,
		StmtFindPlantByBotanical: `
			SELECT plant_id, species_id, full_botanical_name, created_at, updated_at
			FROM plants
			WHERE full_botanical_name = $1
		`,
		StmtCountPlants: `
			SELECT COUNT(*) FROM plants
		`,

		// Country statements
		StmtFindCountryByID: `
			SELECT country_id, country_code, country_name, climate_systems,
			       default_climate_system, ST_AsGeoJSON(country_boundary),
			       created_at, updated_at
			FROM countries
			WHERE country_id = $1
		`,
		StmtFindCountryByCode: `
			SELECT country_id, country_code, country_name, climate_systems,
			       default_climate_system, ST_AsGeoJSON(country_boundary),
			       created_at, updated_at
			FROM countries
			WHERE country_code = $1
		`,

		// Climate zone statements
		StmtFindClimateZoneByID: `
			SELECT zone_id, country_id, zone_system, zone_code,
			       ST_AsGeoJSON(zone_geometry), min_temp_c, max_temp_c, created_at
			FROM climate_zones
			WHERE zone_id = $1
		`,

		// Language statements
		StmtFindLanguageByID: `
			SELECT language_id, language_code, language_name, native_name, is_active, created_at
			FROM languages
			WHERE language_id = $1
		`,
		StmtFindLanguageByCode: `
			SELECT language_id, language_code, language_name, native_name, is_active, created_at
			FROM languages
			WHERE language_code = $1
		`,

		// Data source statements
		StmtFindDataSourceByID: `
			SELECT source_id, source_name, source_type, website_url,
			       reliability_score, last_verified, created_at
			FROM data_sources
			WHERE source_id = $1
		`,
	}

	psm.mu.Lock()
	defer psm.mu.Unlock()

	for name, query := range statements {
		stmt, err := psm.db.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement %s: %w", name, err)
		}
		psm.statements[name] = stmt
	}

	return nil
}

// Get retrieves a prepared statement by name
func (psm *PreparedStatementManager) Get(name string) (*sql.Stmt, error) {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	stmt, ok := psm.statements[name]
	if !ok {
		return nil, fmt.Errorf("prepared statement not found: %s", name)
	}
	return stmt, nil
}

// Close closes all prepared statements
func (psm *PreparedStatementManager) Close() error {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	var errs []error
	for name, stmt := range psm.statements {
		if err := stmt.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close statement %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing prepared statements: %v", errs)
	}

	psm.statements = make(map[string]*sql.Stmt)
	return nil
}

// Recreate recreates all prepared statements (useful after connection reset)
func (psm *PreparedStatementManager) Recreate(ctx context.Context) error {
	if err := psm.Close(); err != nil {
		return fmt.Errorf("failed to close existing statements: %w", err)
	}
	return psm.PrepareStatements(ctx)
}

// Usage example in repository:
//
// type PostgresCountryRepository struct {
//     db      *sql.DB
//     stmtMgr *PreparedStatementManager
// }
//
// func (r *PostgresCountryRepository) FindByID(ctx context.Context, countryID string) (*entity.Country, error) {
//     stmt, err := r.stmtMgr.Get(StmtFindCountryByID)
//     if err != nil {
//         // Fallback to regular query
//         return r.findByIDRegular(ctx, countryID)
//     }
//
//     var country entity.Country
//     err = stmt.QueryRowContext(ctx, countryID).Scan(...)
//     return &country, err
// }
