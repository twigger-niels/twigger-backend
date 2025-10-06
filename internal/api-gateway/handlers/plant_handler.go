package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"twigger-backend/backend/plant-service/domain/entity"
	plantService "twigger-backend/backend/plant-service/domain/service"
	"twigger-backend/backend/plant-service/domain/repository"
	"twigger-backend/backend/plant-service/pkg/types"
	"twigger-backend/internal/api-gateway/utils"
)

// PlantHandler handles plant-related HTTP requests
type PlantHandler struct {
	service *plantService.PlantService
}

// NewPlantHandler creates a new plant handler
func NewPlantHandler(service *plantService.PlantService) *PlantHandler {
	return &PlantHandler{
		service: service,
	}
}

// GetPlant handles GET /api/v1/plants/:id
// @Summary Get plant by ID
// @Description Get detailed information about a specific plant by its ID
// @Tags plants
// @Accept json
// @Produce json
// @Param id path string true "Plant ID (UUID)"
// @Param include_details query boolean false "Include detailed characteristics" default(false)
// @Param country_id query string false "Country ID for region-specific growing conditions"
// @Header 200 {string} Accept-Language "Language for localized plant names (e.g., 'en', 'es', 'fr')"
// @Success 200 {object} utils.SuccessResponse "Plant details"
// @Failure 400 {object} utils.ErrorResponse "Invalid plant ID format"
// @Failure 404 {object} utils.ErrorResponse "Plant not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /plants/{id} [get]
func (h *PlantHandler) GetPlant(w http.ResponseWriter, r *http.Request) {
	plantID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(plantID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	// Extract language context (TODO: pass to service when language param added)
	_ = utils.ExtractLanguageContext(r)

	// Get query parameters
	includeDetails := utils.GetQueryParamBool(r, "include_details", false)
	countryID := utils.GetQueryParam(r, "country_id")

	// Get plant with language context
	// Note: We need to update the service to accept language context
	// For now, the service uses hardcoded "en", but this will be fixed
	var plant interface{}
	var err error

	if countryID != "" {
		plant, err = h.service.GetPlantWithConditions(r.Context(), plantID, countryID)
	} else {
		plant, err = h.service.GetPlant(r.Context(), plantID, includeDetails)
	}

	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, plant, nil)
}

// SearchPlants handles GET /api/v1/plants/search
// @Summary Search plants
// @Description Search plants by name or botanical name with optional filters
// @Tags plants
// @Accept json
// @Produce json
// @Param q query string false "Search query (matches common names and botanical names)"
// @Param limit query integer false "Maximum number of results" default(20) maximum(100)
// @Param cursor query string false "Pagination cursor from previous response"
// @Param min_height query number false "Minimum mature height in meters"
// @Param max_height query number false "Maximum mature height in meters"
// @Param growth_rate query string false "Growth rate (slow, medium, fast)" Enums(slow, medium, fast)
// @Param hardiness_zone query string false "USDA hardiness zone (e.g., '5a', '6b')"
// @Param sun_requirement query string false "Sun requirement" Enums(full_sun, partial_sun, partial_shade, full_shade)
// @Param water_needs query string false "Water needs" Enums(low, medium, high)
// @Param evergreen query boolean false "Filter for evergreen plants"
// @Param deciduous query boolean false "Filter for deciduous plants"
// @Param toxic query boolean false "Filter toxic/non-toxic plants"
// @Header 200 {string} Accept-Language "Language for localized plant names"
// @Success 200 {object} utils.SuccessResponse{data=[]entity.Plant,meta=utils.Meta} "List of matching plants"
// @Failure 400 {object} utils.ErrorResponse "Invalid filter parameters"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /plants/search [get]
func (h *PlantHandler) SearchPlants(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	query := utils.GetQueryParam(r, "q")
	limit := utils.GetQueryParamInt(r, "limit", 20)
	limit = utils.ValidateLimit(limit, 100)
	cursor := utils.GetQueryParam(r, "cursor")

	// Build search filter
	filter := &repository.SearchFilter{
		Limit:  limit,
		Cursor: nil,
	}
	if cursor != "" {
		filter.Cursor = &cursor
	}

	// Physical characteristic filters
	if minHeight := utils.GetQueryParam(r, "min_height"); minHeight != "" {
		if height, err := utils.ParseFloat64(minHeight); err == nil {
			filter.MinHeight = &height
		}
	}
	if maxHeight := utils.GetQueryParam(r, "max_height"); maxHeight != "" {
		if height, err := utils.ParseFloat64(maxHeight); err == nil {
			filter.MaxHeight = &height
		}
	}
	if growthRate := utils.GetQueryParam(r, "growth_rate"); growthRate != "" {
		rate := types.GrowthRate(growthRate)
		filter.GrowthRate = &rate
	}
	if evergreen := utils.GetQueryParam(r, "evergreen"); evergreen != "" {
		filter.Evergreen = utils.StringToBoolPtr(evergreen)
	}
	if deciduous := utils.GetQueryParam(r, "deciduous"); deciduous != "" {
		filter.Deciduous = utils.StringToBoolPtr(deciduous)
	}
	if toxic := utils.GetQueryParam(r, "toxic"); toxic != "" {
		filter.Toxic = utils.StringToBoolPtr(toxic)
	}

	// Extract language context (TODO: pass to service when language param added)
	_ = utils.ExtractLanguageContext(r)

	// Perform search
	result, err := h.service.SearchPlants(r.Context(), query, filter)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	// Build metadata
	meta := &utils.Meta{
		Cursor:  result.NextCursor,
		HasMore: result.HasMore,
		Limit:   limit,
	}

	utils.RespondSuccess(w, result.Plants, meta)
}

