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
		Port:     getEnv("TEST_DB_PORT", "5432"),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnv("TEST_DB_NAME", "twigger"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}
}

// GetTestDBConfig returns test database configuration
func GetTestDBConfig() *TestDBConfig {
	// Check if full DATABASE_URL is provided
	if dbURL := os.Getenv("TEST_DATABASE_URL"); dbURL != "" {
		// For now, use defaults and let connection string override
		return DefaultTestDBConfig()
	}
	return DefaultTestDBConfig()
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

	// Clean and run migrations
	if err := cleanDatabase(ctx, db, t); err != nil {
		db.Close()
		t.Fatalf("failed to clean database: %v", err)
	}

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

	// Close connection
	if err := db.Close(); err != nil {
		t.Errorf("failed to close database connection: %v", err)
	}
}

// cleanDatabase drops all schema objects to ensure clean state
func cleanDatabase(ctx context.Context, db *sql.DB, t *testing.T) error {
	t.Helper()

	// Safety check: Only run on test databases
	var dbName string
	err := db.QueryRowContext(ctx, "SELECT current_database()").Scan(&dbName)
	if err != nil {
		return fmt.Errorf("failed to get database name: %w", err)
	}

	// Allow both 'twigger' and 'plantdb_test' for flexibility
	if dbName != "twigger" && dbName != "plantdb_test" {
		return fmt.Errorf("refusing to clean non-test database: %s", dbName)
	}

	t.Log("Cleaning database schema...")

	// Drop all tables with CASCADE to remove dependencies
	_, err = db.ExecContext(ctx, `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO postgres;
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

	// Enable uuid-ossp extension for UUID generation
	_, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		return fmt.Errorf("failed to enable uuid-ossp: %w", err)
	}

	t.Log("Database cleaned successfully")
	return nil
}

// runMigrations executes SQL migration files
func runMigrations(ctx context.Context, db *sql.DB, t *testing.T) error {
	t.Helper()

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

	// Filter for auth-specific migrations only
	// For auth service tests, we only need migration 008 (auth tables)
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			// Only include migration 008 for auth service
			if strings.Contains(entry.Name(), "008_add_auth_and_workspaces") {
				migrationFiles = append(migrationFiles, entry.Name())
			}
		}
	}
	sort.Strings(migrationFiles)

	if len(migrationFiles) == 0 {
		return fmt.Errorf("no auth migration files found (looking for 008_add_auth_and_workspaces)")
	}

	t.Logf("Running %d auth migration files from %s", len(migrationFiles), migrationsDir)

	// Create minimal base schema first (needed by migration 008)
	if err := createMinimalAuthSchema(ctx, db, t); err != nil {
		return fmt.Errorf("failed to create minimal schema: %w", err)
	}

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

	// Create audit log partition for current month
	if err := createCurrentMonthPartition(ctx, db, t); err != nil {
		return fmt.Errorf("failed to create audit partition: %w", err)
	}

	t.Logf("Migrations completed successfully")
	return nil
}

// createMinimalAuthSchema creates minimal tables needed for auth migration 008
func createMinimalAuthSchema(ctx context.Context, db *sql.DB, t *testing.T) error {
	t.Helper()
	t.Log("Creating minimal base schema for auth tests...")

	schema := `
		-- Minimal users table (migration 008 will extend this)
		CREATE TABLE users (
			user_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(100) UNIQUE NOT NULL,
			location GEOGRAPHY(POINT, 4326),
			detected_hardiness_zone VARCHAR(10),
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		-- Minimal gardens table (referenced by migration 008)
		CREATE TABLE gardens (
			garden_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			garden_name VARCHAR(200) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create minimal schema: %w", err)
	}

	t.Log("Minimal schema created successfully")
	return nil
}

// createCurrentMonthPartition creates a partition for the current month's audit log
func createCurrentMonthPartition(ctx context.Context, db *sql.DB, t *testing.T) error {
	t.Helper()
	t.Log("Creating audit log partition for current month...")

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Calculate partition bounds
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0) // First day of next month

	partitionName := fmt.Sprintf("auth_audit_log_y%dm%02d", year, month)

	createPartitionSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF auth_audit_log
		FOR VALUES FROM ('%s') TO ('%s')
	`, partitionName, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if _, err := db.ExecContext(ctx, createPartitionSQL); err != nil {
		return fmt.Errorf("failed to create partition %s: %w", partitionName, err)
	}

	t.Logf("Created partition: %s", partitionName)
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

// SeedTestUsers inserts test user data
func SeedTestUsers(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()

	ctx := context.Background()

	users := []struct {
		UserID      string
		FirebaseUID string
		Email       string
		Username    string
	}{
		{
			"550e8400-e29b-41d4-a716-446655440001",
			"firebase-uid-test-user-1",
			"test1@example.com",
			"testuser1",
		},
		{
			"550e8400-e29b-41d4-a716-446655440002",
			"firebase-uid-test-user-2",
			"test2@example.com",
			"testuser2",
		},
		{
			"550e8400-e29b-41d4-a716-446655440003",
			"firebase-uid-test-user-3",
			"test3@example.com",
			"testuser3",
		},
	}

	userMap := make(map[string]string)

	for _, user := range users {
		query := `
			INSERT INTO users (
				user_id, firebase_uid, email, username,
				email_verified, provider, created_at
			) VALUES ($1, $2, $3, $4, true, 'google.com', NOW())
			ON CONFLICT (user_id) DO NOTHING
		`
		_, err := db.ExecContext(ctx, query,
			user.UserID, user.FirebaseUID, user.Email, user.Username)
		if err != nil {
			t.Fatalf("failed to seed user %s: %v", user.Email, err)
		}
		userMap[user.Username] = user.UserID
	}

	t.Logf("Seeded %d test users", len(users))
	return userMap
}

// SeedTestWorkspaces inserts test workspace data
func SeedTestWorkspaces(t *testing.T, db *sql.DB, userIDs map[string]string) map[string]string {
	t.Helper()

	ctx := context.Background()

	workspaces := []struct {
		WorkspaceID string
		OwnerKey    string // Key in userIDs map
		Name        string
	}{
		{
			"650e8400-e29b-41d4-a716-446655440001",
			"testuser1",
			"Test User 1's Garden",
		},
		{
			"650e8400-e29b-41d4-a716-446655440002",
			"testuser2",
			"Test User 2's Garden",
		},
	}

	workspaceMap := make(map[string]string)

	for _, ws := range workspaces {
		ownerID := userIDs[ws.OwnerKey]
		if ownerID == "" {
			t.Fatalf("owner not found for key: %s", ws.OwnerKey)
		}

		query := `
			INSERT INTO workspaces (workspace_id, owner_id, name, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT (workspace_id) DO NOTHING
		`
		_, err := db.ExecContext(ctx, query, ws.WorkspaceID, ownerID, ws.Name)
		if err != nil {
			t.Fatalf("failed to seed workspace %s: %v", ws.Name, err)
		}

		// Add owner as admin member
		memberQuery := `
			INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
			VALUES ($1, $2, 'admin', NOW())
			ON CONFLICT (workspace_id, user_id) DO NOTHING
		`
		_, err = db.ExecContext(ctx, memberQuery, ws.WorkspaceID, ownerID)
		if err != nil {
			t.Fatalf("failed to add workspace member for %s: %v", ws.Name, err)
		}

		workspaceMap[ws.OwnerKey] = ws.WorkspaceID
	}

	t.Logf("Seeded %d test workspaces", len(workspaces))
	return workspaceMap
}

// CleanupTestData removes all test data from tables
func CleanupTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := context.Background()

	tables := []string{
		"auth_audit_log",         // Partitioned table - might need special handling
		"auth_sessions",
		"workspace_members",
		"workspaces",
		"users",
	}

	for _, table := range tables {
		// Use DELETE for partitioned tables, TRUNCATE for regular tables
		var query string
		if table == "auth_audit_log" {
			query = fmt.Sprintf("DELETE FROM %s", table)
		} else {
			query = fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		}

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			t.Logf("Warning: failed to clean %s: %v", table, err)
		}
	}
}
