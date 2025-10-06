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

	// Check schema_migrations table
	fmt.Println("\n=== Checking migration status ===")
	var version int
	var dirty bool
	err = db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	if err != nil {
		log.Printf("Error reading schema_migrations: %v", err)
	} else {
		fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)
	}

	// List all tables
	fmt.Println("\n=== Existing tables ===")
	rows, err := db.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Error reading table name: %v", err)
		}
		fmt.Printf("  - %s\n", tableName)
	}

	// Check for specific critical tables
	fmt.Println("\n=== Checking critical tables ===")
	criticalTables := []string{
		"plants",
		"plant_families",
		"plant_genera",
		"plant_species",
		"plant_common_names",
		"gardens",
		"garden_zones",
	}

	for _, table := range criticalTables {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists)
		if err != nil {
			log.Printf("Error checking %s: %v", table, err)
			continue
		}
		status := "❌ MISSING"
		if exists {
			status = "✅ EXISTS"
		}
		fmt.Printf("  %s: %s\n", table, status)
	}
}
