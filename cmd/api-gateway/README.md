# Twigger API Gateway

## Overview
REST API Gateway for the Twigger Plant Database & Garden Management System.
**Status**: 🚧 In Development (Part 5 - ~80% Complete)

## Architecture

```
cmd/api-gateway/          # Main application entry point
├── main.go              # Server initialization and dependency injection

internal/api-gateway/     # API Gateway internals
├── handlers/            # HTTP request handlers
│   ├── plant_handler.go         # Plant API endpoints
│   ├── garden_handler.go        # Garden API endpoints
│   ├── zone_handler.go          # Garden zone endpoints
│   ├── plant_placement_handler.go # Plant placement endpoints
│   └── health_handler.go        # Health check endpoints
├── middleware/          # HTTP middleware
│   ├── auth.go         # Firebase JWT authentication
│   ├── logging.go      # Request logging
│   ├── cors.go         # CORS configuration
│   └── ratelimit.go    # Rate limiting (token bucket)
├── router/             # Route configuration
│   └── router.go       # Main router setup with all routes
└── utils/              # Shared utilities
    ├── response.go     # Standardized JSON responses
    ├── language.go     # Language context extraction
    └── validation.go   # Request validation helpers
```

## Implemented Features ✅

### 1. **Complete REST API Structure**
   - **24 endpoints** implemented across 5 handler types
   - Standardized JSON request/response format
   - Comprehensive error handling with error codes

### 2. **Authentication & Authorization**
   - Firebase JWT token validation middleware
   - User context extraction (user_id, language preferences)
   - Optional auth for public endpoints, required for user-specific endpoints
   - Development mode (AUTH_ENABLED=false) for testing

### 3. **Language Context & Localization**
   - Automatic language extraction from `Accept-Language` header
   - Support for language+country combinations (e.g., `en-US`, `es-MX`)
   - Fallback to user preferences from Firebase claims
   - **Replaces hardcoded "en"** in service layer (primary goal achieved ✅)

### 4. **Middleware Stack**
   - **Logging**: Request/response logging with duration tracking
   - **CORS**: Configurable allowed origins
   - **Rate Limiting**: 100 requests/minute per IP (token bucket algorithm)
   - **Request ID Tracking**: For distributed tracing

### 5. **Plant API Endpoints**
   | Method | Endpoint | Auth | Description |
   |--------|----------|------|-------------|
   | GET | `/api/v1/plants/:id` | Optional | Get plant by ID with language context |
   | GET | `/api/v1/plants/search` | Optional | Search plants with filters (physical characteristics) |
   | GET | `/api/v1/plants/:id/companions` | Optional | Get companion plants |
   | GET | `/api/v1/plants/family/:name` | Optional | Find plants by family |
   | GET | `/api/v1/plants/genus/:name` | Optional | Find plants by genus |
   | GET | `/api/v1/plants/recommend` | Optional | Get plant recommendations |
   | POST | `/api/v1/plants` | Required (Admin) | Create plant |
   | PUT | `/api/v1/plants/:id` | Required (Admin) | Update plant |
   | DELETE | `/api/v1/plants/:id` | Required (Admin) | Delete plant |

### 6. **Garden API Endpoints**
   | Method | Endpoint | Auth | Description |
   |--------|----------|------|-------------|
   | POST | `/api/v1/gardens` | Required | Create garden (auto-detect hardiness zone) |
   | GET | `/api/v1/gardens` | Required | List user's gardens |
   | GET | `/api/v1/gardens/:id` | Required | Get garden details |
   | PUT | `/api/v1/gardens/:id` | Required | Update garden |
   | DELETE | `/api/v1/gardens/:id` | Required | Delete garden |
   | GET | `/api/v1/gardens/stats` | Required | Get garden statistics |
   | GET | `/api/v1/gardens/nearby` | Required | Find nearby gardens |

### 7. **Zone Management Endpoints**
   | Method | Endpoint | Auth | Description |
   |--------|----------|------|-------------|
   | POST | `/api/v1/gardens/:id/zones` | Required | Create zone in garden |
   | GET | `/api/v1/gardens/:id/zones` | Required | List garden zones |
   | GET | `/api/v1/zones/:id` | Required | Get zone details |
   | PUT | `/api/v1/zones/:id` | Required | Update zone |
   | DELETE | `/api/v1/zones/:id` | Required | Delete zone |
   | GET | `/api/v1/zones/:id/area` | Required | Calculate zone area |

### 8. **Plant Placement Endpoints**
   | Method | Endpoint | Auth | Description |
   |--------|----------|------|-------------|
   | POST | `/api/v1/gardens/:id/plants` | Required | Place plant in garden |
   | POST | `/api/v1/gardens/:id/plants/bulk` | Required | Bulk place plants (transaction) |
   | GET | `/api/v1/gardens/:id/plants` | Required | List garden plants |
   | GET | `/api/v1/garden-plants/:id` | Required | Get placement details |
   | PUT | `/api/v1/garden-plants/:id` | Required | Update plant placement |
   | DELETE | `/api/v1/garden-plants/:id` | Required | Remove plant |

### 9. **Health Check Endpoints**
   | Method | Endpoint | Auth | Description |
   |--------|----------|------|-------------|
   | GET | `/health` | None | Health check with database status |
   | GET | `/ready` | Readiness probe for orchestration |

## Configuration

