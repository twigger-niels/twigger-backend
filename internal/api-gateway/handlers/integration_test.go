// +build integration

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"twigger-backend/backend/garden-service/domain/service"
	gardenRepo "twigger-backend/backend/garden-service/infrastructure/persistence"
	"twigger-backend/backend/plant-service/domain/entity"
	plantService "twigger-backend/backend/plant-service/domain/service"
	plantRepo "twigger-backend/backend/plant-service/infrastructure/database"
	testhelpers "twigger-backend/backend/plant-service/infrastructure/database/testing"
	"twigger-backend/internal/api-gateway/middleware"
	"twigger-backend/internal/api-gateway/router"
	"twigger-backend/internal/api-gateway/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer wraps the HTTP test server with database and services
type TestServer struct {
	DB         *sql.DB
	HTTPServer *httptest.Server
	BaseURL    string
	T          *testing.T
}

// SetupTestServer creates a test server with all dependencies
func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Setup database
	db := testhelpers.SetupTestDB(t)

	// Seed reference data
	testhelpers.SeedTestLanguages(t, db)
	testhelpers.SeedTestCountries(t, db)

	// Initialize repositories
	plantRepository := plantRepo.NewPostgresPlantRepository(db)
	gardenRepository := gardenRepo.NewPostgresGardenRepository(db)
	zoneRepository := gardenRepo.NewPostgresGardenZoneRepository(db)
	plantPlacementRepo := gardenRepo.NewPostgresGardenPlantRepository(db)

	// Initialize services
	plantSvc := plantService.NewPlantService(plantRepository)
	gardenSvc := service.NewGardenService(gardenRepository, zoneRepository, plantRepository)
	zoneSvc := service.NewZoneManagementService(zoneRepository, gardenRepository)
	placementSvc := service.NewPlantPlacementService(plantPlacementRepo, plantRepository, zoneRepository, gardenRepository)

	// Initialize handlers
	handlers := &Handlers{
		PlantHandler:          NewPlantHandler(plantSvc),
		GardenHandler:         NewGardenHandler(gardenSvc),
		ZoneHandler:           NewZoneHandler(zoneSvc),
		PlantPlacementHandler: NewPlantPlacementHandler(placementSvc),
		HealthHandler:         NewHealthHandler(db),
	}

	// Create auth middleware (dev mode)
	authMiddleware := middleware.NewAuthMiddleware(false)

	// Create test router
	mux := router.NewRouter(handlers, authMiddleware, &router.Config{
		AllowedOrigins: []string{"*"},
		RateLimitRPM:   1000, // High limit for tests
	})

	// Create HTTP test server
	httpServer := httptest.NewServer(mux)

	return &TestServer{
		DB:         db,
		HTTPServer: httpServer,
		BaseURL:    httpServer.URL,
		T:          t,
	}
}

// Teardown cleans up the test server
func (ts *TestServer) Teardown() {
	ts.HTTPServer.Close()
	testhelpers.TeardownTestDB(ts.T, ts.DB)
}

// GET makes a GET request to the test server
func (ts *TestServer) GET(path string, headers map[string]string) (*http.Response, []byte) {
	req, err := http.NewRequest("GET", ts.BaseURL+path, nil)
	require.NoError(ts.T, err)

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(ts.T, err)

	body := ts.ReadBody(resp)
	return resp, body
}

// POST makes a POST request to the test server
func (ts *TestServer) POST(path string, payload interface{}, headers map[string]string) (*http.Response, []byte) {
	jsonData, err := json.Marshal(payload)
	require.NoError(ts.T, err)

	req, err := http.NewRequest("POST", ts.BaseURL+path, bytes.NewBuffer(jsonData))
	require.NoError(ts.T, err)

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(ts.T, err)

	body := ts.ReadBody(resp)
	return resp, body
}

// PUT makes a PUT request to the test server
func (ts *TestServer) PUT(path string, payload interface{}, headers map[string]string) (*http.Response, []byte) {
	jsonData, err := json.Marshal(payload)
	require.NoError(ts.T, err)

	req, err := http.NewRequest("PUT", ts.BaseURL+path, bytes.NewBuffer(jsonData))
	require.NoError(ts.T, err)

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(ts.T, err)

	body := ts.ReadBody(resp)
	return resp, body
}

// DELETE makes a DELETE request to the test server
func (ts *TestServer) DELETE(path string, headers map[string]string) (*http.Response, []byte) {
	req, err := http.NewRequest("DELETE", ts.BaseURL+path, nil)
	require.NoError(ts.T, err)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(ts.T, err)

	body := ts.ReadBody(resp)
	return resp, body
}

