package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("🌱 Seeding minimal test data...")

	// 1. Languages
	langIDs := map[string]string{}
	for _, lang := range []struct{ code, name string }{
		{"en", "English"},
		{"es", "Spanish"},
		{"fr", "French"},
		{"de", "German"},
	} {
		var id string
		err := db.QueryRow(`
			INSERT INTO languages (language_id, language_code, language_name, is_active)
			VALUES (uuid_generate_v4(), $1, $2, true)
			ON CONFLICT (language_code) DO UPDATE SET language_code = EXCLUDED.language_code
			RETURNING language_id
		`, lang.code, lang.name).Scan(&id)
		if err != nil {
			log.Fatalf("Failed to insert language %s: %v", lang.code, err)
		}
		langIDs[lang.code] = id
		fmt.Printf("  ✓ %s (%s): %s\n", lang.name, lang.code, id[:8])
	}

	// 2. Get or create tomato plant
	var tomatoID string
	err = db.QueryRow("SELECT plant_id FROM plants WHERE full_botanical_name = 'Solanum lycopersicum'").Scan(&tomatoID)
	if err == sql.ErrNoRows {
		// Create minimal plant hierarchy
		fmt.Println("\n📦 Creating Solanum lycopersicum...")

		// Family
		var familyID string
		err = db.QueryRow(`INSERT INTO plant_families (family_id, family_name) VALUES (uuid_generate_v4(), 'Solanaceae')
			ON CONFLICT (family_name) DO UPDATE SET family_name = EXCLUDED.family_name RETURNING family_id`).Scan(&familyID)
		if err != nil {
			log.Fatalf("Failed to create family: %v", err)
		}
		fmt.Printf("    Family ID: %s\n", familyID[:8])

		// Genus
		var genusID string
		err = db.QueryRow(`INSERT INTO plant_genera (genus_id, family_id, genus_name) VALUES (uuid_generate_v4(), $1, 'Solanum')
			ON CONFLICT DO NOTHING RETURNING genus_id`, familyID).Scan(&genusID)
		if err == sql.ErrNoRows {
			db.QueryRow("SELECT genus_id FROM plant_genera WHERE genus_name = 'Solanum'").Scan(&genusID)
		} else if err != nil {
			log.Fatalf("Failed to create genus: %v", err)
		}
		fmt.Printf("    Genus ID: %s\n", genusID[:8])

		// Species
		var speciesID string
		err = db.QueryRow(`INSERT INTO plant_species (species_id, genus_id, species_name, plant_type) VALUES (uuid_generate_v4(), $1, 'lycopersicum', 'annual'::plant_type)
			ON CONFLICT DO NOTHING RETURNING species_id`, genusID).Scan(&speciesID)
		if err == sql.ErrNoRows {
			db.QueryRow("SELECT species_id FROM plant_species WHERE species_name = 'lycopersicum'").Scan(&speciesID)
		} else if err != nil {
			log.Fatalf("Failed to create species: %v", err)
		}
		fmt.Printf("    Species ID: %s\n", speciesID[:8])

		// Plant
		err = db.QueryRow(`INSERT INTO plants (plant_id, species_id, full_botanical_name) VALUES (uuid_generate_v4(), $1, 'Solanum lycopersicum')
			RETURNING plant_id`, speciesID).Scan(&tomatoID)
		if err != nil {
			log.Fatalf("Failed to create plant: %v", err)
		}
		fmt.Printf("  ✓ Created tomato: %s\n", tomatoID[:8])
	} else if err != nil {
		log.Fatalf("Error checking tomato: %v", err)
	} else {
		fmt.Printf("\n📦 Found existing tomato: %s\n", tomatoID[:8])
	}

	// 3. Add localized names
	fmt.Println("\n🌍 Adding localized names...")
	names := []struct {
		lang, name string
	}{
		{"en", "Tomato"},
		{"es", "Tomate"},
		{"fr", "Tomate"},
		{"de", "Tomate"},
	}

	for _, n := range names {
		langID := langIDs[n.lang]
		_, err := db.Exec(`
			INSERT INTO plant_common_names (common_name_id, plant_id, language_id, country_id, common_name, is_primary)
			VALUES ($1, $2, $3, NULL, $4, true)
			ON CONFLICT (plant_id, language_id, country_id, common_name) DO NOTHING
		`, uuid.New().String(), tomatoID, langID, n.name)
		if err != nil {
			log.Fatalf("Failed to insert name %s: %v", n.name, err)
		}
		fmt.Printf("  ✓ %s: %s\n", n.lang, n.name)
	}

	fmt.Println("\n✅ Test data seeded successfully!")
	fmt.Println("\nTest the API:")
	fmt.Println("  curl http://localhost:8080/api/v1/plants/search?q=tomato")
	fmt.Println("  curl -H \"Accept-Language: es\" http://localhost:8080/api/v1/plants/search?q=tomate")
	fmt.Println("  curl -H \"Accept-Language: fr\" http://localhost:8080/api/v1/plants/search?q=tomate")
}
