package testing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// GetTestDB returns a connection to the test database
func GetTestDB(t *testing.T) *sql.DB {
	// Use environment variable or default test database URL
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/plantdb_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// CleanDatabase drops and recreates the public schema (Gotcha #17)
// This ensures a clean state for each test
func CleanDatabase(ctx context.Context, db *sql.DB, t *testing.T) error {
	// Safety check: Only run on test databases
	var dbName string
	err := db.QueryRowContext(ctx, "SELECT current_database()").Scan(&dbName)
	if err != nil {
		return fmt.Errorf("failed to get database name: %w", err)
	}

	if dbName != "plantdb_test" {
		return fmt.Errorf("refusing to clean non-test database: %s", dbName)
	}

	// Drop entire schema with CASCADE to remove all dependencies
	_, err = db.ExecContext(ctx, `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO postgres;
		GRANT ALL ON SCHEMA public TO public;
	`)
	if err != nil {
		return fmt.Errorf("failed to clean schema: %w", err)
	}

	// Re-enable PostGIS extension (lost in schema drop)
	_, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS postgis;`)
	if err != nil {
		return fmt.Errorf("failed to recreate postgis extension: %w", err)
	}

	_, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		return fmt.Errorf("failed to recreate uuid-ossp extension: %w", err)
	}

	return nil
}

// CreateTestSchema creates the minimal schema needed for garden service tests
func CreateTestSchema(ctx context.Context, db *sql.DB, t *testing.T) error {
	schema := `
		-- Users table (simplified for testing)
		CREATE TABLE users (
			user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		-- Gardens table
		CREATE TABLE gardens (
			garden_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			garden_name VARCHAR(200) NOT NULL,
			boundary GEOMETRY(Polygon, 4326),
			location GEOGRAPHY(Point, 4326),
			elevation_m NUMERIC(5,2),
			slope_degrees NUMERIC(3,1) CHECK (slope_degrees >= 0 AND slope_degrees <= 90),
			aspect VARCHAR(10) CHECK (aspect IN ('N','NE','E','SE','S','SW','W','NW','flat')),
			hardiness_zone VARCHAR(10),
			garden_type VARCHAR(20) CHECK (garden_type IN ('ornamental', 'vegetable', 'mixed', 'orchard', 'greenhouse')),
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_gardens_boundary ON gardens USING GIST(boundary);
		CREATE INDEX idx_gardens_location ON gardens USING GIST(location);
		CREATE INDEX idx_gardens_user ON gardens(user_id);

		-- Garden zones table
		CREATE TABLE garden_zones (
			zone_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			garden_id UUID NOT NULL REFERENCES gardens(garden_id) ON DELETE CASCADE,
			zone_name VARCHAR(100),
			zone_type VARCHAR(20) CHECK (zone_type IN ('bed', 'border', 'lawn', 'path', 'water', 'structure', 'compost')),
			geometry GEOMETRY(Polygon, 4326) NOT NULL,
			area_m2 NUMERIC(10,2),
			soil_type TEXT,
			soil_amended BOOLEAN DEFAULT FALSE,
			irrigation_type VARCHAR(20) CHECK (irrigation_type IN ('none', 'drip', 'sprinkler', 'soaker', 'manual')),
			sun_hours_summer INTEGER CHECK (sun_hours_summer >= 0 AND sun_hours_summer <= 24),
			sun_hours_winter INTEGER CHECK (sun_hours_winter >= 0 AND sun_hours_winter <= 24),
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_garden_zones_geometry ON garden_zones USING GIST(geometry);
		CREATE INDEX idx_garden_zones_garden ON garden_zones(garden_id);

		-- Garden features table
		CREATE TABLE garden_features (
			feature_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			garden_id UUID NOT NULL REFERENCES gardens(garden_id) ON DELETE CASCADE,
			feature_type VARCHAR(20) CHECK (feature_type IN ('tree', 'shrub', 'building', 'fence', 'wall', 'greenhouse', 'shed', 'pond', 'path')),
			feature_name VARCHAR(200),
			geometry GEOMETRY(Geometry, 4326) NOT NULL,
			height_m NUMERIC(5,2),
			canopy_diameter_m NUMERIC(5,2),
			deciduous BOOLEAN,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_garden_features_geometry ON garden_features USING GIST(geometry);
		CREATE INDEX idx_garden_features_garden ON garden_features(garden_id);

		-- Plants table (minimal for FK reference)
		CREATE TABLE plants (
			plant_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			full_botanical_name TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		-- Garden plants table
		CREATE TABLE garden_plants (
			garden_plant_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			garden_id UUID NOT NULL REFERENCES gardens(garden_id) ON DELETE CASCADE,
			zone_id UUID REFERENCES garden_zones(zone_id) ON DELETE SET NULL,
			plant_id UUID NOT NULL REFERENCES plants(plant_id) ON DELETE CASCADE,
			location GEOMETRY(Point, 4326) NOT NULL,
			planted_date DATE,
			removed_date DATE,
			quantity INTEGER DEFAULT 1,
			plant_source TEXT,
			health_status VARCHAR(20) CHECK (health_status IN ('thriving', 'healthy', 'struggling', 'diseased', 'dead')),
			notes TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_garden_plants_location ON garden_plants USING GIST(location);
		CREATE INDEX idx_garden_plants_garden ON garden_plants(garden_id);
		CREATE INDEX idx_garden_plants_zone ON garden_plants(zone_id);
		CREATE INDEX idx_garden_plants_plant ON garden_plants(plant_id);

		-- Climate zones table (for hardiness zone detection)
		CREATE TABLE climate_zones (
			zone_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			zone_system VARCHAR(20) NOT NULL,
			zone_code VARCHAR(10) NOT NULL,
			zone_geometry GEOMETRY(MultiPolygon, 4326) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_climate_zones_geometry ON climate_zones USING GIST(zone_geometry);
	`

	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create test schema: %w", err)
	}

	return nil
}

// SeedTestUsers creates test users with known UUIDs
func SeedTestUsers(ctx context.Context, db *sql.DB, t *testing.T) error {
	users := []struct {
		UserID   string
		Email    string
		Username string
	}{
		{"550e8400-e29b-41d4-a716-446655440001", "test1@example.com", "testuser1"},
		{"550e8400-e29b-41d4-a716-446655440002", "test2@example.com", "testuser2"},
		{"550e8400-e29b-41d4-a716-446655440003", "test3@example.com", "testuser3"},
	}

	for _, user := range users {
		_, err := db.ExecContext(ctx, `
			INSERT INTO users (user_id, email, username)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO NOTHING
		`, user.UserID, user.Email, user.Username)

		if err != nil {
			return fmt.Errorf("failed to seed user %s: %w", user.Username, err)
		}
	}

	return nil
}

// SeedTestPlants creates test plants for garden_plants FK references
func SeedTestPlants(ctx context.Context, db *sql.DB, t *testing.T) error {
	plants := []struct {
		PlantID      string
		BotanicalName string
	}{
		{"650e8400-e29b-41d4-a716-446655440001", "Solanum lycopersicum"},   // Tomato
		{"650e8400-e29b-41d4-a716-446655440002", "Lactuca sativa"},         // Lettuce
		{"650e8400-e29b-41d4-a716-446655440003", "Capsicum annuum"},        // Pepper
		{"650e8400-e29b-41d4-a716-446655440004", "Rosa 'Peace'"},           // Rose
		{"650e8400-e29b-41d4-a716-446655440005", "Thymus vulgaris"},        // Thyme
	}

	for _, plant := range plants {
		_, err := db.ExecContext(ctx, `
			INSERT INTO plants (plant_id, full_botanical_name)
			VALUES ($1, $2)
			ON CONFLICT (plant_id) DO NOTHING
		`, plant.PlantID, plant.BotanicalName)

		if err != nil {
			return fmt.Errorf("failed to seed plant %s: %w", plant.BotanicalName, err)
		}
	}

	return nil
}

// SeedTestClimateZones creates test climate zones for hardiness detection
func SeedTestClimateZones(ctx context.Context, db *sql.DB, t *testing.T) error {
	// San Francisco Bay Area - Zone 10a
	sfZoneGeoJSON := `{
		"type": "MultiPolygon",
		"coordinates": [[[[
			[-122.5194, 37.8749],
			[-122.3194, 37.8749],
			[-122.3194, 37.6749],
			[-122.5194, 37.6749],
			[-122.5194, 37.8749]
		]]]]
	}`

	_, err := db.ExecContext(ctx, `
		INSERT INTO climate_zones (zone_system, zone_code, zone_geometry)
		VALUES ('USDA', '10a', ST_GeomFromGeoJSON($1))
		ON CONFLICT DO NOTHING
	`, sfZoneGeoJSON)

	if err != nil {
		return fmt.Errorf("failed to seed climate zone: %w", err)
	}

	return nil
}

// TestGeoJSON provides test GeoJSON strings for spatial testing
var TestGeoJSON = struct {
	// Valid garden boundary (rectangle in San Francisco)
	ValidGardenBoundary string
	// Valid garden location (center of SF)
	ValidGardenLocation string
	// Valid zone geometry (smaller rectangle within garden)
	ValidZoneGeometry string
	// Valid plant location (point within garden)
	ValidPlantLocation string
	// Invalid polygon (unclosed ring)
	InvalidPolygon string
	// Out of bounds coordinates
	OutOfBoundsPoint string
}{
	ValidGardenBoundary: `{
		"type": "Polygon",
		"coordinates": [[
			[-122.4194, 37.7749],
			[-122.4184, 37.7749],
			[-122.4184, 37.7739],
			[-122.4194, 37.7739],
			[-122.4194, 37.7749]
		]]
	}`,
	ValidGardenLocation: `{
		"type": "Point",
		"coordinates": [-122.4194, 37.7749]
	}`,
	ValidZoneGeometry: `{
		"type": "Polygon",
		"coordinates": [[
			[-122.4190, 37.7746],
			[-122.4188, 37.7746],
			[-122.4188, 37.7744],
			[-122.4190, 37.7744],
			[-122.4190, 37.7746]
		]]
	}`,
	ValidPlantLocation: `{
		"type": "Point",
		"coordinates": [-122.4189, 37.7745]
	}`,
	InvalidPolygon: `{
		"type": "Polygon",
		"coordinates": [[
			[-122.4194, 37.7749],
			[-122.4184, 37.7749],
			[-122.4184, 37.7739]
		]]
	}`,
	OutOfBoundsPoint: `{
		"type": "Point",
		"coordinates": [-200.0, 95.5]
	}`,
}