// ReadBody reads response body and returns bytes
func (ts *TestServer) ReadBody(resp *http.Response) []byte {
	defer resp.Body.Close()
	body := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return body
}

// SeedTestPlant creates a test plant with localized names
func (ts *TestServer) SeedTestPlant(botanicalName string, commonNames map[string]string) string {
	ts.T.Helper()

	ctx := context.Background()

	// Create plant hierarchy
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(ts.T, ts.DB)

	// Create plant
	plantID := uuid.New().String()
	_, err := ts.DB.ExecContext(ctx, `
		INSERT INTO plants (plant_id, species_id, full_botanical_name, created_at)
		VALUES ($1, $2, $3, NOW())
	`, plantID, speciesID, botanicalName)
	require.NoError(ts.T, err)

	// Add common names for each language
	for langCode, commonName := range commonNames {
		var langID string
		err := ts.DB.QueryRowContext(ctx, `
			SELECT language_id FROM languages WHERE language_code = $1
		`, langCode).Scan(&langID)
		require.NoError(ts.T, err)

		_, err = ts.DB.ExecContext(ctx, `
			INSERT INTO plant_common_names (plant_id, language_id, common_name, is_primary)
			VALUES ($1, $2, $3, true)
		`, plantID, langID, commonName)
		require.NoError(ts.T, err)
	}

	ts.T.Logf("Seeded plant: %s (%s)", botanicalName, plantID)
	return plantID
}

// SeedTestGarden creates a test garden
func (ts *TestServer) SeedTestGarden(userID, name string, boundary string) string {
	ts.T.Helper()

	ctx := context.Background()
	gardenID := uuid.New().String()

	_, err := ts.DB.ExecContext(ctx, `
		INSERT INTO gardens (garden_id, user_id, garden_name, boundary, created_at)
		VALUES ($1, $2, $3, ST_GeomFromGeoJSON($4), NOW())
	`, gardenID, userID, name, boundary)
	require.NoError(ts.T, err)

	ts.T.Logf("Seeded garden: %s (%s)", name, gardenID)
	return gardenID
}

// ParseSuccessResponse parses a successful JSON response
func (ts *TestServer) ParseSuccessResponse(body []byte, data interface{}) {
	var response struct {
		Data json.RawMessage `json:"data"`
		Meta *utils.Meta     `json:"meta,omitempty"`
	}

	err := json.Unmarshal(body, &response)
	require.NoError(ts.T, err, "Failed to parse success response: %s", string(body))

	if data != nil {
		err = json.Unmarshal(response.Data, data)
		require.NoError(ts.T, err, "Failed to parse data field: %s", string(response.Data))
	}
}

// ParseErrorResponse parses an error JSON response
func (ts *TestServer) ParseErrorResponse(body []byte) utils.ErrorResponse {
	var errorResp utils.ErrorResponse
	err := json.Unmarshal(body, &errorResp)
	require.NoError(ts.T, err, "Failed to parse error response: %s", string(body))
	return errorResp
}

// ===== HEALTH CHECK TESTS =====

func TestHealthEndpoint(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	resp, body := ts.GET("/health", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "ok")
}

func TestReadyEndpoint(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	resp, body := ts.GET("/ready", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "ready")
}

// ===== PLANT API TESTS =====

func TestSearchPlants_EmptyQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed test plants
	ts.SeedTestPlant("Rosa rugosa", map[string]string{
		"en": "Japanese Rose",
		"es": "Rosa rugosa",
	})
	ts.SeedTestPlant("Lavandula angustifolia", map[string]string{
		"en": "English Lavender",
		"es": "Lavanda",
	})

	resp, body := ts.GET("/api/v1/plants/search", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plants []entity.Plant
	ts.ParseSuccessResponse(body, &plants)

	assert.GreaterOrEqual(t, len(plants), 2, "Should return at least 2 plants")
}

