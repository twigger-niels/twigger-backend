package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// Database connection
	db, cleanup, err := connectToDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer cleanup()

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to Cloud SQL PostgreSQL!")

	// Health check endpoint
	http.HandleFunc("/health", healthHandler(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func connectToDatabase(ctx context.Context) (*pgxpool.Pool, func(), error) {
	// Get database URL from environment or use default
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default for direct connection (development/testing)
		dsn = "postgres://postgres:%2FR%7CP8JqlSJ%5Br%25cl%7D@162.222.181.26:5432/postgres?sslmode=require"

		// For local development with Cloud SQL Proxy
		if os.Getenv("CLOUD_SQL_PROXY") == "true" {
			dsn = "postgres://postgres:%2FR%7CP8JqlSJ%5Br%25cl%7D@127.0.0.1:5432/postgres?sslmode=disable"
		}
	}

	// Create a connection pool
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Note: Cloud SQL connector disabled for now due to permissions
	// Will use direct connection with authorized networks instead

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup, nil
}

func healthHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Check database health
		if err := db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"unhealthy","database":"error","error":"%s"}`, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"healthy","database":"connected"}`)
	}
}