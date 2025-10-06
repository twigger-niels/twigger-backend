package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	// Read migration file
	migrationSQL, err := os.ReadFile("migrations/005_add_localization.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Execute migration
	fmt.Println("Applying migration 005_add_localization.sql...")
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}
