package handlers

import "database/sql"

// Handlers aggregates all API handlers
type Handlers struct {
	PlantHandler          *PlantHandler
	GardenHandler         *GardenHandler
	ZoneHandler           *ZoneHandler
	PlantPlacementHandler *PlantPlacementHandler
	HealthHandler         *HealthHandler
	AuthHandler           *AuthHandler
}

// NewHandlers creates all handlers
func NewHandlers(db *sql.DB, plantHandler *PlantHandler, gardenHandler *GardenHandler, zoneHandler *ZoneHandler, plantPlacementHandler *PlantPlacementHandler) *Handlers {
	return &Handlers{
		PlantHandler:          plantHandler,
		GardenHandler:         gardenHandler,
		ZoneHandler:           zoneHandler,
		PlantPlacementHandler: plantPlacementHandler,
		HealthHandler:         NewHealthHandler(db),
		AuthHandler:           NewAuthHandler(db),
	}
}
