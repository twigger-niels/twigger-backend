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

	fmt.Println("🔍 Cloud SQL Schema Debug Information")
	fmt.Println("====================================")

	// 1. Check current database
	var currentDB string
	err = pgConn.QueryRow(ctx, "SELECT current_database()").Scan(&currentDB)
	if err != nil {
		fmt.Printf("❌ Failed to get current database: %v\n", err)
	} else {
		fmt.Printf("📍 Current database: %s\n", currentDB)
	}

	// 2. List all databases
	fmt.Println("\n📋 Available databases:")
	rows, err := pgConn.Query(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
	if err != nil {
		fmt.Printf("❌ Failed to list databases: %v\n", err)
	} else {
		for rows.Next() {
			var dbName string
			rows.Scan(&dbName)
			fmt.Printf("  - %s\n", dbName)
		}
		rows.Close()
	}

	// 3. Check current schema
	var currentSchema string
	err = pgConn.QueryRow(ctx, "SELECT current_schema()").Scan(&currentSchema)
	if err != nil {
		fmt.Printf("❌ Failed to get current schema: %v\n", err)
	} else {
		fmt.Printf("\n📁 Current schema: %s\n", currentSchema)
	}

	// 4. List all schemas
	fmt.Println("\n📂 Available schemas:")
	rows, err = pgConn.Query(ctx, "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
	if err != nil {
		fmt.Printf("❌ Failed to list schemas: %v\n", err)
	} else {
		for rows.Next() {
			var schemaName string
			rows.Scan(&schemaName)
			fmt.Printf("  - %s\n", schemaName)
		}
		rows.Close()
	}

	// 5. List tables in current database
	fmt.Println("\n🗂️ Tables in public schema:")
	rows, err = pgConn.Query(ctx, `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		fmt.Printf("❌ Failed to list tables: %v\n", err)
	} else {
		tableCount := 0
		for rows.Next() {
			var tableName, tableType string
			rows.Scan(&tableName, &tableType)
			fmt.Printf("  - %s (%s)\n", tableName, tableType)
			tableCount++
		}
		rows.Close()
		fmt.Printf("Total tables: %d\n", tableCount)
	}

	// 6. Check extensions
	fmt.Println("\n🔌 Installed extensions:")
	rows, err = pgConn.Query(ctx, "SELECT extname, extversion FROM pg_extension ORDER BY extname")
	if err != nil {
		fmt.Printf("❌ Failed to list extensions: %v\n", err)
	} else {
		for rows.Next() {
			var extName, extVersion string
			rows.Scan(&extName, &extVersion)
			fmt.Printf("  - %s (v%s)\n", extName, extVersion)
		}
		rows.Close()
	}

	// 7. Check PostGIS specifically
	fmt.Println("\n🗺️ PostGIS status:")
	var postgisVersion string
	err = pgConn.QueryRow(ctx, "SELECT PostGIS_Version()").Scan(&postgisVersion)
	if err != nil {
		fmt.Printf("❌ PostGIS not available: %v\n", err)
	} else {
		fmt.Printf("✅ PostGIS version: %s\n", postgisVersion)
	}

	// 8. Check some specific tables from our schema
	fmt.Println("\n🌱 Plant database specific tables:")
	plantTables := []string{"plants", "plant_families", "plant_genera", "plant_species", "gardens", "users"}
	for _, tableName := range plantTables {
		var exists bool
		err := pgConn.QueryRow(ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)",
			tableName).Scan(&exists)
		if err != nil {
			fmt.Printf("❌ Error checking %s: %v\n", tableName, err)
		} else if exists {
			var rowCount int
			err = pgConn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&rowCount)
			if err != nil {
				fmt.Printf("✅ %s exists (count query failed)\n", tableName)
			} else {
				fmt.Printf("✅ %s exists (%d rows)\n", tableName, rowCount)
			}
		} else {
			fmt.Printf("❌ %s does not exist\n", tableName)
		}
	}

	// 9. Check connection details
	fmt.Println("\n🔗 Connection information:")
	var host, port, user string
	err = pgConn.QueryRow(ctx, "SELECT inet_server_addr(), inet_server_port(), current_user").Scan(&host, &port, &user)
	if err != nil {
		fmt.Printf("❌ Failed to get connection info: %v\n", err)
	} else {
		fmt.Printf("  Host: %s\n", host)
		fmt.Printf("  Port: %s\n", port)
		fmt.Printf("  User: %s\n", user)
	}

	// 10. Show search_path
	var searchPath string
	err = pgConn.QueryRow(ctx, "SHOW search_path").Scan(&searchPath)
	if err != nil {
		fmt.Printf("❌ Failed to get search_path: %v\n", err)
	} else {
		fmt.Printf("  Search path: %s\n", searchPath)
	}

	fmt.Println("\n🎯 Cloud SQL Studio Troubleshooting Tips:")
	fmt.Println("1. Ensure you're connecting to the 'postgres' database")
	fmt.Println("2. Check that you're looking in the 'public' schema")
	fmt.Println("3. Try refreshing the schema view in Cloud SQL Studio")
	fmt.Println("4. Verify your user has proper permissions")
	fmt.Println("5. Check if Cloud SQL Studio is using a different connection")
}