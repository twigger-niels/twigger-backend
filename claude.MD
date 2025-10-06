# Claude Code Instructions

## Code Style Guidelines

### Go Code
```go
// Domain-driven design structure
domain/
  entity.go      // Core business objects
  repository.go  // Interface definitions
  service.go     // Business logic

// Error handling: always wrap errors
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}

// Context propagation
func (s *Service) Method(ctx context.Context, ...) error {
    // Always propagate context
}

// Dependency injection - use interfaces
type Service struct {
    repo Repository  // Interface, not concrete type
}
```

### SQL/PostGIS
```sql
-- Spatial operations
ST_Contains(garden.boundary, zone.geometry)
ST_Distance(point1::geography, point2::geography)
ST_Area(polygon::geography)

-- Always index foreign keys and spatial columns
CREATE INDEX idx_gardens_boundary ON gardens USING GIST (boundary);
```

### Testing
```go
// Unit tests with mocks
func TestServiceMethod(t *testing.T) {
    mock := mocks.NewMockRepository()
    service := NewService(mock)
}

// Integration tests
// +build integration
func TestRepositoryMethod(t *testing.T) {
    // Test with PostgreSQL
}
```

## Spatial Data

### GeoJSON
```go
// Always validate before ST_GeomFromGeoJSON
func ValidateGeoJSON(geojson string) error {
    // Check: valid JSON, type field, coordinates array, polygon closure
}
```

### Coordinate Systems
- **Storage**: WGS84 (SRID 4326)
- **Display**: Web Mercator
- **Calculations**: geography type for accurate measurements

## Localization

### Language Context
```go
// ALWAYS include language context
func GetPlant(plantID, languageID string, countryID *string) (*Plant, error)
```

### Fallback Chain
```sql
-- 4-tier fallback in SQL
SELECT COALESCE(
    (SELECT name WHERE language_id = $1 AND country_id = $2),  -- 1. Country + Language
    (SELECT name WHERE language_id = $1 AND country_id IS NULL), -- 2. Language global
    (SELECT name WHERE language_id = 'en' AND country_id IS NULL), -- 3. English
    botanical_name  -- 4. Fallback to key
)
```

### Cache Keys
```go
// Include ALL parameters that affect results
key := fmt.Sprintf("plant:%s:lang:%s", plantID, languageID)
if countryID != nil {
    key = fmt.Sprintf("plant:%s:lang:%s:country:%s", plantID, languageID, *countryID)
}

// Invalidate all variants
cache.DeletePattern(ctx, fmt.Sprintf("plant:%s:*", plantID))
```

## API Patterns

### GraphQL - Prevent N+1
```go
func (r *Resolver) Plants(ctx context.Context, gardenID string) ([]*Plant, error) {
    return r.PlantLoader.Load(ctx, gardenID)  // Use DataLoader
}
```

### REST - Consistent Errors
```go
type ErrorResponse struct {
    Error string `json:"error"`
    Code  int    `json:"code"`
}
```

## Database

### Connection Pooling
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Transactions
```go
tx, err := db.BeginTx(ctx, nil)
defer tx.Rollback()  // Always rollback if not committed
// ... operations
tx.Commit()
```

## Critical Gotchas (45)

### Repository Implementation (10-14)

**10. Every Entity Field in ALL Queries**
- SELECT: All fields
- INSERT: All fields (including created_at, updated_at)
- UPDATE: All mutable fields + updated_at
- Scan(): Matching field count

**11. Timestamp Handling**
- Create: Set both `created_at` and `updated_at` to `time.Now()`
- Update: Set `updated_at` to `time.Now()`

**12. Scan() Count Must Match SELECT Columns**
Count columns vs destinations before running

**13. Nullable Fields = Pointer Types**
`*time.Time`, `*string`, `*int` for nullable fields

**14. scanX() Helpers Must Match ALL Queries**
Grep and verify: `rg "scanEntities"`

### Localization (1-9)

**1. N+1 Query Risk**
Implement `loadXForMultipleY()` batch methods for one-to-many

**2. Language Context Required**
Always include `languageID` and `countryID` parameters

**3. Language-Aware Cache Keys**
Include language in cache key, use pattern deletion

**4. Composite Indexes**
Match WHERE clause order: `(plant_id, language_id, country_id)`

**5. Input Validation at Boundary**
Accept UUIDs OR ISO codes (en, es, US, MX)

**6. Fallback Chain**
4-tier: Country+Lang → Lang → English → Key

