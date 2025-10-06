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

### Critical Repository Implementation Gotchas

#### 10. Every Entity Field MUST Appear in ALL Repository Methods
**Problem**: Missing fields in SELECT/INSERT/UPDATE queries cause silent data loss
```go
// WRONG - Missing patent_expiry field (data loss!)
query := `
    SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
           propagation_restricted, created_at
    FROM cultivars
    WHERE cultivar_id = $1
`

// CORRECT - All entity fields included
query := `
    SELECT cultivar_id, species_id, cultivar_name, trade_name, patent_number,
           patent_expiry, propagation_restricted, created_at
    FROM cultivars
    WHERE cultivar_id = $1
`
```

**Checklist When Creating Repository Methods**:
1. ✅ All entity fields in SELECT clause
2. ✅ Matching number of `Scan()` destinations
3. ✅ All fields in INSERT VALUES
4. ✅ All updatable fields in UPDATE SET
5. ✅ Corresponding ExecContext parameters

**Critical Fields Often Forgotten**:
- `updated_at` - Must be in SELECT, set in Create (with `created_at`), updated in Update
- `patent_expiry` / `date_deprecated` - Nullable time fields
- Enum/status fields with default values

#### 11. Timestamp Handling Pattern (created_at/updated_at)
**Problem**: Inconsistent timestamp initialization causes tracking issues
```go
// WRONG - Missing updated_at initialization in Create
func (r *Repository) Create(ctx context.Context, entity *Entity) error {
    query := `INSERT INTO table (id, name, created_at) VALUES ($1, $2, $3)`
    entity.CreatedAt = time.Now()
    // Missing: entity.UpdatedAt = entity.CreatedAt
}

// WRONG - Not updating updated_at in Update
func (r *Repository) Update(ctx context.Context, entity *Entity) error {
    query := `UPDATE table SET name = $2 WHERE id = $1`
    // Missing: entity.UpdatedAt = time.Now()
}

// CORRECT - Both timestamps managed properly
func (r *Repository) Create(ctx context.Context, entity *Entity) error {
    query := `INSERT INTO table (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`
    now := time.Now()
    entity.CreatedAt = now
    entity.UpdatedAt = now  // Initialize both to same value
    _, err := r.db.ExecContext(ctx, query, entity.ID, entity.Name, entity.CreatedAt, entity.UpdatedAt)
}

func (r *Repository) Update(ctx context.Context, entity *Entity) error {
    query := `UPDATE table SET name = $2, updated_at = $3 WHERE id = $1`
    entity.UpdatedAt = time.Now()  // Update timestamp
    _, err := r.db.ExecContext(ctx, query, entity.ID, entity.Name, entity.UpdatedAt)
}
```

**Rule**:
- **Create**: Set both `created_at` and `updated_at` to `time.Now()`
- **Update**: Set `updated_at` to `time.Now()`, don't touch `created_at`
- **SELECT**: Always retrieve both fields from database

#### 12. Scan() Destination Count Must Match SELECT Columns
**Problem**: Mismatched Scan() arguments cause runtime panics
```go
// WRONG - 7 SELECT columns but only 6 Scan destinations
query := `SELECT id, name, created_at, updated_at, status, email, phone FROM users`
err := db.Scan(&id, &name, &created_at, &updated_at, &status, &email)  // Missing &phone

// CORRECT - Matching counts
err := db.Scan(&id, &name, &created_at, &updated_at, &status, &email, &phone)
```

**Prevention**:
1. Count SELECT columns vs Scan arguments
2. Use consistent ordering (entity field declaration order)
3. Test with actual database queries (integration tests catch this)

#### 13. Nullable Fields Require Pointer Types
**Problem**: sql.Scan into non-pointer fields fails for NULL values
```go
// WRONG - Cannot scan NULL into time.Time
type Cultivar struct {
    PatentExpiry time.Time  // Fails if database value is NULL
}

// CORRECT - Use pointer for nullable fields
type Cultivar struct {
    PatentExpiry *time.Time  // Can handle NULL
}

// Scanning pattern
var patentExpiry *time.Time
err := row.Scan(..., &patentExpiry, ...)
cultivar.PatentExpiry = patentExpiry  // nil if NULL
```

**Common Nullable Fields**:
- Optional dates: `*time.Time`
- Optional foreign keys: `*string`, `*int`
- Optional enums: `*string`

#### 14. Helper Method scanX() Must Match Query Columns
**Problem**: Reusable scan helpers hardcode column list, breaks when query changes
```go
// Repository has multiple query methods
func (r *Repository) FindByID(...) { /* SELECT with 8 columns */ }
func (r *Repository) FindByName(...) { /* SELECT with 8 columns */ }
func (r *Repository) scanEntities(rows *sql.Rows) { /* Scans 7 columns - BUG! */ }

// CORRECT - Ensure scanX() helpers match ALL queries using them
func (r *Repository) scanEntities(rows *sql.Rows) ([]*Entity, error) {
    // Must scan EXACTLY the columns every query SELECTs
    err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt, &e.UpdatedAt, ...) // All 8 fields
}
```

**Verification**:
1. Grep all queries that call the scan helper: `rg "scanEntities"`
2. Ensure every SELECT has identical column list
3. Update scan helper when adding entity fields

**Code Review Checklist for Repositories**:
- [ ] Every entity field appears in FindByID SELECT
- [ ] Every entity field has Scan() destination in FindByID
- [ ] Create includes all fields (including created_at, updated_at)
- [ ] Update includes all mutable fields plus updated_at
- [ ] Scan helper methods match ALL queries using them
- [ ] Nullable fields use pointer types in entity struct
- [ ] ExecContext parameter count matches INSERT/UPDATE placeholders
- [ ] Time.Now() called for timestamps (not hardcoded values)

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

### Critical Integration Testing Gotchas

#### 15. Always Use UUIDs for Language/Country IDs in Tests
**Problem**: Tests using ISO codes ("en", "es", "US") fail with "invalid input syntax for type uuid"
**Root Cause**: Database foreign keys use UUIDs, not ISO code strings
```go
// WRONG - Using ISO codes directly
plant, err := repo.FindByID(ctx, plantID, "en", nil)
// ERROR: pq: invalid input syntax for type uuid: "en"

// CORRECT - Use UUID from seeded test data
englishID := "550e8400-e29b-41d4-a716-446655440001"
plant, err := repo.FindByID(ctx, plantID, englishID, nil)
```
**Fix Pattern**:
1. Define language/country UUIDs as constants in test setup
2. Use `SeedTestLanguages()` and `SeedTestCountries()` with known UUIDs
3. Reference these UUIDs throughout tests
**Discovered**: 2025-10-02 - Fixed all 11 integration tests + 4 benchmarks

#### 16. Test Database Schema Must Match Production Schema EXACTLY
**Problem**: Column name mismatches cause test failures
**Common Mismatches Found**:
```go
// WRONG column names
languages.iso_639_1_code  // Actual: language_code
countries.iso_code        // Actual: country_code
plant_families.updated_at // Doesn't exist (no updated_at on families/genera/species)

// CORRECT - Match actual schema
INSERT INTO languages (language_id, language_code, language_name, is_active)
INSERT INTO countries (country_id, country_code, country_name, climate_systems)
INSERT INTO plant_families (family_id, family_name, created_at)
```
**Prevention**:
1. ✅ Read actual migration files before writing seed functions
2. ✅ Run `\d table_name` in psql to verify column names
3. ✅ Check for required fields (e.g., `climate_systems TEXT[] NOT NULL`)
4. ❌ Never assume column names from entity struct field names

#### 17. Integration Test Database Cleanup: DROP SCHEMA CASCADE
**Problem**: `DROP TABLE IF EXISTS` fails with "cannot drop table X because other objects depend on it"
**Solution**: Drop entire schema with CASCADE, then recreate
```go
func cleanDatabase(ctx context.Context, db *sql.DB, t *testing.T) error {
    // Nuclear option - drop everything
    _, err := db.ExecContext(ctx, `
        DROP SCHEMA public CASCADE;
        CREATE SCHEMA public;
        GRANT ALL ON SCHEMA public TO plant_api_test;
        GRANT ALL ON SCHEMA public TO public;
    `)

    // Re-enable PostGIS (lost in schema drop)
    _, err = db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS postgis;`)

    return err
}
```
**Why This Works**:
- ✅ Removes all tables, views, functions, sequences in dependency order
- ✅ No need to track 50+ table dependencies manually
- ✅ Works regardless of schema changes
- ✅ Fast - single command vs many DROP TABLE calls
**Tradeoff**: Must recreate PostGIS extension after schema drop
**See**: ADR-013 for full rationale

#### 18. Integration Tests Need Minimal Schema, Not Full Production
**Problem**: Migration 005 (localization) references tables from Part 4 (not yet implemented)
**Solution**: Create only the tables tests actually need
```go
// In test_helpers.go - create minimal localization table
_, err := db.ExecContext(ctx, `
    CREATE TABLE IF NOT EXISTS plant_common_names (
        plant_id UUID NOT NULL REFERENCES plants(plant_id) ON DELETE CASCADE,
        language_id UUID NOT NULL REFERENCES languages(language_id),
        country_id UUID REFERENCES countries(country_id),
        common_name VARCHAR(200) NOT NULL,
        is_primary BOOLEAN DEFAULT false,
        created_at TIMESTAMP DEFAULT NOW(),
        PRIMARY KEY (plant_id, language_id, common_name)
    );

    CREATE INDEX IF NOT EXISTS idx_plant_common_names_lookup
    ON plant_common_names(plant_id, language_id, country_id);
`)
```
**Rationale**:
- ✅ Tests isolated from unimplemented features
- ✅ Faster test execution (fewer tables)
- ✅ Clear test requirements (self-documenting)
- ❌ Schema duplication (also in migration file)
**When to Use**: Only when production migrations have unimplemented dependencies
**See**: ADR-015 for full rationale

#### 19. Docker Compose Test Database Image Must Include PostGIS
**Problem**: `postgres:17` doesn't include PostGIS extension
```yaml
# WRONG - Missing PostGIS
services:
  postgres-test:
    image: postgres:17
    command: shared_preload_libraries=postgis-3  # ❌ FAILS

