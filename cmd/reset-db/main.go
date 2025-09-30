package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	// Connect to database
	dsn := "postgres://postgres:%2FR%7CP8JqlSJ%5Br%25cl%7D@162.222.181.26:5432/postgres?sslmode=require"

	pgConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		os.Exit(1)
	}
	defer pgConn.Close(ctx)

	fmt.Println("🗑️ Dropping all existing tables and types...")

	// Drop all tables
	dropQueries := []string{
		"DROP TABLE IF EXISTS companion_relationships CASCADE;",
		"DROP TABLE IF EXISTS garden_plants CASCADE;",
		"DROP TABLE IF EXISTS garden_features CASCADE;",
		"DROP TABLE IF EXISTS garden_zones CASCADE;",
		"DROP TABLE IF EXISTS gardens CASCADE;",
		"DROP TABLE IF EXISTS users CASCADE;",
		"DROP TABLE IF EXISTS physical_characteristics CASCADE;",
		"DROP TABLE IF EXISTS growing_conditions_assertions CASCADE;",
		"DROP TABLE IF EXISTS country_plants CASCADE;",
		"DROP TABLE IF EXISTS plant_synonyms CASCADE;",
		"DROP TABLE IF EXISTS plants CASCADE;",
		"DROP TABLE IF EXISTS cultivars CASCADE;",
		"DROP TABLE IF EXISTS plant_species CASCADE;",
		"DROP TABLE IF EXISTS plant_genera CASCADE;",
		"DROP TABLE IF EXISTS plant_families CASCADE;",
		"DROP TABLE IF EXISTS data_sources CASCADE;",
		"DROP TABLE IF EXISTS languages CASCADE;",
		"DROP TABLE IF EXISTS climate_zones CASCADE;",
		"DROP TABLE IF EXISTS countries CASCADE;",

		// Drop old tables
		"DROP TABLE IF EXISTS garden_analysis CASCADE;",
		"DROP TABLE IF EXISTS plant_placements CASCADE;",
		"DROP TABLE IF EXISTS workspace_members CASCADE;",
		"DROP TABLE IF EXISTS workspaces CASCADE;",

		// Drop migration table to reset
		"DROP TABLE IF EXISTS schema_migrations CASCADE;",

		// Drop types
		"DROP TYPE IF EXISTS size_range CASCADE;",
		"DROP TYPE IF EXISTS ph_range CASCADE;",
		"DROP TYPE IF EXISTS temp_range CASCADE;",
		"DROP TYPE IF EXISTS growth_rate CASCADE;",
		"DROP TYPE IF EXISTS soil_drainage CASCADE;",
		"DROP TYPE IF EXISTS water_needs CASCADE;",
		"DROP TYPE IF EXISTS sun_requirement CASCADE;",
		"DROP TYPE IF EXISTS season CASCADE;",
		"DROP TYPE IF EXISTS confidence_level CASCADE;",
		"DROP TYPE IF EXISTS plant_type CASCADE;",

		// Drop old types
		"DROP TYPE IF EXISTS soil_type CASCADE;",
		"DROP TYPE IF EXISTS water_requirement CASCADE;",
		"DROP TYPE IF EXISTS growing_season CASCADE;",
		"DROP TYPE IF EXISTS plant_category CASCADE;",
		"DROP TYPE IF EXISTS subscription_tier CASCADE;",
		"DROP TYPE IF EXISTS user_role CASCADE;",

		// Drop domains
		"DROP DOMAIN IF EXISTS rating CASCADE;",
		"DROP DOMAIN IF EXISTS ph_value CASCADE;",
		"DROP DOMAIN IF EXISTS percentage CASCADE;",
		"DROP DOMAIN IF EXISTS years CASCADE;",
		"DROP DOMAIN IF EXISTS hours CASCADE;",
		"DROP DOMAIN IF EXISTS days CASCADE;",
		"DROP DOMAIN IF EXISTS area_m2 CASCADE;",
		"DROP DOMAIN IF EXISTS weight_kg CASCADE;",
		"DROP DOMAIN IF EXISTS weight_g CASCADE;",
		"DROP DOMAIN IF EXISTS length_m CASCADE;",
		"DROP DOMAIN IF EXISTS length_cm CASCADE;",
		"DROP DOMAIN IF EXISTS length_mm CASCADE;",
		"DROP DOMAIN IF EXISTS temperature_c CASCADE;",
	}

	for _, query := range dropQueries {
		_, err := pgConn.Exec(ctx, query)
		if err != nil {
			fmt.Printf("⚠️ Warning executing %s: %v\n", query, err)
		}
	}

	fmt.Println("✅ Database reset complete!")
	fmt.Println("💡 Now run migrations to set up the new schema")
}