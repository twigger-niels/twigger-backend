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

	fmt.Println("=== gardens table columns ===")
	rows, _ := db.Query(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'gardens'
		ORDER BY ordinal_position
	`)
	defer rows.Close()
	for rows.Next() {
		var col, dtype string
		rows.Scan(&col, &dtype)
		fmt.Printf("  %s (%s)\n", col, dtype)
	}

	fmt.Println("\n=== garden_zones table columns ===")
	rows2, _ := db.Query(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'garden_zones'
		ORDER BY ordinal_position
	`)
	defer rows2.Close()
	for rows2.Next() {
		var col, dtype string
		rows2.Scan(&col, &dtype)
		fmt.Printf("  %s (%s)\n", col, dtype)
	}

	fmt.Println("\n=== garden_plants table columns ===")
	rows3, _ := db.Query(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'garden_plants'
		ORDER BY ordinal_position
	`)
	defer rows3.Close()
	for rows3.Next() {
		var col, dtype string
		rows3.Scan(&col, &dtype)
		fmt.Printf("  %s (%s)\n", col, dtype)
	}
}
