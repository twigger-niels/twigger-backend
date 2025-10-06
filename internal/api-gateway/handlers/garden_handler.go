package handlers

import (
	"encoding/json"
	"net/http"

	gardenService "twigger-backend/backend/garden-service/domain/service"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/internal/api-gateway/utils"
)

// GardenHandler handles garden-related HTTP requests
type GardenHandler struct {
	service gardenService.GardenService
}

// NewGardenHandler creates a new garden handler
func NewGardenHandler(service gardenService.GardenService) *GardenHandler {
	return &GardenHandler{
		service: service,
	}
}

// CreateGarden handles POST /api/v1/gardens
func (h *GardenHandler) CreateGarden(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := utils.GetUserID(r.Context())
	if userID == "" {
		utils.RespondUnauthorized(w, "User not authenticated")
		return
	}

	// Decode request body
	var req createGardenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Build garden entity
	var aspect *entity.Aspect
	if req.Aspect != nil {
		a := entity.Aspect(*req.Aspect)
		aspect = &a
	}

	garden := &entity.Garden{
		UserID:          userID,
		GardenName:      req.Name,
		LocationGeoJSON: req.LocationGeoJSON,
		BoundaryGeoJSON: req.BoundaryGeoJSON,
		Aspect:          aspect,
		SlopeDegrees:    req.SlopeDegrees,
		ElevationM:      req.ElevationM,
	}

	// Create garden
	created, err := h.service.CreateGarden(r.Context(), garden)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondCreated(w, created)
}

// GetGarden handles GET /api/v1/gardens/:id
func (h *GardenHandler) GetGarden(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	garden, err := h.service.GetGarden(r.Context(), gardenID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, garden, nil)
}

// ListGardens handles GET /api/v1/gardens
func (h *GardenHandler) ListGardens(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := utils.GetUserID(r.Context())
	if userID == "" {
		utils.RespondUnauthorized(w, "User not authenticated")
		return
	}

	limit := utils.GetQueryParamInt(r, "limit", 10)
	limit = utils.ValidateLimit(limit, 100)
	offset := utils.GetQueryParamInt(r, "offset", 0)

	gardens, err := h.service.ListUserGardens(r.Context(), userID, limit, offset)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, gardens, nil)
}

// UpdateGarden handles PUT /api/v1/gardens/:id
func (h *GardenHandler) UpdateGarden(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	// Decode request body
	var req updateGardenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Get existing garden to ensure it exists and user owns it
	existing, err := h.service.GetGarden(r.Context(), gardenID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	// Verify user owns this garden
	userID := utils.GetUserID(r.Context())
	if existing.UserID != userID {
		utils.RespondForbidden(w, "You don't have permission to update this garden")
		return
	}

	// Update fields
	if req.Name != nil {
		existing.GardenName = *req.Name
	}
	if req.BoundaryGeoJSON != nil {
		existing.BoundaryGeoJSON = req.BoundaryGeoJSON
	}
	if req.LocationGeoJSON != nil {
		existing.LocationGeoJSON = req.LocationGeoJSON
	}
	if req.Aspect != nil {
		a := entity.Aspect(*req.Aspect)
		existing.Aspect = &a
	}
	if req.SlopeDegrees != nil {
		existing.SlopeDegrees = req.SlopeDegrees
	}
	if req.ElevationM != nil {
		existing.ElevationM = req.ElevationM
	}

	// Update garden
	updated, err := h.service.UpdateGarden(r.Context(), existing)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, updated, nil)
}

// DeleteGarden handles DELETE /api/v1/gardens/:id
func (h *GardenHandler) DeleteGarden(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	// Verify ownership
	garden, err := h.service.GetGarden(r.Context(), gardenID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	userID := utils.GetUserID(r.Context())
	if garden.UserID != userID {
		utils.RespondForbidden(w, "You don't have permission to delete this garden")
		return
	}

	// Delete garden
	if err := h.service.DeleteGarden(r.Context(), gardenID); err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondNoContent(w)
}

// GetGardenStats handles GET /api/v1/gardens/stats
func (h *GardenHandler) GetGardenStats(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r.Context())
	if userID == "" {
		utils.RespondUnauthorized(w, "User not authenticated")
		return
	}

	stats, err := h.service.GetGardenStats(r.Context(), userID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, stats, nil)
}

// FindNearbyGardens handles GET /api/v1/gardens/nearby
func (h *GardenHandler) FindNearbyGardens(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	latStr := utils.GetQueryParam(r, "lat")
	lngStr := utils.GetQueryParam(r, "lng")
	radiusStr := utils.GetQueryParam(r, "radius_km")

	if latStr == "" || lngStr == "" {
		utils.RespondValidationError(w, "lat,lng", "Latitude and longitude are required")
		return
	}

	lat, err := utils.ParseFloat64(latStr)
	if err != nil {
		utils.RespondValidationError(w, "lat", "Invalid latitude")
		return
	}

	lng, err := utils.ParseFloat64(lngStr)
	if err != nil {
		utils.RespondValidationError(w, "lng", "Invalid longitude")
		return
	}

	if err := utils.ValidateCoordinates(lat, lng); err != nil {
		utils.RespondValidationError(w, "coordinates", err.Error())
		return
	}

	radiusKm := utils.ParseFloat64OrDefault(radiusStr, 10.0) // default 10km
	if radiusKm <= 0 || radiusKm > 100 {
		radiusKm = 10.0 // cap at 100km
	}

	gardens, err := h.service.FindNearbyGardens(r.Context(), lat, lng, radiusKm)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, gardens, nil)
}

// Request DTOs
type createGardenRequest struct {
	Name            string   `json:"name"`
	LocationGeoJSON *string  `json:"location_geojson,omitempty"`
	BoundaryGeoJSON *string  `json:"boundary_geojson,omitempty"`
	Aspect          *string  `json:"aspect,omitempty"`          // "N", "NE", "E", etc.
	SlopeDegrees    *float64 `json:"slope_degrees,omitempty"`   // 0-90 degrees
	ElevationM      *float64 `json:"elevation_m,omitempty"`
}

type updateGardenRequest struct {
	Name            *string  `json:"name,omitempty"`
	LocationGeoJSON *string  `json:"location_geojson,omitempty"`
	BoundaryGeoJSON *string  `json:"boundary_geojson,omitempty"`
	Aspect          *string  `json:"aspect,omitempty"`          // "N", "NE", "E", etc.
	SlopeDegrees    *float64 `json:"slope_degrees,omitempty"`   // 0-90 degrees
	ElevationM      *float64 `json:"elevation_m,omitempty"`
}
