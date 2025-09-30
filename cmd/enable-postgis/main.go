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

	fmt.Println("✅ Connected to database")

	// Check existing tables
	fmt.Println("\n📋 Checking existing tables...")
	rows, err := pgConn.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name")
	if err != nil {
		fmt.Printf("❌ Failed to query tables: %v\n", err)
		os.Exit(1)
	}

	tables := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			fmt.Printf("❌ Failed to scan table name: %v\n", err)
			continue
		}
		tables = append(tables, tableName)
	}
	rows.Close()

	if len(tables) > 0 {
		fmt.Printf("Found %d existing tables: %v\n", len(tables), tables)
	} else {
		fmt.Println("No tables found")
	}

	// Enable PostGIS extensions
	fmt.Println("\n🗺️ Enabling PostGIS extensions...")

	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS postgis;",
		"CREATE EXTENSION IF NOT EXISTS postgis_topology;",
	}

	for _, ext := range extensions {
		_, err := pgConn.Exec(ctx, ext)
		if err != nil {
			fmt.Printf("❌ Failed to create extension: %v\n", err)
		} else {
			fmt.Printf("✅ Extension enabled: %s\n", ext)
		}
	}

	// Test PostGIS
	fmt.Println("\n🧪 Testing PostGIS...")
	var postgisVersion string
	err = pgConn.QueryRow(ctx, "SELECT PostGIS_Version()").Scan(&postgisVersion)
	if err != nil {
		fmt.Printf("❌ PostGIS not available: %v\n", err)
	} else {
		fmt.Printf("✅ PostGIS Version: %s\n", postgisVersion)
	}

	// Test spatial function
	var result string
	err = pgConn.QueryRow(ctx, "SELECT ST_AsText(ST_Point(1, 2))").Scan(&result)
	if err != nil {
		fmt.Printf("❌ Spatial function test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Spatial function test: %s\n", result)
	}

	fmt.Println("\n🎉 Database connection and PostGIS setup complete!")
}