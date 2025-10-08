package router

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	"twigger-backend/internal/api-gateway/handlers"
	"twigger-backend/internal/api-gateway/middleware"
)

// Config holds router configuration
type Config struct {
	Environment string
}

// NewRouter creates and configures the main application router
func NewRouter(h *handlers.Handlers, authMiddleware *middleware.AuthMiddleware, config *Config) *mux.Router {
	r := mux.NewRouter()

	// Global middleware
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware)

	// Rate limiting middleware (token bucket algorithm with endpoint-specific limits)
	rateLimiter := middleware.NewRateLimiter()
	r.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Health check endpoints (no auth required)
	r.HandleFunc("/health", h.HealthHandler.HealthCheck).Methods("GET")
	r.HandleFunc("/ready", h.HealthHandler.ReadinessCheck).Methods("GET")

	// Swagger documentation endpoint (no auth required)
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Auth routes
	authRouter := api.PathPrefix("/auth").Subrouter()

	// POST /api/v1/auth/verify - Complete authentication (uses middleware to verify token)
	verifyRouter := authRouter.PathPrefix("/verify").Subrouter()
	verifyRouter.Use(authMiddleware.RequireAuth)
	verifyRouter.HandleFunc("", h.AuthHandler.HandleVerify).Methods("POST", "OPTIONS")

	// POST /api/v1/auth/register - Register with optional username (requires auth)
	registerRouter := authRouter.PathPrefix("/register").Subrouter()
	registerRouter.Use(authMiddleware.RequireAuth)
	registerRouter.HandleFunc("", h.AuthHandler.HandleRegister).Methods("POST", "OPTIONS")

	// POST /api/v1/auth/logout - Logout (requires auth)
	logoutRouter := authRouter.PathPrefix("/logout").Subrouter()
	logoutRouter.Use(authMiddleware.RequireAuth)
	logoutRouter.HandleFunc("", h.AuthHandler.HandleLogout).Methods("POST", "OPTIONS")

	// GET /api/v1/auth/me - Get current user (requires auth)
	meRouter := authRouter.PathPrefix("/me").Subrouter()
	meRouter.Use(authMiddleware.RequireAuth)
	meRouter.HandleFunc("", h.AuthHandler.HandleMe).Methods("GET", "OPTIONS")

	// Plant routes (optional auth for search, required for modifications)
	plantRouter := api.PathPrefix("/plants").Subrouter()

	// Public plant routes (optional auth)
	plantRouter.HandleFunc("/search", h.PlantHandler.SearchPlants).Methods("GET")
	plantRouter.HandleFunc("/recommend", h.PlantHandler.RecommendPlants).Methods("GET")
	plantRouter.HandleFunc("/family/{name}", h.PlantHandler.FindByFamily).Methods("GET")
	plantRouter.HandleFunc("/genus/{name}", h.PlantHandler.FindByGenus).Methods("GET")
	plantRouter.HandleFunc("/{id}", h.PlantHandler.GetPlant).Methods("GET")
	plantRouter.HandleFunc("/{id}/companions", h.PlantHandler.GetCompanions).Methods("GET")

	// Admin plant routes (require auth)
	authPlantRouter := plantRouter.NewRoute().Subrouter()
	authPlantRouter.Use(authMiddleware.RequireAuth)
	authPlantRouter.HandleFunc("", h.PlantHandler.CreatePlant).Methods("POST")
	authPlantRouter.HandleFunc("/{id}", h.PlantHandler.UpdatePlant).Methods("PUT")
	authPlantRouter.HandleFunc("/{id}", h.PlantHandler.DeletePlant).Methods("DELETE")

	// Garden routes (all require auth)
	gardenRouter := api.PathPrefix("/gardens").Subrouter()
	gardenRouter.Use(authMiddleware.RequireAuth)

	gardenRouter.HandleFunc("", h.GardenHandler.ListGardens).Methods("GET")
	gardenRouter.HandleFunc("", h.GardenHandler.CreateGarden).Methods("POST")
	gardenRouter.HandleFunc("/stats", h.GardenHandler.GetGardenStats).Methods("GET")
	gardenRouter.HandleFunc("/nearby", h.GardenHandler.FindNearbyGardens).Methods("GET")
	gardenRouter.HandleFunc("/{id}", h.GardenHandler.GetGarden).Methods("GET")
	gardenRouter.HandleFunc("/{id}", h.GardenHandler.UpdateGarden).Methods("PUT")
	gardenRouter.HandleFunc("/{id}", h.GardenHandler.DeleteGarden).Methods("DELETE")

	// Zone routes (nested under gardens)
	gardenRouter.HandleFunc("/{id}/zones", h.ZoneHandler.CreateZone).Methods("POST")
	gardenRouter.HandleFunc("/{id}/zones", h.ZoneHandler.ListGardenZones).Methods("GET")

	// Zone routes (standalone)
	zoneRouter := api.PathPrefix("/zones").Subrouter()
	zoneRouter.Use(authMiddleware.RequireAuth)

	zoneRouter.HandleFunc("/{id}", h.ZoneHandler.GetZone).Methods("GET")
	zoneRouter.HandleFunc("/{id}", h.ZoneHandler.UpdateZone).Methods("PUT")
	zoneRouter.HandleFunc("/{id}", h.ZoneHandler.DeleteZone).Methods("DELETE")
	zoneRouter.HandleFunc("/{id}/area", h.ZoneHandler.CalculateZoneArea).Methods("GET")

	// Plant placement routes (nested under gardens)
	gardenRouter.HandleFunc("/{id}/plants", h.PlantPlacementHandler.PlacePlant).Methods("POST")
	gardenRouter.HandleFunc("/{id}/plants", h.PlantPlacementHandler.ListGardenPlants).Methods("GET")
	gardenRouter.HandleFunc("/{id}/plants/bulk", h.PlantPlacementHandler.BulkPlacePlants).Methods("POST")

	// Garden plant routes (standalone)
	gardenPlantRouter := api.PathPrefix("/garden-plants").Subrouter()
	gardenPlantRouter.Use(authMiddleware.RequireAuth)

	gardenPlantRouter.HandleFunc("/{id}", h.PlantPlacementHandler.GetGardenPlant).Methods("GET")
	gardenPlantRouter.HandleFunc("/{id}", h.PlantPlacementHandler.UpdatePlantPlacement).Methods("PUT")
	gardenPlantRouter.HandleFunc("/{id}", h.PlantPlacementHandler.RemovePlant).Methods("DELETE")

	return r
}
