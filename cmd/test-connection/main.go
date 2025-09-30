package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	// Test different connection methods
	connections := []struct {
		name string
		dsn  string
	}{
		{
			name: "Direct Connection (Public IP)",
			dsn:  "postgres://postgres:%2FR%7CP8JqlSJ%5Br%25cl%7D@162.222.181.26:5432/postgres?sslmode=require",
		},
		{
			name: "Cloud SQL Proxy (Local)",
			dsn:  "postgres://postgres:%2FR%7CP8JqlSJ%5Br%25cl%7D@127.0.0.1:5432/postgres?sslmode=disable",
		},
	}

	for _, conn := range connections {
		fmt.Printf("\n🔗 Testing %s...\n", conn.name)
		fmt.Printf("DSN: %s\n", conn.dsn)

		// Try to connect
		pgConn, err := pgx.Connect(ctx, conn.dsn)
		if err != nil {
			fmt.Printf("❌ Connection failed: %v\n", err)
			continue
		}
		defer pgConn.Close(ctx)

		// Test basic query
		var version string
		err = pgConn.QueryRow(ctx, "SELECT version()").Scan(&version)
		if err != nil {
			fmt.Printf("❌ Query failed: %v\n", err)
			continue
		}

		fmt.Printf("✅ Connection successful!\n")
		fmt.Printf("📊 PostgreSQL Version: %s\n", version[:50]+"...")

		// Test PostGIS extension
		var postgisVersion string
		err = pgConn.QueryRow(ctx, "SELECT PostGIS_Version()").Scan(&postgisVersion)
		if err != nil {
			fmt.Printf("⚠️  PostGIS not available: %v\n", err)
		} else {
			fmt.Printf("🗺️  PostGIS Version: %s\n", postgisVersion)
		}

		break // Stop on first successful connection
	}

	// Check environment variables
	fmt.Printf("\n📋 Environment Check:\n")
	envVars := []string{"DATABASE_URL", "GOOGLE_APPLICATION_CREDENTIALS", "CLOUD_SQL_PROXY"}
	for _, env := range envVars {
		value := os.Getenv(env)
		if value != "" {
			fmt.Printf("✅ %s: %s\n", env, value)
		} else {
			fmt.Printf("❌ %s: not set\n", env)
		}
	}
}