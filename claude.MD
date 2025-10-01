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

### Localization Patterns

#### Always Query with Language Context
```go
// WRONG - Missing localization
func GetPlant(plantID string) (*Plant, error) {
    query := `SELECT * FROM plants WHERE plant_id = $1`
}

// CORRECT - Include language context
func GetPlant(plantID, languageID, countryID string) (*Plant, error) {
    query := `
        SELECT 
            p.*,
            array_agg(pcn.common_name) AS common_names
        FROM plants p
        LEFT JOIN plant_common_names pcn ON p.plant_id = pcn.plant_id
        WHERE p.plant_id = $1 
          AND pcn.language_id = $2
          AND (pcn.country_id = $3 OR pcn.country_id IS NULL)
        GROUP BY p.plant_id
    `
}
```

#### Implement Fallback Strategy
```go
func GetLocalizedText(key, langID, countryID string) string {
    // 1. Try country + language specific
    text := getCountrySpecific(key, langID, countryID)
    if text != "" {
        return text
    }
    
    // 2. Try language global
    text = getLanguageGlobal(key, langID)
    if text != "" {
        return text
    }
    
    // 3. Fallback to English
    text = getEnglish(key)
    if text != "" {
        return text
    }
    
    // 4. Return key as last resort
    return key
}
```

#### Cache Translations
```go
// Cache translations per language to avoid repeated queries
type TranslationCache struct {
    mu sync.RWMutex
    cache map[string]map[string]string // [language_id][key]value
}

func (tc *TranslationCache) Get(langID, key string) string {
    tc.mu.RLock()
    defer tc.mu.RUnlock()
    
    if lang, ok := tc.cache[langID]; ok {
        if val, ok := lang[key]; ok {
            return val
        }
    }
    return ""
}
```

#### Handle Multi-Language Search
```go
// Search should work across all languages
func SearchPlants(query, languageID string) ([]*Plant, error) {
    searchQuery := `
        WITH name_matches AS (
            SELECT DISTINCT p.plant_id
            FROM plants p
            LEFT JOIN plant_common_names pcn ON p.plant_id = pcn.plant_id
            WHERE pcn.common_name ILIKE $1
               OR p.full_botanical_name ILIKE $1
        )
        SELECT ... FROM plants WHERE plant_id IN (SELECT plant_id FROM name_matches)
    `
    return db.Query(searchQuery, "%"+query+"%")
}
```

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

### Critical Localization Gotchas

#### 1. N+1 Query Risk with Localized Data
**Problem**: Loading localized data (common names, descriptions) in loops causes N+1 queries
```go
// WRONG - Executes 1 query per plant (N+1 problem)
for _, plant := range plants {
    names, _ := repo.GetCommonNames(plant.ID, languageID)
    plant.CommonNames = names
}

// CORRECT - Batch loading with single IN clause query
func (r *Repository) loadCommonNamesForMultiplePlants(ctx context.Context, plants []*Plant, languageID string) error {
    plantIDs := extractIDs(plants)
    query := `
        SELECT plant_id, common_name
        FROM plant_common_names
        WHERE plant_id IN (...)  -- Single query for all plants
          AND language_id = $1
    `
    // Group results and assign to plants
}
```
**Rule**: Always implement `loadXForMultipleY()` batch methods for one-to-many localized relationships

#### 2. Language Context Required Everywhere
**Problem**: Missing languageID/countryID parameters causes incomplete data
```go
// WRONG - Missing language context
func FindByID(ctx context.Context, plantID string) (*Plant, error)

// CORRECT - Always include language context
func FindByID(ctx context.Context, plantID, languageID string, countryID *string) (*Plant, error)
```
**Current State**: Service layer uses hardcoded `"en"` with TODO comments. These MUST be replaced when API layer extracts language from request context.

**Migration Path**:
```go
// Service layer (temporary - until Part 4)
plant, err := s.repo.FindByID(ctx, plantID, "en", nil)  // TODO: Extract from context

// API layer (Part 4 implementation)
langCtx := middleware.ExtractLanguageContext(ctx)
plant, err := s.repo.FindByID(ctx, plantID, langCtx.LanguageID, langCtx.CountryID)
```

