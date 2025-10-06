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
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	tables := []string{"plant_species", "plant_families", "plant_genera", "cultivars"}

	for _, table := range tables {
		fmt.Printf("\n=== Columns in %s ===\n", table)
		rows, err := db.Query(`
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1
			ORDER BY ordinal_position
		`, table)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var colName, dataType string
			if err := rows.Scan(&colName, &dataType); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  - %s (%s)\n", colName, dataType)
		}
	}
}
