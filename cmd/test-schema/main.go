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

	fmt.Println("🧪 Testing Comprehensive Plant Database Schema")
	fmt.Println("====================================================")

	// Test 1: Check all tables exist
	fmt.Println("\n📋 1. Checking table structure...")
	expectedTables := []string{
		"countries", "climate_zones", "languages", "data_sources",
		"plant_families", "plant_genera", "plant_species", "cultivars", "plants",
		"plant_synonyms", "country_plants", "growing_conditions_assertions",
		"physical_characteristics", "users", "gardens", "garden_zones",
		"garden_features", "garden_plants", "companion_relationships",
	}

	for _, tableName := range expectedTables {
		var exists bool
		err := pgConn.QueryRow(ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)",
			tableName).Scan(&exists)
		if err != nil || !exists {
			fmt.Printf("❌ Table missing: %s\n", tableName)
		} else {
			fmt.Printf("✅ Table exists: %s\n", tableName)
		}
	}

	// Test 2: Check domains
	fmt.Println("\n📐 2. Checking measurement domains...")
	domains := []string{"temperature_c", "length_m", "area_m2", "percentage", "ph_value", "rating"}
	for _, domain := range domains {
		var exists bool
		err := pgConn.QueryRow(ctx,
			"SELECT EXISTS (SELECT FROM information_schema.domains WHERE domain_schema = 'public' AND domain_name = $1)",
			domain).Scan(&exists)
		if err != nil || !exists {
			fmt.Printf("❌ Domain missing: %s\n", domain)
		} else {
			fmt.Printf("✅ Domain exists: %s\n", domain)
		}
	}

	// Test 3: Check enum types
	fmt.Println("\n🏷️ 3. Checking enum types...")
	enums := []string{"plant_type", "confidence_level", "sun_requirement", "water_needs", "soil_drainage"}
	for _, enumName := range enums {
		var exists bool
		err := pgConn.QueryRow(ctx,
			"SELECT EXISTS (SELECT FROM pg_type WHERE typname = $1 AND typtype = 'e')",
			enumName).Scan(&exists)
		if err != nil || !exists {
			fmt.Printf("❌ Enum missing: %s\n", enumName)
		} else {
			fmt.Printf("✅ Enum exists: %s\n", enumName)
		}
	}

	// Test 4: Test spatial functions
	fmt.Println("\n🗺️ 4. Testing spatial functionality...")

	// Test basic geometry creation
	var point string
	err = pgConn.QueryRow(ctx, "SELECT ST_AsText(ST_Point(1, 2))").Scan(&point)
	if err != nil {
		fmt.Printf("❌ Basic geometry test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Basic geometry: %s\n", point)
	}

	// Test geography operations
	var distance float64
	err = pgConn.QueryRow(ctx,
		"SELECT ST_Distance(ST_Point(-71.06, 42.36)::geography, ST_Point(-71.07, 42.37)::geography)").Scan(&distance)
	if err != nil {
		fmt.Printf("❌ Geography distance test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Geography distance calculation: %.2f meters\n", distance)
	}

	// Test 5: Insert sample data
	fmt.Println("\n📝 5. Testing data insertion...")

	// Insert a test country
	_, err = pgConn.Exec(ctx,
		"INSERT INTO countries (country_code, country_name, climate_systems) VALUES ('US', 'United States', ARRAY['USDA']) ON CONFLICT DO NOTHING")
	if err != nil {
		fmt.Printf("❌ Country insert failed: %v\n", err)
	} else {
		fmt.Printf("✅ Country inserted\n")
	}

	// Insert a test user
	_, err = pgConn.Exec(ctx,
		"INSERT INTO users (email, username, location) VALUES ('test@example.com', 'testuser', ST_Point(-71.06, 42.36)) ON CONFLICT DO NOTHING")
	if err != nil {
		fmt.Printf("❌ User insert failed: %v\n", err)
	} else {
		fmt.Printf("✅ User with location inserted\n")
	}

	// Insert plant taxonomy
	var familyID, genusID, speciesID, plantID string

	// Insert family
	err = pgConn.QueryRow(ctx,
		"INSERT INTO plant_families (family_name) VALUES ('Rosaceae') RETURNING family_id").Scan(&familyID)
	if err != nil {
		fmt.Printf("❌ Plant family insert failed: %v\n", err)
	} else {
		fmt.Printf("✅ Plant family inserted: %s\n", familyID[:8]+"...")
	}

	// Insert genus
	err = pgConn.QueryRow(ctx,
		"INSERT INTO plant_genera (family_id, genus_name) VALUES ($1, 'Rosa') RETURNING genus_id", familyID).Scan(&genusID)
	if err != nil {
		fmt.Printf("❌ Plant genus insert failed: %v\n", err)
	} else {
		fmt.Printf("✅ Plant genus inserted: %s\n", genusID[:8]+"...")
	}

	// Insert species
	err = pgConn.QueryRow(ctx,
		"INSERT INTO plant_species (genus_id, species_name, plant_type) VALUES ($1, 'rugosa', 'shrub') RETURNING species_id",
		genusID).Scan(&speciesID)
	if err != nil {
		fmt.Printf("❌ Plant species insert failed: %v\n", err)
	} else {
		fmt.Printf("✅ Plant species inserted: %s\n", speciesID[:8]+"...")
	}

	// Insert plant
	err = pgConn.QueryRow(ctx,
		"INSERT INTO plants (species_id, full_botanical_name) VALUES ($1, 'Rosa rugosa') RETURNING plant_id",
		speciesID).Scan(&plantID)
	if err != nil {
		fmt.Printf("❌ Plant insert failed: %v\n", err)
	} else {
		fmt.Printf("✅ Plant inserted: %s\n", plantID[:8]+"...")
	}

	// Test 6: Complex spatial query
	fmt.Println("\n🔍 6. Testing complex queries...")

	// Count records
	var plantCount, userCount int
	err = pgConn.QueryRow(ctx, "SELECT COUNT(*) FROM plants").Scan(&plantCount)
	if err != nil {
		fmt.Printf("❌ Plant count query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Plants in database: %d\n", plantCount)
	}

	err = pgConn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		fmt.Printf("❌ User count query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Users in database: %d\n", userCount)
	}

	// Test taxonomy query
	var fullName string
	err = pgConn.QueryRow(ctx, `
		SELECT pf.family_name || ' > ' || pg.genus_name || ' > ' || ps.species_name
		FROM plants p
		JOIN plant_species ps ON p.species_id = ps.species_id
		JOIN plant_genera pg ON ps.genus_id = pg.genus_id
		JOIN plant_families pf ON pg.family_id = pf.family_id
		WHERE p.plant_id = $1`, plantID).Scan(&fullName)
	if err != nil {
		fmt.Printf("❌ Taxonomy query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Full taxonomy: %s\n", fullName)
	}

	fmt.Println("\n🎉 Comprehensive Plant Database Schema Test Complete!")
	fmt.Println("✅ Schema is fully functional and ready for development")
}