# CORRECT - Use official PostGIS image
services:
  postgres-test:
    image: postgis/postgis:17-3.5  # ✅ PostGIS pre-installed
    environment:
      POSTGRES_DB: plantdb_test
```
**Error Signs**:
- "could not access file postgis-3: No such file or directory"
- "extension postgis is not available"
**Fix**: Change to `postgis/postgis:17-3.5` (PostgreSQL 17 + PostGIS 3.5)

#### 20. Invalid UUID Formats Cause Cryptic Errors
**Problem**: Typos in test UUIDs cause "invalid input syntax" errors
```go
// WRONG - UUIDs can't start with 'g' or 's' (not hex)
genusID := "g50e8400-e29b-41d4-a716-446655440001"   // ❌ 'g' not hex
speciesID := "s50e8400-e29b-41d4-a716-446655440001" // ❌ 's' not hex

// CORRECT - First character must be 0-9 or a-f
genusID := "150e8400-e29b-41d4-a716-446655440001"   // ✅ Valid hex
speciesID := "250e8400-e29b-41d4-a716-446655440001" // ✅ Valid hex
```
**Validation**: UUIDs must be 36 characters: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` (0-9, a-f only)
**Prevention**: Use UUID constants defined once at top of test file

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

### Critical Query Filter Gotchas

