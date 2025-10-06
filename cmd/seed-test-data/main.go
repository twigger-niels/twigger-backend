package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Fixed UUIDs for consistent references
var (
	// Languages
	englishID = "550e8400-e29b-41d4-a716-446655440001"
	spanishID = "550e8400-e29b-41d4-a716-446655440002"
	frenchID  = "550e8400-e29b-41d4-a716-446655440003"
	germanID  = "550e8400-e29b-41d4-a716-446655440004"

	// Countries
	usaID     = "650e8400-e29b-41d4-a716-446655440001"
	mexicoID  = "650e8400-e29b-41d4-a716-446655440002"
	franceID  = "650e8400-e29b-41d4-a716-446655440003"
	germanyID = "650e8400-e29b-41d4-a716-446655440004"
	ukID      = "650e8400-e29b-41d4-a716-446655440005"

	// Plant taxonomy
	solanaceaeID = "750e8400-e29b-41d4-a716-446655440001" // Nightshade family
	solanumID    = "750e8400-e29b-41d4-a716-446655440002" // Solanum genus
	lycopersicon = "750e8400-e29b-41d4-a716-446655440003" // Tomato species

	rosacea      = "750e8400-e29b-41d4-a716-446655440004" // Rose family
	fragaria     = "750e8400-e29b-41d4-a716-446655440005" // Strawberry genus
	fragariaAnan = "750e8400-e29b-41d4-a716-446655440006" // Garden strawberry

	// Plants
	tomatoID     = "850e8400-e29b-41d4-a716-446655440001"
	strawberryID = "850e8400-e29b-41d4-a716-446655440002"

	// Data source
	seedDataSourceID = "950e8400-e29b-41d4-a716-446655440001"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	fmt.Println("🌱 Seeding test data...")

	// Use transaction for atomic seeding
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Languages
	fmt.Println("\n1. Seeding languages...")
	seedLanguages(tx)

	// 2. Countries (note: US already exists, skip if duplicate)
	fmt.Println("2. Seeding countries...")
	seedCountries(tx)

	// 3. Data source
	fmt.Println("3. Seeding data source...")
	seedDataSource(tx)

	// 4. Plant families
	fmt.Println("4. Seeding plant families...")
	seedPlantFamilies(tx)

	// 5. Plant genera
	fmt.Println("5. Seeding plant genera...")
	seedPlantGenera(tx)

	// 6. Plant species
	fmt.Println("6. Seeding plant species...")
	seedPlantSpecies(tx)

	// 7. Plants
	fmt.Println("7. Seeding plants...")
	seedPlants(tx)

	// 8. Plant common names (localized)
	fmt.Println("8. Seeding plant common names...")
	seedPlantCommonNames(tx)

	// 9. Plant descriptions (localized)
	fmt.Println("9. Seeding plant descriptions...")
	seedPlantDescriptions(tx)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	fmt.Println("\n✅ Test data seeded successfully!")
	fmt.Println("\nYou can now test:")
	fmt.Println("  curl http://localhost:8080/api/v1/plants/search?q=tomato")
	fmt.Println("  curl -H \"Accept-Language: es-MX\" http://localhost:8080/api/v1/plants/search?q=tomate")
	fmt.Println("  curl -H \"Accept-Language: fr\" http://localhost:8080/api/v1/plants/search?q=fraise")
}

func seedLanguages(tx *sql.Tx) {
	languages := []struct {
		id     string
		code   string
		name   string
		active bool
	}{
		{englishID, "en", "English", true},
		{spanishID, "es", "Spanish", true},
		{frenchID, "fr", "French", true},
		{germanID, "de", "German", true},
	}

	for _, lang := range languages {
		_, err := tx.Exec(`
			INSERT INTO languages (language_id, language_code, language_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (language_id) DO NOTHING
		`, lang.id, lang.code, lang.name, lang.active)
		if err != nil {
			log.Fatalf("Failed to insert language %s: %v", lang.code, err)
		}
		fmt.Printf("  ✓ %s (%s)\n", lang.name, lang.code)
	}
}

