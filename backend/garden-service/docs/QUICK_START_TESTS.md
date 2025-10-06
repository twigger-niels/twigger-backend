# Quick Start - Garden Service Integration Tests

This guide helps you run the Garden Service integration tests to validate PostGIS spatial queries.

## Prerequisites

1. **PostgreSQL 17+ with PostGIS 3.5** installed and running
2. **Go 1.21+** installed
3. **PostgreSQL user** with database creation privileges (default: `postgres`)

## Quick Start (Windows)

```bash
# Run from repository root
.\scripts\run-garden-integration-tests.bat
```

## Quick Start (Linux/Mac)

```bash
# Run from repository root
./scripts/run-garden-integration-tests.sh
```

## Manual Setup

If you prefer to run tests manually:

### 1. Create Test Database

```bash
psql -U postgres -c "CREATE DATABASE plantdb_test;"
psql -U postgres -d plantdb_test -c "CREATE EXTENSION IF NOT EXISTS postgis;"
psql -U postgres -d plantdb_test -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
```

### 2. Set Environment Variable

```bash
# Windows
set TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/plantdb_test?sslmode=disable

# Linux/Mac
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/plantdb_test?sslmode=disable"
```

### 3. Run Tests

```bash
cd backend/garden-service
go test -v -tags=integration ./infrastructure/persistence/...
```

## What Gets Tested

### GardenRepository (15 test suites)

| Test | PostGIS Function | What It Validates |
|------|------------------|-------------------|
| **Create** | ST_GeomFromGeoJSON | GeoJSON to PostGIS conversion |
| **Create_InvalidGeoJSON** | Validation | Rejects unclosed polygons |
| **FindByID** | ST_AsGeoJSON | PostGIS to GeoJSON conversion |
| **FindByID_NotFound** | Error handling | NotFoundError type |
| **FindByUserID** | Ordering | created_at DESC ordering |
| **FindByUserID_Pagination** | LIMIT/OFFSET | Pagination correctness |
| **Update** | Timestamp handling | updated_at updates |
| **Delete** | CASCADE | Foreign key cascade |
| **FindByLocation** | ST_DWithin | Radius search (5km) |
| **CalculateArea** | ST_Area | Square meter calculation |
| **DetectHardinessZone** | ST_Contains | Spatial join with climate_zones |
| **ValidateBoundary** | Validation | GeoJSON validation |
| **CountByUserID** | COUNT | Aggregation |
| **GetTotalArea** | SUM + ST_Area | Area aggregation |
| **GetCenterPoint** | ST_Centroid | Center point calculation |

## Expected Output

```
=== RUN   TestGardenRepository_Create
--- PASS: TestGardenRepository_Create (0.05s)
=== RUN   TestGardenRepository_Create_InvalidGeoJSON
--- PASS: TestGardenRepository_Create_InvalidGeoJSON (0.01s)
=== RUN   TestGardenRepository_FindByID
--- PASS: TestGardenRepository_FindByID (0.02s)
...
=== RUN   TestGardenRepository_GetTotalArea
--- PASS: TestGardenRepository_GetTotalArea (0.03s)
PASS
ok      twigger-backend/backend/garden-service/infrastructure/persistence       2.345s
```

## Performance Benchmarks

Expected performance with GIST indexes:

| Operation | Target | Typical |
|-----------|--------|---------|
| Create with GeoJSON | <20ms | 10-15ms |
| FindByID | <10ms | 5-8ms |
| FindByLocation (5km) | <50ms | 30-40ms |
| CalculateArea | <10ms | 5-8ms |
| DetectHardinessZone | <20ms | 15-20ms |

## Troubleshooting

### Error: "PostgreSQL is not running"

**Solution**: Start PostgreSQL service
```bash
# Windows
net start postgresql-x64-17

# Linux
sudo systemctl start postgresql

# Mac
brew services start postgresql
```

### Error: "extension postgis is not available"

**Solution**: Install PostGIS extension
```bash
# Ubuntu/Debian
sudo apt-get install postgresql-17-postgis-3

# Mac
brew install postgis

# Windows
# Download and install from: https://postgis.net/windows_downloads/
```

### Error: "invalid input syntax for type uuid"

**Cause**: Test is using ISO codes ("en", "US") instead of UUIDs

**Solution**: Check test is using UUIDs from `SeedTestUsers()`:
```go
userID := "550e8400-e29b-41d4-a716-446655440001" // Correct
userID := "test-user-1" // Wrong!
```

### Error: "pq: relation 'gardens' does not exist"

**Cause**: Test schema not created

**Solution**: Ensure `CreateTestSchema()` is called in test setup:
```go
err := testhelpers.CreateTestSchema(ctx, db, t)
require.NoError(t, err)
```

### Tests Pass Locally But Fail in CI

**Cause**: Different PostgreSQL/PostGIS versions

**Solution**: Verify versions match:
```bash
psql -U postgres -c "SELECT version();"
psql -U postgres -c "SELECT PostGIS_Version();"
```

Expected:
- PostgreSQL 17.x
- PostGIS 3.5.x

## Test Data

Tests use predefined GeoJSON from `test_helpers.go`:

### Valid Garden Boundary (San Francisco)
```json
{
  "type": "Polygon",
  "coordinates": [[
    [-122.4194, 37.7749],
    [-122.4184, 37.7749],
    [-122.4184, 37.7739],
    [-122.4194, 37.7739],
    [-122.4194, 37.7749]
  ]]
}
```

### Valid Garden Location (SF Center)
```json
{
  "type": "Point",
  "coordinates": [-122.4194, 37.7749]
}
```

### Test Users
- `550e8400-e29b-41d4-a716-446655440001` - testuser1
- `550e8400-e29b-41d4-a716-446655440002` - testuser2
- `550e8400-e29b-41d4-a716-446655440003` - testuser3

### Test Climate Zone
- Zone System: USDA
- Zone Code: 10a
- Geometry: Covers San Francisco Bay Area

## Verifying Spatial Indexes

Run this query to verify GIST indexes exist:

```sql
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE indexname LIKE 'idx_garden%'
ORDER BY tablename, indexname;
```

Expected indexes:
- `idx_gardens_boundary` (GIST on boundary)
- `idx_gardens_location` (GIST on location)
- `idx_garden_zones_geometry` (GIST on geometry)
- `idx_garden_features_geometry` (GIST on geometry)
- `idx_garden_plants_location` (GIST on location)

## Next Steps

After all tests pass:

1. ✅ Validate spatial query performance
2. ✅ Review `SPATIAL_QUERIES.md` for query patterns
3. 📋 Implement remaining repositories (GardenFeature, GardenPlant)
4. 📋 Write additional integration tests
5. 📋 Implement service layer

## Documentation

- **Spatial Queries**: See `SPATIAL_QUERIES.md`
- **Architecture**: See `../../epics/plant-database/architecture.md`
- **Task Tracking**: See `../../epics/plant-database/tasks.md`

---

*Last Updated: 2025-10-03*
*Part 3: Garden Spatial Service*