#### 21. Zone Matching Requires Exact Boundaries (Not LIKE)
**Problem**: Using LIKE for comma-separated zone lists causes false matches
```go
// WRONG - "1a" matches "11a" due to substring match
query += " AND hardiness_zones LIKE '%1a%'"

// CORRECT - Use regex with word boundaries
query += " AND hardiness_zones ~ '(^|,)1a(,|$)'"
```
**Why**: Hardiness zones are stored as comma-separated strings: `"5a,5b,6a,6b,7a,7b"`
- LIKE `'%1a%'` matches both "1a" and "11a"
- Regex `(^|,)1a(,|$)` only matches "1a" at start, end, or between commas

**Rule**: For comma-separated enum lists, always use PostgreSQL regex `~` operator with boundary anchors

#### 22. Height Filter Logic Must Match User Intent
**Problem**: Ambiguous filter semantics cause wrong results
```go
// WRONG - OR logic is too permissive
WHERE (pc.mature_height).typical_m >= $1 OR (pc.mature_height).max_m >= $1  // MinHeight
WHERE (pc.mature_height).typical_m <= $1 OR (pc.mature_height).min_m <= $1  // MaxHeight

// CORRECT - Match user's actual intent
WHERE (pc.mature_height).max_m >= $1      // MinHeight: can the plant grow this tall?
WHERE (pc.mature_height).typical_m <= $1  // MaxHeight: will the plant fit in this space?
```

**User Intent**:
- **MinHeight filter** (e.g., 5m): User wants plants that CAN grow at least 5m tall → use `max_m`
- **MaxHeight filter** (e.g., 2m): User wants plants that typically fit in 2m space → use `typical_m`

**Test Case**: With plants at heights 0.2m, 2.0m, 15.0m:
- `MaxHeight = 1.0` should return ONLY 0.2m plant (thyme)
- Using OR would incorrectly return 2.0m plant (rhododendron has min=1.0)

#### 23. Physical Characteristics Need LEFT JOIN (Not INNER)
**Problem**: INNER JOIN excludes plants without physical data
```go
// WRONG - Only returns plants with physical_characteristics rows
INNER JOIN physical_characteristics pc ON p.plant_id = pc.plant_id

// CORRECT - Returns all plants, filters only when characteristic specified
LEFT JOIN physical_characteristics pc ON p.plant_id = pc.plant_id
```
**Why**: Not all plants have physical characteristics data yet. LEFT JOIN allows:
- Search without filters: returns all plants
- Search with filters: only matches plants with that data
- Graceful degradation as database grows

