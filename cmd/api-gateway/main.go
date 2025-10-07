// Package main Twigger Plant Database API
//
//	@title						Twigger Plant Database API
//	@version					1.0.0
//	@description				The Twigger Plant Database API provides comprehensive plant data, garden management, and spatial analysis.
//	@contact.name				Twigger Team
//	@contact.email				support@twigger.com
//
//	@license.name				MIT
//
//	@host						localhost:8080
//	@BasePath					/api/v1
//
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				Firebase JWT token. Format: 'Bearer {token}'
//
//	@schemes					http https
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"twigger-backend/internal/api-gateway/firebase"
	"twigger-backend/internal/api-gateway/handlers"
	"twigger-backend/internal/api-gateway/middleware"
	"twigger-backend/internal/api-gateway/router"

	// Plant Service
	plantService "twigger-backend/backend/plant-service/domain/service"
	plantRepo "twigger-backend/backend/plant-service/infrastructure/database"

	// Garden Service
	gardenService "twigger-backend/backend/garden-service/domain/service"
	gardenRepo "twigger-backend/backend/garden-service/infrastructure/persistence"

	// Swagger docs (imported in router package)
	_ "twigger-backend/docs/swagger"
)

func main() {
	ctx := context.Background()

	// Load configuration
	config := loadConfig()

	// Connect to database
	db, err := connectToDatabase(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Initialize Firebase (optional - will auto-initialize on first request if not done here)
	if err := firebase.InitializeFirebase(ctx); err != nil {
		// Log warning but don't fail - Firebase can be initialized lazily
		log.Printf("Warning: Firebase initialization failed (will retry on first request): %v", err)
		log.Println("If Firebase auth is required, set FIREBASE_PROJECT_ID and FIREBASE_CREDENTIALS_PATH")
	} else {
		log.Println("Successfully initialized Firebase")
	}

	// Initialize repositories
	plantRepository := plantRepo.NewPostgresPlantRepository(db)
	gardenRepository := gardenRepo.NewPostgresGardenRepository(db)
	zoneRepository := gardenRepo.NewPostgresGardenZoneRepository(db)
	gardenPlantRepository := gardenRepo.NewPostgresGardenPlantRepository(db)

	// Initialize services
	plantSvc := plantService.NewPlantService(plantRepository)
	gardenSvc := gardenService.NewGardenService(gardenRepository)
	zoneSvc := gardenService.NewZoneManagementService(zoneRepository, gardenRepository)
	plantPlacementSvc := gardenService.NewPlantPlacementService(gardenPlantRepository, gardenRepository, zoneRepository)

	// Initialize handlers
	h := handlers.NewHandlers(
		db,
		handlers.NewPlantHandler(plantSvc),
		handlers.NewGardenHandler(gardenSvc),
		handlers.NewZoneHandler(zoneSvc),
		handlers.NewPlantPlacementHandler(plantPlacementSvc),
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(config.FirebaseProjectID)

	// Create router with router config
	routerConfig := &router.Config{
		Environment: config.Environment,
	}
	r := router.NewRouter(h, authMiddleware, routerConfig)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting API Gateway on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Config holds application configuration
type Config struct {
	Port              string
	DatabaseURL       string
	FirebaseProjectID string
	Environment       string
	LogLevel          string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:%2FR%7CP8JqlSJ%5Br%25cl%7D@162.222.181.26:5432/postgres?sslmode=require"),
		FirebaseProjectID: getEnv("FIREBASE_PROJECT_ID", "twigger"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}
}

// connectToDatabase creates a database connection
func connectToDatabase(config *Config) (*sql.DB, error) {
	dsn := config.DatabaseURL

	// For local development with Cloud SQL Proxy
	if os.Getenv("CLOUD_SQL_PROXY") == "true" {
		dsn = "postgres://postgres:%2FR%7CP8JqlSJ%5Br%25cl%7D@127.0.0.1:5432/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