// GetCompanions handles GET /api/v1/plants/:id/companions
// @Summary Get companion plants
// @Description Get list of companion plants (beneficial or antagonistic) for a given plant
// @Tags plants
// @Accept json
// @Produce json
// @Param id path string true "Plant ID (UUID)"
// @Param beneficial_only query boolean false "Only return beneficial companions" default(false)
// @Success 200 {object} utils.SuccessResponse "List of companion plants with relationship details"
// @Failure 400 {object} utils.ErrorResponse "Invalid plant ID"
// @Failure 404 {object} utils.ErrorResponse "Plant not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /plants/{id}/companions [get]
func (h *PlantHandler) GetCompanions(w http.ResponseWriter, r *http.Request) {
	plantID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(plantID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	// Check query params
	beneficialOnly := utils.GetQueryParamBool(r, "beneficial_only", false)

	// Extract language context (TODO: pass to service when language param added)
	_ = utils.ExtractLanguageContext(r)

	// Get companions
	companions, err := h.service.GetCompanionPlants(r.Context(), plantID, beneficialOnly)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, companions, nil)
}

// FindByFamily handles GET /api/v1/plants/family/:name
// @Summary Find plants by family
// @Description Get all plants belonging to a specific plant family
// @Tags plants
// @Accept json
// @Produce json
// @Param name path string true "Family name (e.g., 'Rosaceae', 'Solanaceae')"
// @Param limit query integer false "Maximum number of results" default(20) maximum(100)
// @Param offset query integer false "Number of results to skip" default(0)
// @Success 200 {object} utils.SuccessResponse "List of plants in the family"
// @Failure 400 {object} utils.ErrorResponse "Invalid family name"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /plants/family/{name} [get]
func (h *PlantHandler) FindByFamily(w http.ResponseWriter, r *http.Request) {
	familyName := utils.GetPathParam(r, "name")
	if familyName == "" {
		utils.RespondValidationError(w, "name", "Family name is required")
		return
	}

	limit := utils.GetQueryParamInt(r, "limit", 20)
	limit = utils.ValidateLimit(limit, 100)
	offset := utils.GetQueryParamInt(r, "offset", 0)

	// Extract language context (TODO: pass to service when language param added)
	_ = utils.ExtractLanguageContext(r)

	plants, err := h.service.FindPlantsByFamily(r.Context(), familyName, limit, offset)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, plants, nil)
}

// FindByGenus handles GET /api/v1/plants/genus/:name
// @Summary Find plants by genus
// @Description Get all plants belonging to a specific genus
// @Tags plants
// @Accept json
// @Produce json
// @Param name path string true "Genus name (e.g., 'Rosa', 'Lavandula')"
// @Param limit query integer false "Maximum number of results" default(20) maximum(100)
// @Param offset query integer false "Number of results to skip" default(0)
// @Success 200 {object} utils.SuccessResponse "List of plants in the genus"
// @Failure 400 {object} utils.ErrorResponse "Invalid genus name"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /plants/genus/{name} [get]
func (h *PlantHandler) FindByGenus(w http.ResponseWriter, r *http.Request) {
	genusName := utils.GetPathParam(r, "name")
	if genusName == "" {
		utils.RespondValidationError(w, "name", "Genus name is required")
		return
	}

	limit := utils.GetQueryParamInt(r, "limit", 20)
	limit = utils.ValidateLimit(limit, 100)
	offset := utils.GetQueryParamInt(r, "offset", 0)

	// Extract language context (TODO: pass to service when language param added)
	_ = utils.ExtractLanguageContext(r)

	plants, err := h.service.FindPlantsByGenus(r.Context(), genusName, limit, offset)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, plants, nil)
}