#### 24. Enum Array Insertion Requires Type Casting
**Problem**: PostgreSQL rejects text arrays for enum array columns
```sql
-- WRONG - Type mismatch error
INSERT INTO growing_conditions_assertions (sun_requirements, ...)
VALUES (ARRAY['full_sun', 'partial_sun'], ...)
-- ERROR: column "sun_requirements" is of type sun_requirement[] but expression is of type text[]

-- CORRECT - Explicit cast to enum type
VALUES (ARRAY['full_sun', 'partial_sun']::sun_requirement[], ...)
```
**Rule**: Always cast array literals when inserting into enum array columns

#### 25. JSONB Boolean Queries Require Type Casting
**Problem**: JSONB values are text, not boolean
```go
// WRONG - Treats "true" as text, not boolean
WHERE pc.traits->>'evergreen' = true  // Fails: comparing text to boolean

// CORRECT - Cast JSONB text to boolean
WHERE (pc.traits->>'evergreen')::boolean = true
```
**Pattern**: For all JSONB boolean fields, use `(jsonb_col->>'field')::boolean`

**Common JSONB Traits**:
- `(pc.traits->>'evergreen')::boolean`
- `(pc.traits->>'deciduous')::boolean`
- `(pc.traits->>'toxic')::boolean`

#### 26. Integration Test Schema Column Names Must Match Production
**Problem**: Using entity field names instead of actual database column names
```go
// WRONG - Assumes columns match entity field names
INSERT INTO physical_characteristics (mature_width, ...) -- Column doesn't exist!

// CORRECT - Check migration files for actual column names
INSERT INTO physical_characteristics (mature_spread, ...) -- Actual column name
```
**Prevention**:
1. Always read migration files before writing seed functions
2. Run `\d table_name` in psql to verify column names
3. Check for required vs optional fields (e.g., `NOT NULL`, `DEFAULT`)

**Common Mismatches Encountered**:
- `mature_width` → `mature_spread` (physical_characteristics)
- `iso_639_1_code` → `language_code` (languages)
- `iso_code` → `country_code` (countries)

### Critical Performance & Infrastructure Gotchas

#### 27. GIN Trigram Indexes Required for ILIKE Searches
**Problem**: ILIKE queries on text fields are 100-1000x slower without GIN indexes
**Context**: Multi-language common name searches use `ILIKE '%search%'` pattern
```sql
-- WRONG - B-tree indexes don't help ILIKE queries
CREATE INDEX idx_common_name ON plant_common_names(common_name); -- B-tree (default)
SELECT * FROM plant_common_names WHERE common_name ILIKE '%rose%'; -- Seq scan, slow!

-- CORRECT - GIN trigram index supports ILIKE
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_common_name_trgm ON plant_common_names USING GIN (common_name gin_trgm_ops);
SELECT * FROM plant_common_names WHERE common_name ILIKE '%rose%'; -- Index scan, fast!
```
**Performance Impact**:
- Without index: 500ms for 10K rows (sequential scan)
- With GIN index: 5ms for 10K rows (100x improvement)
**Trade-off**: Index size ~3x larger than B-tree (15% database size increase)
**Rule**: Always use GIN trigram indexes for any column with ILIKE searches
**See**: Migration 006, ADR-016

#### 28. Prepared Statements Must Handle Connection Pool Resets
**Problem**: Prepared statements become invalid when connection pool recycles connections
```go
// WRONG - Prepared statement breaks after connection reset
stmt, _ := db.Prepare("SELECT * FROM plants WHERE plant_id = $1")
defer stmt.Close()
// Later: ERROR: prepared statement "stmtcache_1" does not exist

// CORRECT - Fallback to regular query on prepared statement failure
stmt, err := stmtMgr.Get(StmtFindPlantByID)
if err != nil {
    // Prepared statement not available, use regular query
    return db.QueryRowContext(ctx, "SELECT * FROM plants WHERE plant_id = $1", plantID)
}
return stmt.QueryRowContext(ctx, plantID)
```
**Why**: Connection pooling can close/reopen connections, invalidating prepared statements
**Solution**: `PreparedStatementManager` with `Recreate()` method to rebuild on connection reset
**Best Practice**: Always provide fallback to regular query if prepared statement unavailable
**See**: `prepared_statements.go`, ADR-017