**7. Migration Dependency**
Migration 005 required for localization

**8. Pattern-Based Invalidation**
Use Redis SCAN, not KEYS (blocking)

**9. Companion Plants Context**
Pass through user's language context

### Integration Testing (15-20)

**15. Use UUIDs, Not ISO Codes**
Define constants: `englishID := "550e8400-e29b-41d4-a716-446655440001"`

**16. Schema Must Match Production**
Read migration files, run `\d table_name`

**17. Cleanup: DROP SCHEMA CASCADE**
```go
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
CREATE EXTENSION IF NOT EXISTS postgis;
```

**18. Minimal Schema for Tests**
Create only tables tests need

**19. Use PostGIS Docker Image**
`image: postgis/postgis:17-3.5`

**20. Valid UUID Format**
`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` (0-9, a-f only)

### Query Filters (21-26)

**21. Zone Matching - Use Regex**
```go
// Use regex with word boundaries
query += " AND hardiness_zones ~ '(^|,)1a(,|$)'"
```

**22. Height Filter Logic**
- MinHeight: `WHERE max_m >= $1` (can it grow this tall?)
- MaxHeight: `WHERE typical_m <= $1` (will it fit?)

**23. LEFT JOIN for Characteristics**
Not all plants have data - allow graceful degradation

**24. Enum Array Casting**
```sql
VALUES (ARRAY['full_sun']::sun_requirement[], ...)
```

**25. JSONB Boolean Casting**
```go
WHERE (pc.traits->>'evergreen')::boolean = true
```

**26. Column Names from Migrations**
Don't assume from entity field names

### Performance (27-35)

**27. GIN Trigram for ILIKE**
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_name_trgm ON table USING GIN (name gin_trgm_ops);
```
100x improvement

**28. Prepared Statement Fallback**
Handle connection pool resets with fallback to regular query

**29. Cache Keys = Query Parameters**
Include ALL parameters: `language_id`, `country_id`, `limit`, `offset`

**30. Go 1.18+ for Generics**
Zero runtime overhead, compile-time type safety

**31. Transaction Panic Recovery**
```go
defer func() {
    if p := recover(); p != nil {
        tx.Rollback()
        panic(p)
    }
}()
```

**32. Validate GeoJSON First**
Before `ST_GeomFromGeoJSON` - clearer errors, security

**33. Coordinate Bounds**
Lat: -90 to 90, Lng: -180 to 180

**34. Savepoints for Partial Rollback**
`SAVEPOINT name`, `ROLLBACK TO SAVEPOINT name`

**35. PascalCase Repository Structs**
Exported for godoc visibility

### PostGIS Spatial (36-45)

**36. Direct Coordinate Extraction**
```sql
-- 3x faster
SELECT ST_Y(ST_Centroid(boundary)) as lat
```
Don't use `ST_AsText()` then cast back

**37. GIST Indexes REQUIRED**
```sql
CREATE INDEX idx_location ON gardens USING GIST(location);
```
100x improvement

**38. Geography for Distance**
```sql
-- Returns meters
SELECT ST_Distance(point1::geography, point2::geography);
```

**39. Bounding Box Pre-filter**
```sql
WHERE geometry && point           -- Fast bbox
  AND ST_Contains(geometry, point) -- Precise
```

**40. NULL Handling in Queries**
Build query dynamically or use `sql.NullString`

**41. GeoJSON Validation Checklist**
Valid JSON, type field, coordinates, closed polygons, bounds

**42. Polygon Closure**
First point == last point

**43. MultiPolygon Structure**
Different coordinate depth than Polygon

**44. Service ≠ Database Queries**
Repository: All queries | Service: Business logic | API: HTTP/GraphQL

**45. Transactions in Repository**
Not service - keeps service database-agnostic

## Security

### Authentication
```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    token := extractToken(r)
    user, err := validateToken(token)
    ctx := context.WithValue(r.Context(), "user", user)
}
```

### Workspace Isolation
```go
query := `SELECT * FROM gardens WHERE workspace_id = $1 AND id = $2`
```

## Environment Variables

```bash
# Database
DB_HOST=/cloudsql/instance-connection-name
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

## Git Workflow

```bash
# Branch naming
feature/part-1-database-setup
bugfix/spatial-query-performance

# Commit messages
feat: implement plant search with PostGIS
fix: correct spatial index on garden_zones
test: add integration tests
docs: update API documentation
```