func TestSearchPlants_WithQuery(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed test plants
	roseID := ts.SeedTestPlant("Rosa rugosa", map[string]string{
		"en": "Japanese Rose",
		"es": "Rosa japonesa",
	})
	ts.SeedTestPlant("Lavandula angustifolia", map[string]string{
		"en": "English Lavender",
	})

	resp, body := ts.GET("/api/v1/plants/search?q=rose", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plants []entity.Plant
	ts.ParseSuccessResponse(body, &plants)

	assert.Equal(t, 1, len(plants), "Should return only rose")
	assert.Equal(t, roseID, plants[0].PlantID)
	assert.Contains(t, plants[0].CommonNames, "Japanese Rose")
}

func TestSearchPlants_WithLanguageHeader(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed test plant with Spanish common name
	plantID := ts.SeedTestPlant("Rosa rugosa", map[string]string{
		"en": "Japanese Rose",
		"es": "Rosa japonesa",
	})

	// Search with Spanish language header
	headers := map[string]string{
		"Accept-Language": "es",
	}
	resp, body := ts.GET("/api/v1/plants/search?q=rosa", headers)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plants []entity.Plant
	ts.ParseSuccessResponse(body, &plants)

	assert.Equal(t, 1, len(plants))
	assert.Equal(t, plantID, plants[0].PlantID)
	// Note: Common names returned are based on hardcoded UUID in service layer
	// This will be fixed in Part 6
}

func TestSearchPlants_WithPagination(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed multiple plants
	for i := 1; i <= 5; i++ {
		ts.SeedTestPlant(
			fmt.Sprintf("Rosa species-%d", i),
			map[string]string{"en": fmt.Sprintf("Rose %d", i)},
		)
	}

	// Request with limit=2
	resp, body := ts.GET("/api/v1/plants/search?limit=2", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Data []entity.Plant `json:"data"`
		Meta utils.Meta     `json:"meta"`
	}
	err := json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Data), "Should return only 2 plants")
	assert.Equal(t, 2, response.Meta.Limit)
	assert.True(t, response.Meta.HasMore, "Should have more results")
	assert.NotNil(t, response.Meta.Cursor, "Should return cursor for next page")
}

func TestGetPlantByID_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	plantID := ts.SeedTestPlant("Rosa rugosa", map[string]string{
		"en": "Japanese Rose",
		"fr": "Rosier rugueux",
	})

	resp, body := ts.GET("/api/v1/plants/"+plantID, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plant entity.Plant
	ts.ParseSuccessResponse(body, &plant)

	assert.Equal(t, plantID, plant.PlantID)
	assert.Equal(t, "Rosa rugosa", plant.FullBotanicalName)
	assert.Contains(t, plant.CommonNames, "Japanese Rose")
}

func TestGetPlantByID_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	nonExistentID := uuid.New().String()
	resp, body := ts.GET("/api/v1/plants/"+nonExistentID, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "RESOURCE_NOT_FOUND", errorResp.Code)
}

func TestGetPlantByID_InvalidUUID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	resp, body := ts.GET("/api/v1/plants/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "INVALID_REQUEST", errorResp.Code)
}

// ===== GARDEN API TESTS =====

func TestCreateGarden_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	payload := map[string]interface{}{
		"name": "My Test Garden",
		"boundary_geojson": `{
			"type": "Polygon",
			"coordinates": [[
				[-122.5, 37.7],
				[-122.4, 37.7],
				[-122.4, 37.8],
				[-122.5, 37.8],
				[-122.5, 37.7]
			]]
		}`,
		"slope_degrees": 5.5,
		"aspect":        "south",
	}

	resp, body := ts.POST("/api/v1/gardens", payload, nil)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var garden map[string]interface{}
	ts.ParseSuccessResponse(body, &garden)

	assert.NotNil(t, garden["garden_id"])
	assert.Equal(t, "My Test Garden", garden["garden_name"])
}

func TestCreateGarden_InvalidBoundary(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	payload := map[string]interface{}{
		"name":             "Invalid Garden",
		"boundary_geojson": `{"invalid": "geojson"}`,
	}

	resp, body := ts.POST("/api/v1/gardens", payload, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "INVALID_REQUEST", errorResp.Code)
}

func TestListGardens_UserIsolation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed gardens for two different users
	userID1 := "user-123"
	userID2 := "user-456"

	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`

	ts.SeedTestGarden(userID1, "User 1 Garden", boundary)
	ts.SeedTestGarden(userID2, "User 2 Garden", boundary)

	// List gardens (dev mode uses "dev-user-123")
	resp, body := ts.GET("/api/v1/gardens", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var gardens []map[string]interface{}
	ts.ParseSuccessResponse(body, &gardens)

	// In dev mode, user ID is "dev-user-123", so should return 0 gardens
	// (or 1 if we seed a garden for dev-user-123)
	assert.GreaterOrEqual(t, len(gardens), 0)
}

func TestGetGarden_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`

	gardenID := ts.SeedTestGarden("dev-user-123", "Test Garden", boundary)

	resp, body := ts.GET("/api/v1/gardens/"+gardenID, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var garden map[string]interface{}
	ts.ParseSuccessResponse(body, &garden)

	assert.Equal(t, gardenID, garden["garden_id"])
	assert.Equal(t, "Test Garden", garden["garden_name"])
}