#### 29. Lookup Table Cache Keys Must Include All Query Parameters
**Problem**: Cache key doesn't match query parameters, causes wrong data return
```go
// WRONG - Language-specific data cached with language-agnostic key
key := fmt.Sprintf("family:%s", familyID)  // Missing language!
// User requests Spanish, gets cached English data

// CORRECT - Include all query parameters in cache key
key := fmt.Sprintf("family:%s:lang:%s", familyID, languageID)
if countryID != nil {
    key = fmt.Sprintf("family:%s:lang:%s:country:%s", familyID, languageID, *countryID)
}
```
**Cache Invalidation Pattern**:
```go
// When updating family, delete ALL language/country variants
pattern := fmt.Sprintf("family:%s:*", familyID)
cache.DeletePattern(ctx, pattern) // Uses Redis SCAN, not KEYS (blocking)
```
**Rule**: Cache key MUST include every parameter that affects query results
**Common Parameters**: `language_id`, `country_id`, `limit`, `offset` (for pagination)
**See**: `lookup_cache.go`, ADR-018

#### 30. Go Generics Have Zero Runtime Overhead But Require Go 1.18+
**Problem**: Attempting to use generics with Go 1.17 or earlier causes compilation errors
```go
// Generic function (requires Go 1.18+)
func ScanRows[T any](rows *sql.Rows, scanFunc func(*sql.Rows) (*T, error)) ([]*T, error) {
    // ERROR with Go 1.17: syntax error: unexpected [, expecting (
}
```
**Verification**:
```bash
go version  # Must be >= go1.18
go mod edit -go=1.18  # Update go.mod
```
**Benefits**:
- Type safety at compile time (catches errors early)
- Zero runtime overhead (generics compile to concrete types)
- Code reuse: `ScanRows[Plant]`, `ScanRows[Family]` from same function
**Trade-off**: Requires modern Go version (1.18+, released March 2022)
**See**: `scanner.go`, ADR-019

#### 31. Transaction Panics Must Be Recovered to Prevent Data Corruption
**Problem**: Panic during transaction leaves database in inconsistent state
```go
// WRONG - Panic skips rollback, partial data committed
tx, _ := db.BeginTx(ctx, nil)
_, err := tx.Exec("INSERT INTO plants ...") // Succeeds
// Code panics here (nil pointer, etc.)
tx.Commit() // Never reached, transaction left open

// CORRECT - Defer handles rollback even on panic
tx, _ := db.BeginTx(ctx, nil)
defer func() {
    if p := recover(); p != nil {
        _ = tx.Rollback()
        panic(p) // Re-panic after rollback
    }
}()
_, err := tx.Exec("INSERT INTO plants ...")
if err != nil {
    tx.Rollback()
    return err
}
return tx.Commit()
```
**Why**: Panics bypass normal control flow, defer ensures cleanup
**TxManager Pattern**: Encapsulates this logic, prevents forgetting rollback
```go
return txManager.WithTransaction(ctx, func(tx *sql.Tx) error {
    // Operations here
    // Automatic rollback on error or panic
})
```
**See**: `transaction.go`, ADR-020

#### 32. GeoJSON Validation MUST Happen Before ST_GeomFromGeoJSON
**Problem**: Invalid GeoJSON reaches PostGIS, causes cryptic database errors
```go
// WRONG - Malformed GeoJSON causes database error
_, err := db.Exec(`
    INSERT INTO countries (country_boundary)
    VALUES (ST_GeomFromGeoJSON($1))
`, malformedGeoJSON)
// ERROR: invalid GeoJSON representation (SQLSTATE 22023) - unhelpful!

// CORRECT - Validate at application layer first
if err := ValidateGeoJSON(geojsonStr); err != nil {
    return fmt.Errorf("invalid boundary geojson: %w", err)
    // ERROR: missing 'coordinates' field - clear error message!
}
_, err := db.Exec(`INSERT INTO countries (country_boundary) VALUES (ST_GeomFromGeoJSON($1))`, geojsonStr)
```
**Validation Checklist**:
1. ✅ Valid JSON structure
2. ✅ Has `type` field (Point, Polygon, MultiPolygon, etc.)
3. ✅ Has `coordinates` array
4. ✅ Polygon rings are closed (first point == last point)
5. ✅ No self-intersections (for Polygons)
**Security Benefit**: Prevents malicious GeoJSON from exploiting PostGIS vulnerabilities
**Performance Benefit**: Fail-fast before expensive database query
**See**: `geojson_validator.go`, `coordinates_validator.go`, ADR-021