// RecommendPlants handles GET /api/v1/plants/recommend
// @Summary Get plant recommendations
// @Description Get recommended plants based on hardiness zone and sun requirements
// @Tags plants
// @Accept json
// @Produce json
// @Param hardiness_zone query string true "USDA hardiness zone (e.g., '5a', '6b', '7a')"
// @Param sun_requirement query string false "Sun requirement" Enums(full_sun, partial_sun, partial_shade, full_shade) default(full_sun)
// @Param limit query integer false "Maximum number of recommendations" default(10) maximum(50)
// @Success 200 {object} utils.SuccessResponse "List of recommended plants"
// @Failure 400 {object} utils.ErrorResponse "Invalid parameters"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /plants/recommend [get]
func (h *PlantHandler) RecommendPlants(w http.ResponseWriter, r *http.Request) {
	hardinessZone := utils.GetQueryParam(r, "hardiness_zone")
	if hardinessZone == "" {
		utils.RespondValidationError(w, "hardiness_zone", "Hardiness zone is required")
		return
	}

	sunReq := utils.GetQueryParam(r, "sun_requirement")
	if sunReq == "" {
		sunReq = string(types.SunFullSun) // default to full sun
	}

	limit := utils.GetQueryParamInt(r, "limit", 10)
	limit = utils.ValidateLimit(limit, 50)

	// Extract language context (TODO: pass to service when language param added)
	_ = utils.ExtractLanguageContext(r)

	plants, err := h.service.RecommendPlants(r.Context(), hardinessZone, types.SunRequirement(sunReq), limit)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	utils.RespondSuccess(w, plants, nil)
}

// CreatePlant handles POST /api/v1/plants (admin only)
// @Summary Create a new plant (Admin only)
// @Description Create a new plant entry in the database
// @Tags plants,admin
// @Accept json
// @Produce json
// @Param plant body createPlantRequest true "Plant creation data"
// @Security Bearer
// @Success 201 {object} utils.SuccessResponse "Plant created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized - missing or invalid token"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - admin role required"
// @Failure 409 {object} utils.ErrorResponse "Conflict - plant already exists"
// @Router /plants [post]
func (h *PlantHandler) CreatePlant(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req createPlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Validate required fields
	if req.FullBotanicalName == "" {
		utils.RespondValidationError(w, "full_botanical_name", "Botanical name is required")
		return
	}
	if req.SpeciesID == "" {
		utils.RespondValidationError(w, "species_id", "Species ID is required")
		return
	}
	if req.PlantType == "" {
		utils.RespondValidationError(w, "plant_type", "Plant type is required")
		return
	}

	// Validate UUIDs
	if err := utils.ValidateUUID(req.SpeciesID); err != nil {
		utils.RespondValidationError(w, "species_id", "Invalid species ID format")
		return
	}

	// Build plant entity
	plantID := uuid.New().String() // Generate new UUID
	plant := &entity.Plant{
		PlantID:           plantID,
		SpeciesID:         req.SpeciesID,
		FullBotanicalName: req.FullBotanicalName,
		PlantType:         types.PlantType(req.PlantType),
		// Note: Taxonomy info (FamilyName, GenusName, SpeciesName) will be populated
		// by the repository when it joins with plant_species table
	}

	// Set cultivar ID if provided
	if req.CultivarID != nil {
		if err := utils.ValidateUUID(*req.CultivarID); err != nil {
			utils.RespondValidationError(w, "cultivar_id", "Invalid cultivar ID format")
			return
		}
		plant.CultivarID = req.CultivarID
	}

	// Create plant via service
	if err := h.service.CreatePlant(r.Context(), plant); err != nil {
		utils.RespondError(w, err)
		return
	}

	// Return created plant
	// TODO: Extract language from context (Part 6)
	created, err := h.service.GetPlant(r.Context(), plantID, false)
	if err != nil {
		// Plant was created but couldn't retrieve it
		utils.RespondJSON(w, http.StatusCreated, map[string]string{
			"plant_id": plantID,
			"message":  "Plant created successfully",
		})
		return
	}

	utils.RespondCreated(w, created)
}

