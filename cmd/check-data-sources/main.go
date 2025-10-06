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

	rows, _ := db.Query(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'data_sources'
		ORDER BY ordinal_position
	`)
	defer rows.Close()

	fmt.Println("data_sources columns:")
	for rows.Next() {
		var col, dtype string
		rows.Scan(&col, &dtype)
		fmt.Printf("  %s (%s)\n", col, dtype)
	}
}