#### 33. Coordinate Bounds Validation Prevents Spatial Query Failures
**Problem**: Out-of-bounds coordinates cause silent failures or incorrect results
```go
// WRONG - Invalid coordinates accepted, spatial query returns wrong results
lat, lng := 95.5, -200.0  // Invalid! (lat > 90, lng < -180)
country, _ := repo.FindByPoint(ctx, lat, lng)
// Returns nil or wrong country due to coordinate wrapping

// CORRECT - Validate WGS84 bounds before spatial query
if err := ValidateCoordinates(lat, lng); err != nil {
    return nil, fmt.Errorf("invalid coordinates: %w", err)
    // ERROR: latitude must be between -90 and 90 degrees, got: 95.5
}
country, _ := repo.FindByPoint(ctx, lat, lng)
```
**WGS84 Coordinate Bounds**:
- Latitude: -90.0 to 90.0 (South Pole to North Pole)
- Longitude: -180.0 to 180.0 (International Date Line)
**Common Mistakes**:
- Swapping lat/lng (using lat=151.2, lng=-33.8 instead of lat=-33.8, lng=151.2)
- Using degrees/minutes/seconds instead of decimal degrees
**PostGIS Behavior**: Out-of-bounds coords may wrap (lng=370 becomes lng=10) causing silent errors
**See**: `coordinates_validator.go`, ADR-021

#### 34. Savepoints Enable Partial Rollback in Complex Transactions
**Problem**: Multi-step transaction needs to rollback one step without losing all progress
**Use Case**: Import plant data - if growing conditions invalid, keep plant but rollback conditions
```go
// Without savepoints: All-or-nothing (plant + conditions both fail)
return txManager.WithTransaction(ctx, func(tx *sql.Tx) error {
    tx.Exec("INSERT INTO plants ...")           // Step 1
    tx.Exec("INSERT INTO growing_conditions...") // Step 2 fails
    // Both steps rolled back
})

// With savepoints: Partial rollback (keep plant, rollback conditions)
return txManager.WithTransaction(ctx, func(tx *sql.Tx) error {
    sm := NewSavepointManager(tx)

    tx.Exec("INSERT INTO plants ...")  // Step 1 - committed

    sm.WithSavepoint(ctx, "conditions", func() error {
        err := tx.Exec("INSERT INTO growing_conditions...")
        if err != nil {
            return err // Rolls back to savepoint, Step 1 retained
        }
        return nil
    })
    // Plant saved, conditions skipped if invalid
})
```
**PostgreSQL Savepoint Commands**:
- `SAVEPOINT name` - Create checkpoint
- `ROLLBACK TO SAVEPOINT name` - Revert to checkpoint
- `RELEASE SAVEPOINT name` - Commit checkpoint
**Performance Note**: Savepoints are lightweight (microseconds), use liberally
**See**: `transaction.go` SavepointManager, ADR-020

#### 35. Repository Struct Naming Must Be PascalCase for Godoc
**Problem**: Lowercase struct names don't appear in public API documentation
```go
// WRONG - Not exported, invisible to godoc
type postgresPlantRepository struct {
    db *sql.DB
}

func NewPostgresPlantRepository(db *sql.DB) repository.PlantRepository {
    return &postgresPlantRepository{db: db} // Returns unexported type
}

// CORRECT - Exported struct, visible in godoc
type PostgresPlantRepository struct {
    db *sql.DB
}

func NewPostgresPlantRepository(db *sql.DB) repository.PlantRepository {
    return &PostgresPlantRepository{db: db}
}
```
**Why PascalCase Matters**:
1. Exported structs appear in godoc (API documentation)
2. Allows external packages to reference concrete type (for testing, debugging)
3. Consistent with Go naming conventions (exported identifiers start with uppercase)
**Rule**: All repository, service, and manager structs should be PascalCase
**Discovered**: Code review found 9 repositories with lowercase names (all fixed)

### Critical PostGIS Spatial Gotchas

#### 36. Avoid ST_AsText When Extracting Coordinates
**Problem**: Using `ST_AsText()` then casting back to geometry is inefficient and error-prone
```sql
-- WRONG - Converts to WKT text then back to geometry (3x slower)
SELECT ST_Y(ST_AsText(ST_Centroid(boundary))::geometry) as lat

-- CORRECT - Direct coordinate extraction
SELECT ST_Y(ST_Centroid(boundary)) as lat
```
**Why**: `ST_AsText()` creates an intermediate WKT string ("POINT(-122.4 37.7)") which must be parsed back
**Performance**: 3x slower, unnecessary string allocation
**When to Use ST_AsText**: Only when displaying geometry to users as text
**Discovered**: PostGIS expert review (Part 3)

