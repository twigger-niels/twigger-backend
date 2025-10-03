// +build integration

package testing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultTestDBConfig returns the default test database configuration
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5433"),
		User:     getEnv("TEST_DB_USER", "plant_api_test"),
		Password: getEnv("TEST_DB_PASSWORD", "test_password_123"),
		DBName:   getEnv("TEST_DB_NAME", "plantdb_test"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}
}

// SetupTestDB creates a test database connection and runs migrations
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	config := GetTestDBConfig()

	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Verify PostGIS extension
	var postgisVersion string
	err = db.QueryRowContext(ctx, "SELECT PostGIS_version()").Scan(&postgisVersion)
	if err != nil {
		db.Close()
		t.Fatalf("PostGIS extension not available: %v", err)
	}
	t.Logf("PostGIS version: %s", postgisVersion)

	// Run migrations
	if err := runMigrations(ctx, db, t); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// TeardownTestDB closes the database connection and cleans up test data
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if db == nil {
		return
	}

	// Clean up test data (truncate tables in reverse dependency order)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tables := []string{
		// Plant-related localization tables
		"plant_common_names",
		"plant_descriptions",
		"plant_problems_i18n",
		"companion_benefits_i18n",
		"physical_traits_i18n",
		"growing_conditions_i18n",

		// Relationship tables
		"companion_plants",
		"plant_problems",
		"country_plants",
		"plant_synonyms",

		// Plant hierarchy tables
		"plants",
		"cultivars",
		"plant_species",
		"plant_genera",
		"plant_families",

		// Reference tables (don't truncate - preserve test data)
		// "languages", "countries", "climate_zones", "data_sources"
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			// Log error but don't fail - table might not exist yet
			t.Logf("Warning: failed to truncate %s: %v", table, err)
		}
	}

	// Close connection
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database connection: %v", err)
	}
}