### Environment Variables
```bash
# Server
PORT=8080
ENVIRONMENT=development|production

# Database
DATABASE_URL="postgres://user:pass@host:5432/db"
CLOUD_SQL_PROXY=true  # Use Cloud SQL Proxy instead of direct connection

# Authentication
FIREBASE_PROJECT_ID=twigger
AUTH_ENABLED=true  # Set to false for development without Firebase

# CORS
ALLOWED_ORIGINS="http://localhost:3000,http://localhost:8080"

# Logging
LOG_LEVEL=info|debug|warn|error
```

## Request/Response Format

### Success Response
```json
{
  "data": { ... },
  "meta": {
    "cursor": "nextPageCursor",
    "has_more": true,
    "limit": 20,
    "total": 150
  }
}
```

### Error Response
```json
{
  "error": "validation_error",
  "code": "VALIDATION_ERROR",
  "message": "Invalid plant ID format",
  "details": {
    "field": "plant_id"
  }
}
```

### Error Codes
- `RESOURCE_NOT_FOUND` (404) - Resource doesn't exist
- `VALIDATION_ERROR` (400) - Invalid request data
- `UNAUTHORIZED` (401) - Missing or invalid authentication
- `FORBIDDEN` (403) - User lacks permission
- `RATE_LIMIT_EXCEEDED` (429) - Too many requests
- `DATABASE_ERROR` (500) - Internal database error
- `INTERNAL_SERVER_ERROR` (500) - Unexpected error

## Language Context

The API automatically extracts language context from:

1. **User preferences** (from Firebase JWT claims):
   ```json
   {
     "preferred_language": "es",
     "country": "MX"
   }
   ```

2. **Accept-Language header**:
   ```
   Accept-Language: es-MX, es;q=0.9, en;q=0.8
   ```
   Parsed as: `languageID = "es"`, `countryID = "MX"`

3. **Default fallback**: `en` (English)

## Running the API Gateway

### Prerequisites
- Go 1.25+
- PostgreSQL 17 with PostGIS 3.5
- Plant Service and Garden Service domains implemented
- Database migrations applied (001-007)

### Development
```bash
# With Cloud SQL Proxy
export CLOUD_SQL_PROXY=true
export AUTH_ENABLED=false  # Disable Firebase for local dev
go run cmd/api-gateway/main.go

# Direct database connection
export DATABASE_URL="postgres://..."
go run cmd/api-gateway/main.go
```

### Build & Deploy
```bash
# Build binary
go build -o api-gateway cmd/api-gateway/main.go

# Build Docker image
docker build -t twigger-api-gateway -f cmd/api-gateway/Dockerfile .

# Deploy to Cloud Run
gcloud run deploy api-gateway \
  --image gcr.io/twigger/api-gateway \
  --region us-central1 \
  --set-env-vars FIREBASE_PROJECT_ID=twigger
```

## Remaining Work 📋

### Critical (Blocks Testing)
1. **Fix Service Interface Mismatches** (8 errors)
   - `plant_placement_service.go`: Update to use actual `GardenPlant` entity fields
   - `zone_management_service.go`: Fix `CalculateArea` method call
   - These are service layer issues, not API Gateway issues

### Important (Production Ready)
2. **Add Integration Tests**
   - Test all 24 endpoints with real database
   - Test authentication flow
   - Test language context extraction
   - Test error handling

3. **Add OpenAPI/Swagger Documentation**
   - Generate OpenAPI 3.0 spec
   - Host Swagger UI at `/api/docs`
   - Document all request/response schemas

4. **Implement Remaining Plant CRUD** (Admin endpoints)
   - `POST /api/v1/plants` - Create plant
   - `PUT /api/v1/plants/:id` - Update plant
   - `DELETE /api/v1/plants/:id` - Delete plant
   - Currently return `501 Not Implemented`

### Nice to Have
5. **Enhance Middleware**
   - Add distributed tracing (OpenTelemetry)
   - Add Prometheus metrics
   - Add circuit breaker for database calls

6. **Performance Optimizations**
   - Add caching layer (Redis)
   - Add connection pooling monitoring
   - Add slow query logging

## Testing Examples

### Get Plant with Language
```bash
curl http://localhost:8080/api/v1/plants/123 \
  -H "Accept-Language: es-MX"
# Returns plant with Spanish (Mexico) common names
```

### Search Plants
```bash
curl "http://localhost:8080/api/v1/plants/search?q=tomato&min_height=1.0&max_height=3.0&limit=20"
# Returns plants matching search with height filters
```

### Create Garden (Authenticated)
```bash
curl -X POST http://localhost:8080/api/v1/gardens \
  -H "Authorization: Bearer ${FIREBASE_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Garden",
    "location_geojson": "{\"type\":\"Point\",\"coordinates\":[-122.4,37.7]}",
    "boundary_geojson": "{\"type\":\"Polygon\",\"coordinates\":[...]}"
  }'
# Auto-detects hardiness zone based on location
```

## Dependencies

```go
require (
    github.com/gorilla/mux v1.8.1       // HTTP router
    github.com/lib/pq v1.10.9           // PostgreSQL driver
    github.com/google/uuid v1.6.0        // UUID generation
)
```

## Contributing

When adding new endpoints:

1. Create handler in `internal/api-gateway/handlers/`
2. Add route in `internal/api-gateway/router/router.go`
3. Update this README with endpoint documentation
4. Add integration tests
5. Update OpenAPI spec

## License

Proprietary - Twigger

---

**Last Updated**: 2025-10-03
**Part 5 Status**: 80% Complete (Core implementation done, needs service layer fixes + testing)