func seedCountries(tx *sql.Tx) {
	countries := []struct {
		id      string
		code    string
		name    string
		climate string
	}{
		{usaID, "US", "United States", "{}"},
		{mexicoID, "MX", "Mexico", "{}"},
		{franceID, "FR", "France", "{}"},
		{germanyID, "DE", "Germany", "{}"},
		{ukID, "GB", "United Kingdom", "{}"},
	}

	for _, country := range countries {
		// Try insert, if fails on country_code uniqueness, update the ID
		_, err := tx.Exec(`
			INSERT INTO countries (country_id, country_code, country_name, climate_systems)
			VALUES ($1, $2, $3, $4::text[])
			ON CONFLICT (country_code) DO UPDATE
			SET country_id = EXCLUDED.country_id,
			    country_name = EXCLUDED.country_name
		`, country.id, country.code, country.name, country.climate)
		if err != nil {
			log.Fatalf("Failed to insert country %s: %v", country.code, err)
		}
		fmt.Printf("  ✓ %s (%s)\n", country.name, country.code)
	}
}

func seedDataSource(tx *sql.Tx) {
	_, err := tx.Exec(`
		INSERT INTO data_sources (source_id, source_name, source_type, website_url, reliability_score)
		VALUES ($1, 'Test Seed Data', 'website', 'https://example.com', 5)
		ON CONFLICT (source_id) DO NOTHING
	`, seedDataSourceID)
	if err != nil {
		log.Fatalf("Failed to insert data source: %v", err)
	}
	fmt.Println("  ✓ Test Seed Data")
}

func seedPlantFamilies(tx *sql.Tx) {
	families := []struct {
		id          string
		name        string
		commonName  string
	}{
		{solanaceaeID, "Solanaceae", "Nightshade family"},
		{rosacea, "Rosaceae", "Rose family"},
	}

	for _, fam := range families {
		_, err := tx.Exec(`
			INSERT INTO plant_families (family_id, family_name, common_name)
			VALUES ($1, $2, $3)
			ON CONFLICT (family_name) DO NOTHING
		`, fam.id, fam.name, fam.commonName)
		if err != nil {
			log.Fatalf("Failed to insert family %s: %v", fam.name, err)
		}
		fmt.Printf("  ✓ %s (%s)\n", fam.name, fam.commonName)
	}
}

func seedPlantGenera(tx *sql.Tx) {
	genera := []struct {
		id       string
		familyID string
		name     string
	}{
		{solanumID, solanaceaeID, "Solanum"},
		{fragaria, rosacea, "Fragaria"},
	}

	for _, genus := range genera {
		_, err := tx.Exec(`
			INSERT INTO plant_genera (genus_id, family_id, genus_name)
			VALUES ($1, $2, $3)
			ON CONFLICT (genus_id) DO NOTHING
		`, genus.id, genus.familyID, genus.name)
		if err != nil {
			log.Fatalf("Failed to insert genus %s: %v", genus.name, err)
		}
		fmt.Printf("  ✓ %s\n", genus.name)
	}
}

func seedPlantSpecies(tx *sql.Tx) {
	species := []struct {
		id       string
		genusID  string
		name     string
		plantType string
	}{
		{lycopersicon, solanumID, "lycopersicum", "vegetable"},
		{fragariaAnan, fragaria, "ananassa", "fruit"},
	}

	for _, sp := range species {
		_, err := tx.Exec(`
			INSERT INTO plant_species (species_id, genus_id, species_name, plant_type)
			VALUES ($1, $2, $3, $4::plant_type)
			ON CONFLICT (species_id) DO NOTHING
		`, sp.id, sp.genusID, sp.name, sp.plantType)
		if err != nil {
			log.Fatalf("Failed to insert species %s: %v", sp.name, err)
		}
		fmt.Printf("  ✓ %s (%s)\n", sp.name, sp.plantType)
	}
}

func seedPlants(tx *sql.Tx) {
	plants := []struct {
		id              string
		familyID        string
		genusID         string
		speciesID       string
		botanicalName   string
	}{
		{tomatoID, solanaceaeID, solanumID, lycopersicon, "Solanum lycopersicum"},
		{strawberryID, rosacea, fragaria, fragariaAnan, "Fragaria × ananassa"},
	}

	for _, plant := range plants {
		_, err := tx.Exec(`
			INSERT INTO plants (plant_id, family_id, genus_id, species_id, full_botanical_name)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (plant_id) DO NOTHING
		`, plant.id, plant.familyID, plant.genusID, plant.speciesID, plant.botanicalName)
		if err != nil {
			log.Fatalf("Failed to insert plant %s: %v", plant.botanicalName, err)
		}
		fmt.Printf("  ✓ %s\n", plant.botanicalName)
	}
}

