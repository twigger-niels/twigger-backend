package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check languages
	var langCount int
	db.QueryRow("SELECT COUNT(*) FROM languages").Scan(&langCount)
	fmt.Printf("Languages: %d rows\n", langCount)

	if langCount > 0 {
		fmt.Println("\nExisting languages:")
		rows, _ := db.Query("SELECT language_code, language_name FROM languages ORDER BY language_code LIMIT 10")
		for rows.Next() {
			var code, name string
			rows.Scan(&code, &name)
			fmt.Printf("  - %s (%s)\n", code, name)
		}
		rows.Close()
	}

	// Check countries
	var countryCount int
	db.QueryRow("SELECT COUNT(*) FROM countries").Scan(&countryCount)
	fmt.Printf("\nCountries: %d rows\n", countryCount)

	if countryCount > 0 {
		fmt.Println("\nExisting countries:")
		rows, _ := db.Query("SELECT country_code, country_name FROM countries ORDER BY country_code LIMIT 10")
		for rows.Next() {
			var code, name string
			rows.Scan(&code, &name)
			fmt.Printf("  - %s (%s)\n", code, name)
		}
		rows.Close()
	}

	// Check plants
	var plantCount int
	db.QueryRow("SELECT COUNT(*) FROM plants").Scan(&plantCount)
	fmt.Printf("\nPlants: %d rows\n", plantCount)

	// Check plant_common_names
	var nameCount int
	db.QueryRow("SELECT COUNT(*) FROM plant_common_names").Scan(&nameCount)
	fmt.Printf("Plant common names: %d rows\n", nameCount)
}
