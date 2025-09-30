# Claude Code Instructions

## Project Context
You are building a plant database system with spatial garden mapping capabilities. The system uses PostgreSQL with PostGIS for spatial operations, Go for backend services, and GraphQL/REST APIs. This is a backend-only implementation (no Flutter frontend in this phase).

## Current State
- Project structure is initialized
- Database schema (v5.0) is designed with full PostGIS support
- Project is divided into 7 independent parts for parallel development
- Mock services are available for testing each part independently

## Development Approach

### When Starting a New Part
1. **Read the part specification** in `/docs/project-parts-breakdown.md`
2. **Check dependencies** - ensure prerequisite parts are complete or mocked
3. **Use the test harness** - `./tests/test-parts.sh [part_number]`
4. **Import mocks** from `/backend/shared/mocks/` for dependencies

### Code Style Guidelines

#### Go Code
```go
// Use domain-driven design structure
domain/
  entity.go      // Core business objects
  repository.go  // Interface definitions
  service.go     // Business logic

// Error handling pattern
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}

// Context propagation
func (s *Service) Method(ctx context.Context, ...) error {
    // Always propagate context
}

// Dependency injection
type Service struct {
    repo Repository  // Interface, not concrete type
    cache Cache
}
```

#### SQL/PostGIS
```sql
-- Use PostGIS functions for spatial operations
ST_Contains(garden.boundary, zone.geometry)
ST_Distance(point1::geography, point2::geography)
ST_Area(polygon::geography)

-- Always create indexes for foreign keys and spatial columns
CREATE INDEX idx_gardens_boundary ON gardens USING GIST (boundary);
```

#### Testing
```go
// Unit tests with mocks
func TestServiceMethod(t *testing.T) {
    mock := mocks.NewMockRepository()
    service := NewService(mock)
    // Test business logic only
}

// Integration tests with real database
// +build integration
func TestRepositoryMethod(t *testing.T) {
    // Test with PostgreSQL
}
```

### Working with Spatial Data

#### GeoJSON Input/Output
```go
// Always validate GeoJSON on input
func ValidateGeoJSON(geojson string) error {
    // Check valid structure
    // Ensure polygon closure
    // Check for self-intersection
}

// Convert for PostGIS storage
geometry := fmt.Sprintf("ST_GeomFromGeoJSON('%s')", geojson)

// Return as GeoJSON
SELECT ST_AsGeoJSON(boundary) FROM gardens
```

#### Coordinate Systems
- **Storage**: Always use WGS84 (SRID 4326)
- **Display**: Convert to Web Mercator for maps
- **Calculations**: Use geography type for accurate measurements

### API Development Patterns

#### GraphQL Resolvers
```go
// Use DataLoader to prevent N+1 queries
func (r *Resolver) Plants(ctx context.Context, gardenID string) ([]*Plant, error) {
    return r.PlantLoader.Load(ctx, gardenID)
}

// Handle errors gracefully
if err != nil {
    return nil, gqlerror.Errorf("failed to load plants: %v", err)
}
```

#### REST Handlers
```go
// Consistent error responses
type ErrorResponse struct {
    Error string `json:"error"`
    Code  int    `json:"code"`
}

// Validation middleware
func ValidateRequest(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Validate before handler
        next(w, r)
    }
}
```

### Database Patterns

#### Connection Management
```go
// Use connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Always use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

#### Transactions
```go
// Use transactions for multi-table operations
tx, err := db.BeginTx(ctx, nil)
defer tx.Rollback() // Always rollback if not committed

// Operations...

if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit failed: %w", err)
}
```

### Common Pitfalls to Avoid

1. **Don't forget PostGIS indexes** - Spatial queries are slow without GIST indexes
2. **Don't ignore context cancellation** - Always propagate context
3. **Don't hardcode configurations** - Use environment variables
4. **Don't skip validation** - Especially for spatial data
5. **Don't mix concerns** - Keep domain logic separate from infrastructure

### Performance Considerations

#### Spatial Queries
- Pre-calculate and cache complex spatial analysis
- Use `ST_SimplifyPreserveTopology` for display geometries
- Consider spatial indexes for all geometry columns
- Use `EXPLAIN ANALYZE` for query optimization

#### Caching Strategy
```go
// Cache key pattern
key := fmt.Sprintf("plant:%s:gardens:%s", plantID, gardenID)