#### 3. Language-Aware Cache Keys
**Problem**: Cache keys without language context cause stale data across languages
```go
// WRONG - Language-agnostic key (different languages overwrite each other)
key := fmt.Sprintf("plant:%s", plantID)

// CORRECT - Include language in cache key
key := fmt.Sprintf("plant:%s:%s", plantID, languageID)
if countryID != nil {
    key = fmt.Sprintf("plant:%s:%s:%s", plantID, languageID, *countryID)
}
```
**Cache Invalidation**: Use pattern matching to delete ALL language variants
```go
// When updating plant, invalidate all language variants
pattern := fmt.Sprintf("plant:%s:*", plantID)  // Matches plant:ID:en, plant:ID:es:MX, etc.
cache.DeletePattern(ctx, pattern)
```

#### 4. Composite Indexes Required for Localization Queries
**Problem**: Single-column indexes insufficient for multi-column localization filters
```sql
-- WRONG - Separate indexes (query planner may not use optimally)
CREATE INDEX idx_plant_id ON plant_common_names(plant_id);
CREATE INDEX idx_language_id ON plant_common_names(language_id);

-- CORRECT - Composite index covering exact query pattern
CREATE INDEX idx_plant_common_names_lookup
ON plant_common_names(plant_id, language_id, country_id);

-- Also needed for name-based searches
CREATE INDEX idx_plant_common_names_name_lang
ON plant_common_names(language_id, common_name);
```
**Rule**: Create composite indexes matching the WHERE clause order in your most common queries

#### 5. Input Validation at Repository Boundary
**Problem**: Malformed language/country IDs reaching SQL layer cause cryptic errors
```go
// Validate at repository entry point
func (r *Repository) FindByID(ctx context.Context, plantID, languageID string, countryID *string) (*Plant, error) {
    if err := ValidatePlantID(plantID); err != nil {
        return nil, fmt.Errorf("invalid plant_id: %w", err)
    }
    if err := ValidateLanguageID(languageID); err != nil {
        return nil, fmt.Errorf("invalid language_id: %w", err)
    }
    if err := ValidateCountryID(countryID); err != nil {
        return nil, fmt.Errorf("invalid country_id: %w", err)
    }
    // ... proceed with query
}
```
**Accept Both Formats**: Validators should accept UUIDs OR ISO codes (en, es, US, MX)

#### 6. Fallback Chain Implementation
**Problem**: Missing fallback logic leaves users with empty data
```sql
-- Implement 4-tier fallback in SQL
SELECT COALESCE(
    (SELECT name WHERE language_id = $1 AND country_id = $2),  -- 1. Country + Language
    (SELECT name WHERE language_id = $1 AND country_id IS NULL), -- 2. Language global
    (SELECT name WHERE language_id = 'en' AND country_id IS NULL), -- 3. English
    botanical_name  -- 4. Fallback to scientific name
)
```

#### 7. Localization Migration Dependency
**Critical**: Migration `005_add_localization.sql` MUST be applied before Part 2 features work
- Adds 8 localization tables (plant_common_names, plant_descriptions, etc.)
- Adds 3 helper functions (get_localized_text, get_characteristic_translation, etc.)
- Adds composite indexes for performance

**Check Before Development**:
```bash
# Verify migration applied
psql -c "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'plant_common_names')"
```

#### 8. Pattern-Based Cache Invalidation
**Problem**: Updating plant only invalidates one language variant
```go
// WRONG - Only deletes English cache entry
cache.Delete(ctx, fmt.Sprintf("plant:%s:en", plantID))

// CORRECT - Delete all language variants using wildcards
cache.DeletePattern(ctx, fmt.Sprintf("plant:%s:*", plantID))
```
**Redis SCAN Note**: Use SCAN instead of KEYS for production (non-blocking)

#### 9. Companion Plants Need Language Context
**Problem**: Companion relationships were hardcoded to English
```go
// WRONG - Hardcoded English for companion names
companions, err := r.repo.GetCompanions(ctx, plantID, "en", nil, filter)

// CORRECT - Pass through user's language context
companions, err := r.repo.GetCompanions(ctx, plantID, languageID, countryID, filter)
```
**Note**: Companion benefits (companion_benefits_i18n table) also localized - don't forget!

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