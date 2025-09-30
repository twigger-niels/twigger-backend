package main

import (
	"context"
	"fmt"
	"io"
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

	fmt.Println("📖 Reading comprehensive schema file...")

	// Read the schema file directly
	schemaFile, err := os.Open("migrations/000001_comprehensive_plant_schema.up.sql")
	if err != nil {
		fmt.Printf("❌ Failed to open schema file: %v\n", err)
		os.Exit(1)
	}
	defer schemaFile.Close()

	schemaSQL, err := io.ReadAll(schemaFile)
	if err != nil {
		fmt.Printf("❌ Failed to read schema file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔄 Applying comprehensive plant database schema...")

	// Execute the entire schema
	_, err = pgConn.Exec(ctx, string(schemaSQL))
	if err != nil {
		fmt.Printf("❌ Failed to apply schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Comprehensive plant database schema applied successfully!")

	// Test some basic functionality
	fmt.Println("🧪 Testing schema...")

	// Test PostGIS
	var postgisVersion string
	err = pgConn.QueryRow(ctx, "SELECT PostGIS_Version()").Scan(&postgisVersion)
	if err != nil {
		fmt.Printf("❌ PostGIS test failed: %v\n", err)
	} else {
		fmt.Printf("✅ PostGIS working: %s\n", postgisVersion)
	}

	// Test table creation
	var tableCount int
	err = pgConn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'").Scan(&tableCount)
	if err != nil {
		fmt.Printf("❌ Table count test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Tables created: %d\n", tableCount)
	}

	// Test domain creation
	var domainCount int
	err = pgConn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.domains WHERE domain_schema = 'public'").Scan(&domainCount)
	if err != nil {
		fmt.Printf("❌ Domain count test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Domains created: %d\n", domainCount)
	}

	// Test enum types
	var enumCount int
	err = pgConn.QueryRow(ctx, "SELECT COUNT(*) FROM pg_type WHERE typtype = 'e'").Scan(&enumCount)
	if err != nil {
		fmt.Printf("❌ Enum count test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Enum types created: %d\n", enumCount)
	}

	fmt.Println("🎉 Comprehensive plant database schema is ready!")
}