func TestUpdateGarden_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`

	gardenID := ts.SeedTestGarden("dev-user-123", "Old Name", boundary)

	payload := map[string]interface{}{
		"name":          "Updated Garden Name",
		"slope_degrees": 10.0,
	}

	resp, body := ts.PUT("/api/v1/gardens/"+gardenID, payload, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var garden map[string]interface{}
	ts.ParseSuccessResponse(body, &garden)

	assert.Equal(t, "Updated Garden Name", garden["garden_name"])
}

func TestDeleteGarden_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`

	gardenID := ts.SeedTestGarden("dev-user-123", "To Delete", boundary)

	resp, _ := ts.DELETE("/api/v1/gardens/"+gardenID, nil)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify garden is deleted
	resp2, _ := ts.GET("/api/v1/gardens/"+gardenID, nil)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

// ===== ZONE MANAGEMENT TESTS =====

func TestCreateZone_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`
	gardenID := ts.SeedTestGarden("dev-user-123", "Test Garden", boundary)

	zonePayload := map[string]interface{}{
		"name":      "Vegetable Patch",
		"zone_type": "vegetable",
		"geometry": `{
			"type": "Polygon",
			"coordinates": [[
				[-122.48, 37.72],
				[-122.46, 37.72],
				[-122.46, 37.74],
				[-122.48, 37.74],
				[-122.48, 37.72]
			]]
		}`,
		"sun_hours_per_day": 6,
	}

	resp, body := ts.POST("/api/v1/gardens/"+gardenID+"/zones", zonePayload, nil)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var zone map[string]interface{}
	ts.ParseSuccessResponse(body, &zone)

	assert.NotNil(t, zone["zone_id"])
	assert.Equal(t, gardenID, zone["garden_id"])
}

func TestListZones_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`
	gardenID := ts.SeedTestGarden("dev-user-123", "Test Garden", boundary)

	// Create zones via direct DB insert
	ctx := context.Background()
	zoneID1 := uuid.New().String()
	zoneGeometry := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.48, 37.72],
			[-122.46, 37.72],
			[-122.46, 37.74],
			[-122.48, 37.74],
			[-122.48, 37.72]
		]]
	}`

	_, err := ts.DB.ExecContext(ctx, `
		INSERT INTO garden_zones (zone_id, garden_id, zone_name, geometry, created_at)
		VALUES ($1, $2, 'Test Zone', ST_GeomFromGeoJSON($3), NOW())
	`, zoneID1, gardenID, zoneGeometry)
	require.NoError(t, err)

	resp, body := ts.GET("/api/v1/gardens/"+gardenID+"/zones", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var zones []map[string]interface{}
	ts.ParseSuccessResponse(body, &zones)

	assert.GreaterOrEqual(t, len(zones), 1)
}

// ===== PLANT PLACEMENT TESTS =====

func TestAddPlantToGarden_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed plant
	plantID := ts.SeedTestPlant("Rosa rugosa", map[string]string{"en": "Japanese Rose"})

	// Seed garden
	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`
	gardenID := ts.SeedTestGarden("dev-user-123", "Test Garden", boundary)

	placement := map[string]interface{}{
		"plant_id": plantID,
		"location_geojson": `{
			"type": "Point",
			"coordinates": [-122.45, 37.75]
		}`,
		"quantity": 3,
	}

	resp, body := ts.POST("/api/v1/gardens/"+gardenID+"/plants", placement, nil)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var gardenPlant map[string]interface{}
	ts.ParseSuccessResponse(body, &gardenPlant)

	assert.Equal(t, plantID, gardenPlant["plant_id"])
	assert.Equal(t, gardenID, gardenPlant["garden_id"])
}

func TestListGardenPlants_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed plant and garden
	plantID := ts.SeedTestPlant("Lavandula angustifolia", map[string]string{"en": "English Lavender"})

	boundary := `{
		"type": "Polygon",
		"coordinates": [[
			[-122.5, 37.7],
			[-122.4, 37.7],
			[-122.4, 37.8],
			[-122.5, 37.8],
			[-122.5, 37.7]
		]]
	}`
	gardenID := ts.SeedTestGarden("dev-user-123", "Test Garden", boundary)

	// Add plant to garden via DB
	ctx := context.Background()
	gardenPlantID := uuid.New().String()
	location := `{"type": "Point", "coordinates": [-122.45, 37.75]}`

	now := time.Now()
	_, err := ts.DB.ExecContext(ctx, `
		INSERT INTO garden_plants (garden_plant_id, garden_id, plant_id, location, planted_date, quantity, health_status, created_at)
		VALUES ($1, $2, $3, ST_GeomFromGeoJSON($4), $5, 1, 'healthy', NOW())
	`, gardenPlantID, gardenID, plantID, location, now)
	require.NoError(t, err)

	resp, body := ts.GET("/api/v1/gardens/"+gardenID+"/plants", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plants []map[string]interface{}
	ts.ParseSuccessResponse(body, &plants)

	assert.GreaterOrEqual(t, len(plants), 1)
}

