package handlers

import (
	"encoding/json"
	"net/http"

	gardenService "twigger-backend/backend/garden-service/domain/service"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/internal/api-gateway/utils"
)

// PlantPlacementHandler handles plant placement-related HTTP requests
type PlantPlacementHandler struct {
	service gardenService.PlantPlacementService
}

// NewPlantPlacementHandler creates a new plant placement handler
func NewPlantPlacementHandler(service gardenService.PlantPlacementService) *PlantPlacementHandler {
	return &PlantPlacementHandler{
		service: service,
	}
}

// PlacePlant handles POST /api/v1/gardens/:id/plants
func (h *PlantPlacementHandler) PlacePlant(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	var req placePlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	gardenPlant := &entity.GardenPlant{
		GardenID:       gardenID,
		PlantID:        req.PlantID,
		LocationGeoJSON: req.LocationGeoJSON,
		ZoneID:         req.ZoneID,
		Quantity:       req.Quantity,
		Notes:          req.Notes,
	}

	created, err := h.service.PlacePlant(r.Context(), gardenPlant)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondCreated(w, created)
}

// BulkPlacePlants handles POST /api/v1/gardens/:id/plants/bulk
func (h *PlantPlacementHandler) BulkPlacePlants(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	var req bulkPlacePlantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Build garden plants from request
	var gardenPlants []*entity.GardenPlant
	for _, plant := range req.Plants {
		gardenPlants = append(gardenPlants, &entity.GardenPlant{
			GardenID:       gardenID,
			PlantID:        plant.PlantID,
			LocationGeoJSON: plant.LocationGeoJSON,
			ZoneID:         plant.ZoneID,
			Quantity:       plant.Quantity,
			Notes:          plant.Notes,
		})
	}

	created, err := h.service.BulkPlacePlants(r.Context(), gardenPlants)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondCreated(w, created)
}

// ListGardenPlants handles GET /api/v1/gardens/:id/plants
func (h *PlantPlacementHandler) ListGardenPlants(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	// Build optional filter
	filter := &gardenService.GardenPlantFilter{
		ActiveOnly: utils.GetQueryParamBool(r, "active_only", false),
	}

	if zoneID := utils.GetQueryParam(r, "zone_id"); zoneID != "" {
		filter.ZoneID = &zoneID
	}

	if healthStatusStr := utils.GetQueryParam(r, "health_status"); healthStatusStr != "" {
		hs := entity.HealthStatus(healthStatusStr)
		filter.HealthStatus = &hs
	}

	plants, err := h.service.ListGardenPlants(r.Context(), gardenID, filter)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, plants, nil)
}

// GetGardenPlant handles GET /api/v1/garden-plants/:id
func (h *PlantPlacementHandler) GetGardenPlant(w http.ResponseWriter, r *http.Request) {
	gardenPlantID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenPlantID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	gardenPlant, err := h.service.GetGardenPlant(r.Context(), gardenPlantID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, gardenPlant, nil)
}

// UpdatePlantPlacement handles PUT /api/v1/garden-plants/:id
func (h *PlantPlacementHandler) UpdatePlantPlacement(w http.ResponseWriter, r *http.Request) {
	gardenPlantID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenPlantID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	var req updatePlantPlacementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Get existing placement
	existing, err := h.service.GetGardenPlant(r.Context(), gardenPlantID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	// Update fields
	if req.LocationGeoJSON != nil {
		existing.LocationGeoJSON = *req.LocationGeoJSON
	}
	if req.ZoneID != nil {
		existing.ZoneID = req.ZoneID
	}
	if req.Quantity != nil {
		existing.Quantity = *req.Quantity
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	updated, err := h.service.UpdatePlantPlacement(r.Context(), existing)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, updated, nil)
}

// RemovePlant handles DELETE /api/v1/garden-plants/:id
func (h *PlantPlacementHandler) RemovePlant(w http.ResponseWriter, r *http.Request) {
	gardenPlantID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenPlantID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	if err := h.service.RemovePlant(r.Context(), gardenPlantID); err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondNoContent(w)
}

// Request DTOs
type placePlantRequest struct {
	PlantID        string  `json:"plant_id"`
	LocationGeoJSON string  `json:"location_geojson"`
	ZoneID         *string `json:"zone_id,omitempty"`
	Quantity       int     `json:"quantity"`
	Notes          *string `json:"notes,omitempty"`
}

type bulkPlacePlantsRequest struct {
	Plants []placePlantRequest `json:"plants"`
}

type updatePlantPlacementRequest struct {
	LocationGeoJSON *string `json:"location_geojson,omitempty"`
	ZoneID          *string `json:"zone_id,omitempty"`
	Quantity        *int    `json:"quantity,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}
