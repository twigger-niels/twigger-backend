package handlers

import (
	"encoding/json"
	"net/http"

	gardenService "twigger-backend/backend/garden-service/domain/service"
	"twigger-backend/backend/garden-service/domain/entity"
	"twigger-backend/internal/api-gateway/utils"
)

// ZoneHandler handles garden zone-related HTTP requests
type ZoneHandler struct {
	service gardenService.ZoneManagementService
}

// NewZoneHandler creates a new zone handler
func NewZoneHandler(service gardenService.ZoneManagementService) *ZoneHandler {
	return &ZoneHandler{
		service: service,
	}
}

// CreateZone handles POST /api/v1/gardens/:id/zones
func (h *ZoneHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	var req createZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Convert string to ZoneType enum
	var zoneType *entity.ZoneType
	if req.ZoneType != "" {
		zt := entity.ZoneType(req.ZoneType)
		zoneType = &zt
	}

	// Convert string to IrrigationType enum
	var irrigationType *entity.IrrigationType
	if req.IrrigationType != nil {
		it := entity.IrrigationType(*req.IrrigationType)
		irrigationType = &it
	}

	zone := &entity.GardenZone{
		GardenID:        gardenID,
		ZoneName:        &req.Name,
		ZoneType:        zoneType,
		GeometryGeoJSON: req.GeometryGeoJSON,
		IrrigationType:  irrigationType,
		SunHoursSummer:  req.SunHoursPerDay,
	}

	created, err := h.service.CreateZone(r.Context(), zone)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondCreated(w, created)
}

// ListGardenZones handles GET /api/v1/gardens/:id/zones
func (h *ZoneHandler) ListGardenZones(w http.ResponseWriter, r *http.Request) {
	gardenID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(gardenID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	zones, err := h.service.ListGardenZones(r.Context(), gardenID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, zones, nil)
}

// GetZone handles GET /api/v1/zones/:id
func (h *ZoneHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	zoneID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(zoneID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	zone, err := h.service.GetZone(r.Context(), zoneID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, zone, nil)
}

// UpdateZone handles PUT /api/v1/zones/:id
func (h *ZoneHandler) UpdateZone(w http.ResponseWriter, r *http.Request) {
	zoneID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(zoneID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	var req updateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Get existing zone
	existing, err := h.service.GetZone(r.Context(), zoneID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	// Update fields
	if req.Name != nil {
		existing.ZoneName = req.Name
	}
	if req.GeometryGeoJSON != nil {
		existing.GeometryGeoJSON = *req.GeometryGeoJSON
	}
	if req.IrrigationType != nil {
		it := entity.IrrigationType(*req.IrrigationType)
		existing.IrrigationType = &it
	}
	if req.SunHoursPerDay != nil {
		existing.SunHoursSummer = req.SunHoursPerDay
	}

	updated, err := h.service.UpdateZone(r.Context(), existing)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, updated, nil)
}

// DeleteZone handles DELETE /api/v1/zones/:id
func (h *ZoneHandler) DeleteZone(w http.ResponseWriter, r *http.Request) {
	zoneID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(zoneID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	if err := h.service.DeleteZone(r.Context(), zoneID); err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondNoContent(w)
}

// CalculateZoneArea handles GET /api/v1/zones/:id/area
func (h *ZoneHandler) CalculateZoneArea(w http.ResponseWriter, r *http.Request) {
	zoneID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(zoneID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	area, err := h.service.CalculateZoneArea(r.Context(), zoneID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, map[string]float64{"area_m2": area}, nil)
}

// Request DTOs
type createZoneRequest struct {
	Name            string  `json:"name"`
	ZoneType        string  `json:"zone_type"`
	GeometryGeoJSON string  `json:"geometry_geojson"`
	IrrigationType  *string `json:"irrigation_type,omitempty"`
	SunHoursPerDay  *int    `json:"sun_hours_per_day,omitempty"` // 0-24 hours
}

type updateZoneRequest struct {
	Name            *string `json:"name,omitempty"`
	GeometryGeoJSON *string `json:"geometry_geojson,omitempty"`
	IrrigationType  *string `json:"irrigation_type,omitempty"`
	SunHoursPerDay  *int    `json:"sun_hours_per_day,omitempty"` // 0-24 hours
}
