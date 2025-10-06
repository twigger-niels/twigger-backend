package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"twigger-backend/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down] [steps]")
		os.Exit(1)
	}

	// Database URL from environment or default for local development
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default for local development with Cloud SQL Proxy
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	command := os.Args[1]

	switch command {
	case "up":
		fmt.Println("Running migrations...")
		if err := db.RunMigrations(dbURL); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations completed successfully!")

	case "down":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run cmd/migrate/main.go down [steps|all]")
			os.Exit(1)
		}

		stepsArg := os.Args[2]
		if stepsArg == "all" {
			fmt.Println("Rolling back all migrations...")
			// For rolling back all, we use a large number
			if err := db.RollbackMigrations(dbURL, 999); err != nil {
				log.Fatalf("Rollback failed: %v", err)
			}
		} else {
			steps, err := strconv.Atoi(stepsArg)
			if err != nil {
				log.Fatalf("Invalid steps number: %v", err)
			}
			fmt.Printf("Rolling back %d migrations...\n", steps)
			if err := db.RollbackMigrations(dbURL, steps); err != nil {
				log.Fatalf("Rollback failed: %v", err)
			}
		}
		fmt.Println("Rollback completed successfully!")

	case "force":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run cmd/migrate/main.go force [version]")
			os.Exit(1)
		}

		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		fmt.Printf("Forcing migration version to %d...\n", version)
		if err := db.ForceMigrationVersion(dbURL, version); err != nil {
			log.Fatalf("Force version failed: %v", err)
		}
		fmt.Println("Migration version forced successfully!")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down|force] [steps|version]")
		os.Exit(1)
	}
}