// cleanDatabase drops all schema objects to ensure clean state
func cleanDatabase(ctx context.Context, db *sql.DB, t *testing.T) error {
	t.Helper()

	t.Log("Cleaning database schema...")

	// Drop all tables with CASCADE to remove dependencies
	_, err := db.ExecContext(ctx, `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO plant_api_test;
		GRANT ALL ON SCHEMA public TO public;
	`)
	if err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Re-enable PostGIS extension
	_, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS postgis;`)
	if err != nil {
		return fmt.Errorf("failed to enable PostGIS: %w", err)
	}

	t.Log("Database cleaned successfully")
	return nil
}

// runMigrations executes SQL migration files
func runMigrations(ctx context.Context, db *sql.DB, t *testing.T) error {
	t.Helper()

	// Clean database first to ensure fresh state
	if err := cleanDatabase(ctx, db, t); err != nil {
		return err
	}

	// Find migrations directory
	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return fmt.Errorf("failed to find migrations directory: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort .up.sql files
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	t.Logf("Running %d migration files", len(migrationFiles))

	// Execute each migration
	for _, filename := range migrationFiles {
		filePath := filepath.Join(migrationsDir, filename)

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		t.Logf("Running migration: %s", filename)

		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
	}

	// Create minimal localization table for testing
	t.Logf("Creating plant_common_names table for testing")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS plant_common_names (
			plant_id UUID NOT NULL REFERENCES plants(plant_id) ON DELETE CASCADE,
			language_id UUID NOT NULL REFERENCES languages(language_id),
			country_id UUID REFERENCES countries(country_id),
			common_name VARCHAR(200) NOT NULL,
			is_primary BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (plant_id, language_id, common_name)
		);

		CREATE INDEX IF NOT EXISTS idx_plant_common_names_lookup
		ON plant_common_names(plant_id, language_id, country_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create plant_common_names table: %w", err)
	}

	t.Logf("Migrations completed successfully")
	return nil
}

// findMigrationsDir locates the migrations directory
func findMigrationsDir() (string, error) {
	// Try multiple potential paths
	possiblePaths := []string{
		"../../../../migrations",
		"../../../migrations",
		"../../migrations",
		"migrations",
		"./migrations",
	}

	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found in any expected location")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SeedTestLanguages inserts test language data
func SeedTestLanguages(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := context.Background()

	languages := []struct {
		LanguageID string
		Code       string
		Name       string
	}{
		{"550e8400-e29b-41d4-a716-446655440001", "en", "English"},
		{"550e8400-e29b-41d4-a716-446655440002", "es", "Spanish"},
		{"550e8400-e29b-41d4-a716-446655440003", "fr", "French"},
		{"550e8400-e29b-41d4-a716-446655440004", "de", "German"},
	}

	for _, lang := range languages {
		query := `
			INSERT INTO languages (language_id, language_code, language_name, is_active)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (language_id) DO NOTHING
		`
		_, err := db.ExecContext(ctx, query, lang.LanguageID, lang.Code, lang.Name)
		if err != nil {
			t.Fatalf("failed to seed language %s: %v", lang.Code, err)
		}
	}

	t.Logf("Seeded %d test languages", len(languages))
}

// SeedTestCountries inserts test country data
func SeedTestCountries(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := context.Background()

	countries := []struct {
		CountryID string
		Code      string
		Name      string
	}{
		{"650e8400-e29b-41d4-a716-446655440001", "US", "United States"},
		{"650e8400-e29b-41d4-a716-446655440002", "MX", "Mexico"},
		{"650e8400-e29b-41d4-a716-446655440003", "GB", "United Kingdom"},
		{"650e8400-e29b-41d4-a716-446655440004", "DE", "Germany"},
	}

	for _, country := range countries {
		query := `
			INSERT INTO countries (country_id, country_code, country_name, climate_systems)
			VALUES ($1, $2, $3, ARRAY['temperate']::text[])
			ON CONFLICT (country_id) DO NOTHING
		`
		_, err := db.ExecContext(ctx, query, country.CountryID, country.Code, country.Name)
		if err != nil {
			t.Fatalf("failed to seed country %s: %v", country.Code, err)
		}
	}

	t.Logf("Seeded %d test countries", len(countries))
}

// SeedTestPlantHierarchy creates test data for plant taxonomy
func SeedTestPlantHierarchy(t *testing.T, db *sql.DB) (familyID, genusID, speciesID string) {
	t.Helper()

	ctx := context.Background()

	// Insert family
	familyID = "f50e8400-e29b-41d4-a716-446655440001"
	_, err := db.ExecContext(ctx, `
		INSERT INTO plant_families (family_id, family_name)
		VALUES ($1, 'Rosaceae')
		ON CONFLICT (family_id) DO NOTHING
	`, familyID)
	if err != nil {
		t.Fatalf("failed to seed family: %v", err)
	}

	// Insert genus
	genusID = "150e8400-e29b-41d4-a716-446655440001"
	_, err = db.ExecContext(ctx, `
		INSERT INTO plant_genera (genus_id, family_id, genus_name)
		VALUES ($1, $2, 'Rosa')
		ON CONFLICT (genus_id) DO NOTHING
	`, genusID, familyID)
	if err != nil {
		t.Fatalf("failed to seed genus: %v", err)
	}

	// Insert species
	speciesID = "250e8400-e29b-41d4-a716-446655440001"
	_, err = db.ExecContext(ctx, `
		INSERT INTO plant_species (species_id, genus_id, species_name, plant_type)
		VALUES ($1, $2, 'rugosa', 'shrub')
		ON CONFLICT (species_id) DO NOTHING
	`, speciesID, genusID)
	if err != nil {
		t.Fatalf("failed to seed species: %v", err)
	}

	t.Logf("Seeded test plant hierarchy (family: %s, genus: %s, species: %s)", familyID, genusID, speciesID)

	return familyID, genusID, speciesID
}

// CleanupTestData removes all test data from tables
func CleanupTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := context.Background()

	tables := []string{
		"plant_common_names",
		"plants",
		"plant_species",
		"plant_genera",
		"plant_families",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			t.Logf("Warning: failed to clean %s: %v", table, err)
		}
	}
}