// UpdatePlant handles PUT /api/v1/plants/:id (admin only)
// @Summary Update an existing plant (Admin only)
// @Description Update plant information
// @Tags plants,admin
// @Accept json
// @Produce json
// @Param id path string true "Plant ID (UUID)"
// @Param plant body updatePlantRequest true "Plant update data"
// @Security Bearer
// @Success 200 {object} utils.SuccessResponse "Plant updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Plant not found"
// @Failure 409 {object} utils.ErrorResponse "Conflict - botanical name already exists"
// @Router /plants/{id} [put]
func (h *PlantHandler) UpdatePlant(w http.ResponseWriter, r *http.Request) {
	plantID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(plantID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	// Decode request body
	var req updatePlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondValidationError(w, "body", "Invalid request body")
		return
	}

	// Validate that at least one field is being updated
	if req.FullBotanicalName == "" {
		utils.RespondValidationError(w, "full_botanical_name", "At least one field must be provided for update")
		return
	}

	// Build plant entity with updates
	// Get existing plant first to preserve required fields
	// TODO: Extract language from context (Part 6)
	existing, err := h.service.GetPlant(r.Context(), plantID, false)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	// Apply updates to existing plant
	updates := &entity.Plant{
		PlantID:           plantID,
		SpeciesID:         existing.SpeciesID, // Immutable
		FullBotanicalName: req.FullBotanicalName,
		FamilyName:        existing.FamilyName,
		GenusName:         existing.GenusName,
		SpeciesName:       existing.SpeciesName,
		PlantType:         existing.PlantType,
		CreatedAt:         existing.CreatedAt,
	}

	// Update plant via service
	if err := h.service.UpdatePlant(r.Context(), plantID, updates); err != nil {
		utils.RespondError(w, err)
		return
	}

	// Return updated plant
	updated, err := h.service.GetPlant(r.Context(), plantID, false)
	if err != nil {
		// Plant was updated but couldn't retrieve it
		utils.RespondJSON(w, http.StatusOK, map[string]string{
			"plant_id": plantID,
			"message":  "Plant updated successfully",
		})
		return
	}

	utils.RespondSuccess(w, updated, nil)
}

// DeletePlant handles DELETE /api/v1/plants/:id (admin only)
// @Summary Delete a plant (Admin only)
// @Description Delete a plant from the database
// @Tags plants,admin
// @Accept json
// @Produce json
// @Param id path string true "Plant ID (UUID)"
// @Security Bearer
// @Success 204 "Plant deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid plant ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Plant not found"
// @Router /plants/{id} [delete]
func (h *PlantHandler) DeletePlant(w http.ResponseWriter, r *http.Request) {
	plantID := utils.GetPathParam(r, "id")
	if err := utils.ValidateUUID(plantID); err != nil {
		utils.RespondValidationError(w, "id", err.Error())
		return
	}

	// Delete plant via service
	if err := h.service.DeletePlant(r.Context(), plantID); err != nil {
		utils.RespondError(w, err)
		return
	}

	// Return 204 No Content on successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// Request DTOs

// createPlantRequest represents the request body for creating a plant
// swagger:model
type createPlantRequest struct {
	FullBotanicalName string  `json:"full_botanical_name" example:"Rosa rugosa"`
	FamilyID          string  `json:"family_id" example:"f50e8400-e29b-41d4-a716-446655440001"`
	GenusID           string  `json:"genus_id" example:"150e8400-e29b-41d4-a716-446655440001"`
	SpeciesID         string  `json:"species_id" example:"250e8400-e29b-41d4-a716-446655440001"`
	CultivarID        *string `json:"cultivar_id,omitempty"`
	PlantType         string  `json:"plant_type" example:"shrub" enums:"annual,perennial,biennial,shrub,tree,vine"`
}

// updatePlantRequest represents the request body for updating a plant
// swagger:model
type updatePlantRequest struct {
	FullBotanicalName string `json:"full_botanical_name,omitempty" example:"Rosa rugosa 'Alba'"`
}