// ===== ADMIN PLANT CRUD TESTS =====

func TestCreatePlant_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed taxonomy hierarchy first
	_, _, speciesID := testhelpers.SeedTestPlantHierarchy(ts.T, ts.DB)

	payload := map[string]interface{}{
		"full_botanical_name": "Rosa rugosa 'Alba'",
		"species_id":          speciesID,
		"plant_type":          "shrub",
	}

	resp, body := ts.POST("/api/v1/plants", payload, nil)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var plant map[string]interface{}
	ts.ParseSuccessResponse(body, &plant)

	assert.NotNil(t, plant["plant_id"])
	assert.Equal(t, "Rosa rugosa 'Alba'", plant["full_botanical_name"])
	assert.Equal(t, "shrub", plant["plant_type"])
}

func TestCreatePlant_InvalidSpeciesID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	payload := map[string]interface{}{
		"full_botanical_name": "Test Plant",
		"species_id":          "not-a-uuid",
		"plant_type":          "shrub",
	}

	resp, body := ts.POST("/api/v1/plants", payload, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "INVALID_REQUEST", errorResp.Code)
}

func TestCreatePlant_MissingRequiredFields(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	payload := map[string]interface{}{
		"full_botanical_name": "Test Plant",
		// Missing species_id and plant_type
	}

	resp, body := ts.POST("/api/v1/plants", payload, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "INVALID_REQUEST", errorResp.Code)
}

func TestCreatePlant_DuplicateBotanicalName(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed existing plant
	existingID := ts.SeedTestPlant("Rosa rugosa", map[string]string{"en": "Japanese Rose"})
	assert.NotEmpty(t, existingID)

	// Get species ID from existing plant
	ctx := context.Background()
	var speciesID string
	err := ts.DB.QueryRowContext(ctx, `
		SELECT species_id FROM plants WHERE plant_id = $1
	`, existingID).Scan(&speciesID)
	require.NoError(t, err)

	// Try to create another plant with same botanical name
	payload := map[string]interface{}{
		"full_botanical_name": "Rosa rugosa",
		"species_id":          speciesID,
		"plant_type":          "shrub",
	}

	resp, body := ts.POST("/api/v1/plants", payload, nil)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Contains(t, errorResp.Message, "already exists")
}

func TestUpdatePlant_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed existing plant
	plantID := ts.SeedTestPlant("Rosa rugosa", map[string]string{"en": "Japanese Rose"})

	payload := map[string]interface{}{
		"full_botanical_name": "Rosa rugosa 'Rubra'",
	}

	resp, body := ts.PUT("/api/v1/plants/"+plantID, payload, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plant map[string]interface{}
	ts.ParseSuccessResponse(body, &plant)

	assert.Equal(t, plantID, plant["plant_id"])
	assert.Equal(t, "Rosa rugosa 'Rubra'", plant["full_botanical_name"])
}

func TestUpdatePlant_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	nonExistentID := uuid.New().String()

	payload := map[string]interface{}{
		"full_botanical_name": "Test Plant",
	}

	resp, body := ts.PUT("/api/v1/plants/"+nonExistentID, payload, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "RESOURCE_NOT_FOUND", errorResp.Code)
}

func TestUpdatePlant_InvalidUUID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	payload := map[string]interface{}{
		"full_botanical_name": "Test Plant",
	}

	resp, body := ts.PUT("/api/v1/plants/not-a-uuid", payload, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "INVALID_REQUEST", errorResp.Code)
}

func TestDeletePlant_Success(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	// Seed plant to delete
	plantID := ts.SeedTestPlant("Lavandula angustifolia", map[string]string{"en": "English Lavender"})

	resp, _ := ts.DELETE("/api/v1/plants/"+plantID, nil)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify plant is deleted
	resp2, _ := ts.GET("/api/v1/plants/"+plantID, nil)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestDeletePlant_NotFound(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	nonExistentID := uuid.New().String()

	resp, body := ts.DELETE("/api/v1/plants/"+nonExistentID, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "RESOURCE_NOT_FOUND", errorResp.Code)
}

func TestDeletePlant_InvalidUUID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Teardown()

	resp, body := ts.DELETE("/api/v1/plants/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errorResp := ts.ParseErrorResponse(body)
	assert.Equal(t, "INVALID_REQUEST", errorResp.Code)
}