func seedPlantCommonNames(tx *sql.Tx) {
	names := []struct {
		plantID    string
		langID     string
		countryID  *string
		name       string
		isPrimary  bool
	}{
		// Tomato - English
		{tomatoID, englishID, nil, "Tomato", true},
		{tomatoID, englishID, &usaID, "Tomato", true},
		{tomatoID, englishID, &ukID, "Tomato", true},

		// Tomato - Spanish
		{tomatoID, spanishID, nil, "Tomate", true},
		{tomatoID, spanishID, &mexicoID, "Jitomate", true},

		// Tomato - French
		{tomatoID, frenchID, nil, "Tomate", true},

		// Tomato - German
		{tomatoID, germanID, nil, "Tomate", true},

		// Strawberry - English
		{strawberryID, englishID, nil, "Strawberry", true},
		{strawberryID, englishID, &usaID, "Strawberry", true},
		{strawberryID, englishID, &ukID, "Strawberry", true},

		// Strawberry - Spanish
		{strawberryID, spanishID, nil, "Fresa", true},
		{strawberryID, spanishID, &mexicoID, "Fresa", true},

		// Strawberry - French
		{strawberryID, frenchID, nil, "Fraise", true},

		// Strawberry - German
		{strawberryID, germanID, nil, "Erdbeere", true},
	}

	for _, name := range names {
		nameID := uuid.New().String()
		countryStr := "global"
		if name.countryID != nil {
			countryStr = *name.countryID
		}

		_, err := tx.Exec(`
			INSERT INTO plant_common_names (
				common_name_id, plant_id, language_id, country_id,
				common_name, is_primary, source_id
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (plant_id, language_id, country_id, common_name) DO NOTHING
		`, nameID, name.plantID, name.langID, name.countryID, name.name, name.isPrimary, seedDataSourceID)

		if err != nil {
			log.Fatalf("Failed to insert common name %s: %v", name.name, err)
		}
		fmt.Printf("  ✓ %s (%s)\n", name.name, countryStr)
	}
}

func seedPlantDescriptions(tx *sql.Tx) {
	descriptions := []struct {
		plantID         string
		langID          string
		shortDesc       string
		fullDesc        string
	}{
		{
			tomatoID,
			englishID,
			"A popular red fruit eaten as a vegetable, essential for salads and sauces.",
			"The tomato is the edible berry of the plant Solanum lycopersicum. Tomatoes are a significant source of umami flavor and are widely cultivated worldwide.",
		},
		{
			tomatoID,
			spanishID,
			"Un fruto rojo popular consumido como verdura, esencial para ensaladas y salsas.",
			"El tomate es la baya comestible de la planta Solanum lycopersicum. Los tomates son una fuente importante de sabor umami y se cultivan ampliamente en todo el mundo.",
		},
		{
			strawberryID,
			englishID,
			"A sweet red fruit popular in desserts and fresh eating.",
			"The garden strawberry is a widely grown hybrid species of the genus Fragaria. It is cultivated worldwide for its fruit, which is appreciated for its characteristic aroma, bright red color, and sweet taste.",
		},
		{
			strawberryID,
			frenchID,
			"Un fruit rouge sucré populaire dans les desserts.",
			"La fraise de jardin est une espèce hybride largement cultivée du genre Fragaria. Elle est cultivée dans le monde entier pour son fruit, apprécié pour son arôme caractéristique, sa couleur rouge vif et son goût sucré.",
		},
	}

	for _, desc := range descriptions {
		descID := uuid.New().String()

		_, err := tx.Exec(`
			INSERT INTO plant_descriptions (
				description_id, plant_id, language_id, country_id,
				short_description, full_description, is_primary, source_id
			)
			VALUES ($1, $2, $3, NULL, $4, $5, true, $6)
			ON CONFLICT (plant_id, language_id, country_id) DO NOTHING
		`, descID, desc.plantID, desc.langID, desc.shortDesc, desc.fullDesc, seedDataSourceID)

		if err != nil {
			log.Fatalf("Failed to insert description: %v", err)
		}
		fmt.Printf("  ✓ Description for plant %s (lang %s)\n", desc.plantID[:8], desc.langID[:8])
	}
}