// Cache with TTL
cache.Set(ctx, key, data, 1*time.Hour)

// Check cache before database
if cached, err := cache.Get(ctx, key); err == nil {
    return cached
}
```

### Security Implementation

#### Authentication Middleware
```go
// Extract and validate Firebase token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        user, err := validateToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        ctx := context.WithValue(r.Context(), "user", user)
        next(w, r.WithContext(ctx))
    }
}
```

#### Workspace Isolation
```go
// Always filter by workspace
query := `
    SELECT * FROM gardens 
    WHERE workspace_id = $1 AND id = $2
`
```

### Testing Each Part

#### Part-Specific Tests
```bash
# Run specific part test
./tests/test-parts.sh 3  # Test Garden Spatial Service

# Run integration tests
docker-compose -f docker-compose.test.yml up -d
go test ./... -tags=integration

# Run benchmarks
go test -bench=. ./...
```

#### Mock Usage
```go
import "github.com/twigger/plant-database/backend/shared/mocks"

// Use mocks for dependencies
plantService := mocks.NewMockPlantService()
gardenRepo := mocks.NewMockGardenRepository()
```

### Deployment Readiness

Before marking a part as complete:
1. ✅ Unit tests pass with >80% coverage
2. ✅ Integration tests pass
3. ✅ API documentation updated
4. ✅ Performance benchmarks meet requirements
5. ✅ Security review completed
6. ✅ Error handling comprehensive
7. ✅ Logging implemented
8. ✅ Metrics exposed

### File Organization

```
backend/
├── [service-name]/
│   ├── cmd/
│   │   └── main.go           # Entry point
│   ├── internal/
│   │   ├── domain/           # Business logic
│   │   ├── infrastructure/   # External services
│   │   └── interfaces/       # API layer
│   ├── pkg/                  # Shared packages
│   ├── Dockerfile
│   └── go.mod
```

### Environment Variables

Always use these environment variables:
```bash
# Database
DB_HOST=/cloudsql/instance-connection-name  # For Cloud SQL
DB_USER=plant_api
DB_PASSWORD=${SECRET_DB_PASSWORD}
DB_NAME=plantdb

# Service
ENVIRONMENT=dev|staging|prod
LOG_LEVEL=debug|info|warn|error
PORT=8080

# Cache
REDIS_HOST=redis-host
REDIS_PORT=6379

# Auth
FIREBASE_PROJECT=twigger
```

### Git Workflow

```bash
# Branch naming
feature/part-1-database-setup
feature/part-2-plant-service
bugfix/spatial-query-performance

# Commit messages
feat: implement plant search with PostGIS
fix: correct spatial index on garden_zones
test: add integration tests for garden service
docs: update API documentation for GraphQL

# PR description template
## What
Description of changes

## Why
Business/technical rationale

## Testing
How to test these changes

## Checklist
- [ ] Tests written and passing
- [ ] Documentation updated
- [ ] Performance verified
```

### Questions to Ask Before Implementation

1. **Is there a mock for this dependency?** Check `/backend/shared/mocks/`
2. **What's the PostGIS function for this?** Check PostGIS docs
3. **Is this query optimized?** Run EXPLAIN ANALYZE
4. **Is error handling complete?** Check all error paths
5. **Is this testable independently?** Write unit tests with mocks

### Getting Help

When stuck:
1. Check existing patterns in codebase
2. Review PostGIS documentation for spatial operations
3. Check test files for usage examples
4. Review mocks for interface contracts
5. Run the test harness for validation

### Important Files Reference

- **Database Schema**: `/database/schema/plant_database_v5_postgis.sql`
- **Project Parts**: `/docs/project-parts-breakdown.md`
- **Mock Services**: `/backend/shared/mocks/mocks.go`
- **Test Harness**: `/tests/test-parts.sh`
- **API Routes**: `/backend/plant-api/internal/interfaces/http/router.go`
- **GraphQL Schema**: `/backend/graphql-gateway/internal/graph/schema.graphqls`
- **Architecture**: `/architecture.md`
- **Tasks**: `/tasks.md`