#### 37. GIST Indexes Are REQUIRED for Spatial Queries
**Problem**: Spatial queries without GIST indexes perform full table scans (100-1000x slower)
```sql
-- Without index: Sequential scan, 500ms for 10K rows
SELECT * FROM gardens WHERE ST_DWithin(location, point, 5000);

-- With GIST index: Index scan, 5ms for 10K rows
CREATE INDEX idx_gardens_location ON gardens USING GIST(location);
```
**Rule**: Create GIST index on EVERY geometry/geography column
**Index Types**:
- ✅ `USING GIST` - Only type that supports spatial operators (&&, ST_Contains, ST_DWithin)
- ❌ `USING BTREE` - Does NOT work for spatial queries
- ❌ `USING HASH` - Does NOT work for spatial queries
**Composite Indexes**:
```sql
-- Combine spatial + non-spatial for common patterns
CREATE INDEX idx_gardens_user_boundary ON gardens USING GIST(user_id, boundary);
```
**Performance Impact**: 100x improvement on typical spatial queries
**Discovered**: Part 3 implementation, Migration 007

#### 38. Geography vs Geometry for Distance Calculations
**Problem**: Using `geometry` type for distances gives meaningless results in degrees
```sql
-- WRONG - Distance in degrees (meaningless to users)
SELECT ST_Distance(point1, point2);  -- Returns 0.045

-- CORRECT - Distance in meters (what users expect)
SELECT ST_Distance(point1::geography, point2::geography);  -- Returns 5000.0
```
**Rule**: Always cast to `::geography` for distance/area calculations
**Schema Design**:
```sql
-- Boundaries: geometry (faster for polygon operations)
boundary GEOMETRY(Polygon, 4326)

-- Locations: geography (accurate for distance)
location GEOGRAPHY(Point, 4326)
```
**Trade-off**: Geography queries are ~10% slower but meter-accurate globally
**Common Functions**:
- `ST_Distance(geog1, geog2)` - Returns meters
- `ST_DWithin(geog1, geog2, meters)` - Radius search
- `ST_Area(geom::geography)` - Returns square meters
**Discovered**: Part 3 implementation (ADR-024)

#### 39. Bounding Box Pre-filtering for Complex Polygons
**Problem**: ST_Contains on complex polygons is expensive without pre-filter
```sql
-- SLOW - ST_Contains checks every vertex
SELECT * FROM zones WHERE ST_Contains(geometry, point);

-- FAST - Bounding box check first (uses GIST index)
SELECT * FROM zones
WHERE geometry && point                      -- Fast bbox check
  AND ST_Contains(geometry, point);          -- Precise check
```
**Explanation**:
- `&&` operator checks if bounding boxes overlap (GIST index, very fast)
- `ST_Contains` does precise point-in-polygon test (expensive)
**Performance**: 2-5x faster for complex polygons (10+ vertices)
**When to Use**: Any ST_Contains, ST_Intersects, ST_Within on multi-vertex polygons
**Discovered**: PostGIS expert review

#### 40. NULL Handling in Spatial Queries Requires Go-side Logic
**Problem**: Go `*string` nil doesn't translate to SQL NULL correctly in prepared statements
```sql
-- This doesn't work as expected with Go nil pointer
AND ($2::uuid IS NULL OR zone_id != $2)
```
**Correct Pattern**:
```go
// Build query dynamically based on nil check
query := "SELECT ... WHERE garden_id = $1"
args := []interface{}{gardenID}

if excludeZoneID != nil && *excludeZoneID != "" {
    query += " AND zone_id != $2"
    args = append(args, *excludeZoneID)
}

db.QueryContext(ctx, query, args...)
```
**Why**: `database/sql` doesn't handle Go nil pointers as SQL NULL in typed casts
**Alternative**: Use `sql.NullString` type
```go
var exclude sql.NullString
if excludeZoneID != nil {
    exclude = sql.NullString{String: *excludeZoneID, Valid: true}
}
db.QueryContext(ctx, query, gardenID, exclude)
```
**Discovered**: PostGIS expert review (CheckZoneOverlaps issue)

