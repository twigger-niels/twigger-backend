package handlers

import (
	"database/sql"
	"net/http"

	"twigger-backend/internal/api-gateway/utils"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthCheck handles GET /health
// @Summary Health check
// @Description Check if the API service is healthy and database is connected
// @Tags health
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=map[string]string} "Service is healthy"
// @Failure 503 {object} map[string]interface{} "Service unavailable"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	if err := h.db.PingContext(r.Context()); err != nil {
		utils.RespondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}

	utils.RespondSuccess(w, map[string]string{
		"status":   "healthy",
		"database": "connected",
	}, nil)
}

// ReadinessCheck handles GET /ready
// @Summary Readiness check
// @Description Check if the API service is ready to accept requests
// @Tags health
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=map[string]bool} "Service is ready"
// @Failure 503 {object} map[string]interface{} "Service not ready"
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// Check if system is ready to accept requests
	// For now, just check database
	if err := h.db.PingContext(r.Context()); err != nil {
		utils.RespondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ready": false,
			"error": err.Error(),
		})
		return
	}

	utils.RespondSuccess(w, map[string]bool{
		"ready": true,
	}, nil)
}