#### 41. Validate GeoJSON Structure Before ST_GeomFromGeoJSON
**Problem**: Invalid GeoJSON passed to PostGIS causes cryptic errors
```go
// WRONG - Database error is cryptic
_, err := db.Exec(`INSERT INTO gardens (boundary) VALUES (ST_GeomFromGeoJSON($1))`, badJSON)
// ERROR: invalid GeoJSON representation (SQLSTATE 22023)

// CORRECT - Validate first for clear error messages
if err := ValidateGeoJSON(geojson); err != nil {
    return fmt.Errorf("invalid boundary: %w", err)
    // ERROR: missing 'coordinates' field
}
_, err := db.Exec(`INSERT INTO gardens (boundary) VALUES (ST_GeomFromGeoJSON($1))`, geojson)
```
**Validation Checklist**:
1. ✅ Valid JSON structure
2. ✅ Has `type` field (Point, Polygon, MultiPolygon, etc.)
3. ✅ Has `coordinates` array
4. ✅ Polygon rings are closed (first point == last point)
5. ✅ Coordinates within bounds (-90 to 90 lat, -180 to 180 lng)
**Implementation**: `geojson_validator.go`, `coordinates_validator.go`
**Performance**: <0.1ms overhead, prevents expensive failed database queries
**Discovered**: Part 2 & 3 implementation (Gotcha #32, ADR-021)

#### 42. Polygon Closure Validation Prevents Subtle Bugs
**Problem**: Unclosed polygon rings cause incorrect spatial calculations
```json
// WRONG - Unclosed polygon (first != last)
{
  "type": "Polygon",
  "coordinates": [[
    [-122.5, 37.7],
    [-122.4, 37.7],
    [-122.4, 37.8],
    [-122.5, 37.8]
  ]]
}

// CORRECT - Closed polygon
{
  "type": "Polygon",
  "coordinates": [[
    [-122.5, 37.7],
    [-122.4, 37.7],
    [-122.4, 37.8],
    [-122.5, 37.8],
    [-122.5, 37.7]  // Same as first point
  ]]
}
```
**Validation**:
```go
func ValidatePolygonClosure(geojsonStr string) error {
    // Extract coordinates array
    ring := coordinates[0]  // First ring
    first := ring[0]
    last := ring[len(ring)-1]

    if first[0] != last[0] || first[1] != last[1] {
        return fmt.Errorf("polygon ring is not closed: first point %v != last point %v", first, last)
    }
    return nil
}
```
**Impact**: Unclosed polygons may:
- Calculate incorrect area (ST_Area)
- Fail containment checks (ST_Contains)
- Cause rendering issues in mapping tools
**Discovered**: Part 3 GeoJSON validator implementation

#### 43. MultiPolygon vs Polygon Requires Different Validation
**Problem**: MultiPolygon has different coordinate structure than Polygon
```json
// Polygon: [][][]float64
{"type": "Polygon", "coordinates": [[[lng, lat], [lng, lat], ...]]}

// MultiPolygon: [][][][]float64
{"type": "MultiPolygon", "coordinates": [[[[lng, lat], ...]], [[[lng, lat], ...]]]}
```
**Validation Pattern**:
```go
switch geom.Type {
case "Polygon":
    return ValidatePolygonClosure(coordinates)  // Single ring array
case "MultiPolygon":
    for _, polygon := range coordinates {        // Iterate each polygon
        return ValidatePolygonClosure(polygon)
    }
}
```
**Common Mistake**: Treating MultiPolygon coordinates as Polygon (causes index out of bounds)
**Discovered**: PostGIS expert review (geojson_validator.go issue)

#### 44. Service Layer Should NOT Perform Database Queries
**Problem**: Service layer querying database directly bypasses repository abstraction
```go
// WRONG - Service querying database
func (s *GardenService) CreateGarden(...) {
    db.QueryContext(ctx, "SELECT COUNT(*) FROM gardens WHERE user_id = $1", userID)
}

// CORRECT - Service calls repository method
func (s *GardenService) CreateGarden(...) {
    count, err := s.gardenRepo.CountByUserID(ctx, userID)
}
```
**Layer Responsibilities**:
- **Repository**: All database queries (SELECT, INSERT, UPDATE, DELETE)
- **Service**: Business logic, validation, orchestration
- **API**: HTTP/GraphQL handling, request/response serialization
**Benefits**:
- ✅ Service tests use mocks (no database)
- ✅ Repository interface can be swapped (PostgreSQL → MySQL)
- ✅ Clear separation of concerns
**Discovered**: Part 3 service layer implementation (ADR-028)

#### 45. Transaction Support Must Be Repository Method, Not Service
**Problem**: Service layer managing transactions couples it to SQL implementation
```go
// WRONG - Service starting transaction
func (s *GardenService) CreateGarden(...) {
    tx, _ := s.db.BeginTx(ctx, nil)  // Service has *sql.DB reference
    defer tx.Rollback()
}

// CORRECT - Repository provides transaction methods
func (s *GardenService) CreateGarden(...) {
    return s.gardenRepo.CreateWithTransaction(ctx, garden, func(tx *sql.Tx) error {
        // Operations using tx
    })
}
```
**Alternative Pattern**: Transaction manager utility
```go
// Repository uses shared transaction manager
func (r *PostgresGardenRepository) CreateGarden(ctx, garden) error {
    return r.txManager.WithTransaction(ctx, func(tx *sql.Tx) error {
        // Multi-step operations
    })
}
```
**Benefits**:
- ✅ Service layer remains database-agnostic
- ✅ Transaction logic centralized in infrastructure layer
- ✅ Panic recovery handled consistently
**Discovered**: Part 3 bulk operations (ADR-027, ADR-020)

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