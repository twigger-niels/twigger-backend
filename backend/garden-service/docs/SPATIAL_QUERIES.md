# PostGIS Spatial Queries and Patterns - Garden Service

This document summarizes all PostGIS spatial operations used in the Garden Service, their purposes, required indexes, and performance considerations.

## Table of Contents
1. [Spatial Data Storage](#spatial-data-storage)
2. [Core Spatial Functions](#core-spatial-functions)
3. [Query Patterns by Repository](#query-patterns-by-repository)
4. [Required Indexes](#required-indexes)
5. [Performance Benchmarks](#performance-benchmarks)
6. [Common Pitfalls](#common-pitfalls)

---

## Spatial Data Storage

### Coordinate System: WGS84 (SRID 4326)
- **Storage**: All geometry columns use `GEOMETRY(..., 4326)` or `GEOGRAPHY(Point, 4326)`
- **Why**: WGS84 is the standard for GPS coordinates and web mapping
- **Conversion**: Use `::geography` cast for accurate distance/area calculations

### GeoJSON Format
All spatial data is stored and retrieved as GeoJSON strings:

```json
// Point (plant location, garden center)
{
  "type": "Point",
  "coordinates": [-122.4194, 37.7749]
}

// Polygon (garden boundary, zone geometry)
{
  "type": "Polygon",
  "coordinates": [
    [
      [-122.4194, 37.7749],
      [-122.4184, 37.7749],
      [-122.4184, 37.7739],
      [-122.4194, 37.7739],
      [-122.4194, 37.7749]  // Must close the ring!
    ]
  ]
}
```

**Critical**: First and last coordinates must be identical (polygon closure validation in `geojson_validator.go`)

---

## Core Spatial Functions

### 1. ST_GeomFromGeoJSON - Convert GeoJSON to PostGIS Geometry
```sql
-- Insert garden boundary from GeoJSON
INSERT INTO gardens (boundary)
VALUES (ST_GeomFromGeoJSON('{"type":"Polygon","coordinates":[...]}'));
```
**Used in**: Create, Update operations
**Validation**: Always validate GeoJSON BEFORE passing to this function (Gotcha #32)
**Performance**: <1ms per conversion

### 2. ST_AsGeoJSON - Convert PostGIS Geometry to GeoJSON
```sql
-- Retrieve garden boundary as GeoJSON
SELECT ST_AsGeoJSON(boundary) as boundary FROM gardens WHERE garden_id = $1;
```
**Used in**: All FindByID, FindByUserID, FindByLocation queries
**Performance**: <1ms per conversion
**Note**: Returns JSON string, not geometry object

### 3. ST_Area - Calculate Polygon Area
```sql
-- Calculate garden area in square meters
SELECT ST_Area(boundary::geography) FROM gardens WHERE garden_id = $1;
```
**Used in**: CalculateArea, GetTotalArea
**Cast to geography**: Required for accurate area in meters (vs degrees)
**Performance**: <10ms for complex polygons
**Generated Column**: `garden_zones.area_m2` auto-calculates area

### 4. ST_DWithin - Find Features Within Distance
```sql
-- Find gardens within 5km of a point
SELECT * FROM gardens
WHERE ST_DWithin(
    location,
    ST_SetSRID(ST_MakePoint($lng, $lat), 4326)::geography,
    5000  -- 5km in meters
)
ORDER BY ST_Distance(location, ST_SetSRID(ST_MakePoint($lng, $lat), 4326)::geography);
```
**Used in**: FindByLocation, CheckPlantSpacing
**Index**: Requires GIST index on `location` column
**Performance**: <50ms with index, 500ms+ without
**Note**: Use `::geography` for distance in meters

### 5. ST_Contains - Check if Geometry Contains Another
```sql
-- Check if zone is within garden boundary
SELECT ST_Contains(g.boundary, $zoneGeometry::geometry)
FROM gardens g
WHERE garden_id = $1;

-- Detect hardiness zone for garden
SELECT zone_code FROM climate_zones
WHERE ST_Contains(zone_geometry, $gardenBoundary);
```
**Used in**: ValidateZoneWithinGarden, DetectHardinessZone, ValidatePlantLocation
**Index**: Requires GIST index on containing geometry (garden boundary, climate zone)
**Performance**: <20ms with index
**Returns**: Boolean (true if first geometry contains second)

### 6. ST_Overlaps - Check if Geometries Overlap
```sql
-- Check if new zone overlaps existing zones
SELECT EXISTS(
    SELECT 1 FROM garden_zones
    WHERE garden_id = $1
      AND zone_id != $2  -- Exclude current zone when updating
      AND ST_Overlaps(geometry, $newZoneGeometry::geometry)
);
```
**Used in**: CheckZoneOverlaps
**Index**: Requires GIST index on `geometry` column
**Performance**: <20ms with index
**Note**: Returns false if geometries are identical or one contains the other

### 7. ST_Centroid - Calculate Center Point
```sql
-- Get center point of garden boundary
SELECT ST_AsText(ST_Centroid(boundary)) FROM gardens WHERE garden_id = $1;
```
**Used in**: GetCenterPoint
**Performance**: <5ms
**Alternative**: Use pre-calculated `location` field for better performance

### 8. ST_IsValid - Validate Geometry
```sql
-- Check if garden boundary is a valid geometry
SELECT ST_IsValid(boundary) FROM gardens WHERE garden_id = $1;
```
**Used in**: CheckBoundaryValid
**Performance**: <5ms
**Note**: Checks for self-intersections, invalid ring orientation, etc.

### 9. ST_Distance - Calculate Distance Between Geometries
```sql
-- Calculate distance between two gardens
SELECT ST_Distance(
    g1.location::geography,
    g2.location::geography
) FROM gardens g1, gardens g2
WHERE g1.garden_id = $1 AND g2.garden_id = $2;
```
**Used in**: FindByLocation (for ordering results)
**Cast to geography**: Required for distance in meters
**Performance**: <5ms per calculation

### 10. ST_MakePoint + ST_SetSRID - Create Point from Coordinates
```sql
-- Create point geometry from lat/lng
ST_SetSRID(ST_MakePoint($lng, $lat), 4326)
```
**Used in**: FindByLocation, spatial query construction
**Note**: Arguments are (longitude, latitude) - NOT (lat, lng)!
**SRID 4326**: WGS84 coordinate system

---

## Query Patterns by Repository

### GardenRepository

#### Create Garden with Boundary Validation
```sql
INSERT INTO gardens (garden_id, boundary, location, ...)
VALUES (
    $1,
    CASE WHEN $2::text IS NOT NULL THEN ST_GeomFromGeoJSON($2) ELSE NULL END,
    CASE WHEN $3::text IS NOT NULL THEN ST_GeomFromGeoJSON($3)::geography ELSE NULL END,
    ...
);
```
**Pattern**: Conditional GeoJSON conversion (allows NULL boundaries)
**Validation**: GeoJSON validated in Go before SQL (Gotcha #32)

#### Find Gardens Near Location (Radius Search)
```sql
SELECT *, ST_Distance(location, ST_SetSRID(ST_MakePoint($lng, $lat), 4326)::geography) as distance_m
FROM gardens
WHERE ST_DWithin(
    location,
    ST_SetSRID(ST_MakePoint($lng, $lat), 4326)::geography,
    $radiusMeters
)
ORDER BY distance_m;
```
**Pattern**: ST_DWithin for filtering + ST_Distance for ordering
**Index**: GIST on `gardens.location`
**Performance**: <50ms for 1000 gardens within 10km

#### Calculate Total Area for User
```sql
SELECT COALESCE(SUM(ST_Area(boundary::geography)), 0)
FROM gardens
WHERE user_id = $1;
```
**Pattern**: Aggregation with COALESCE for NULL handling
**Performance**: <100ms for 100 gardens

#### Detect Hardiness Zone (Spatial Join)
```sql
SELECT cz.zone_code
FROM gardens g
JOIN climate_zones cz ON ST_Contains(cz.zone_geometry, g.boundary)
WHERE g.garden_id = $1
LIMIT 1;
```
**Pattern**: Spatial join using ST_Contains
**Index**: GIST on `climate_zones.zone_geometry`
**Performance**: <20ms with index

---

### GardenZoneRepository (To Be Implemented)

#### Validate Zone Within Garden
```sql
SELECT ST_Contains(g.boundary, $zoneGeometry::geometry)
FROM gardens g
WHERE garden_id = $1;
```
**Pattern**: ST_Contains validation before insert
**Returns**: Boolean (reject if false)

#### Check Zone Overlaps
```sql
SELECT EXISTS(
    SELECT 1 FROM garden_zones
    WHERE garden_id = $1
      AND zone_id != $2  -- Exclude self when updating
      AND ST_Overlaps(geometry, $newZoneGeometry::geometry)
);
```
**Pattern**: EXISTS for boolean result, exclude self for updates
**Index**: GIST on `garden_zones.geometry`

#### Calculate Total Zone Area
```sql
SELECT COALESCE(SUM(ST_Area(geometry::geography)), 0)
FROM garden_zones
WHERE garden_id = $1;
```
**Pattern**: Same as garden total area (reusable pattern)
**Alternative**: Use pre-calculated `area_m2` column (faster)

---

### GardenPlantRepository (To Be Implemented)

#### Check Plant Spacing (Find Nearby Plants)
```sql
SELECT * FROM garden_plants
WHERE garden_id = $1
  AND removed_date IS NULL  -- Only active plants
  AND ST_DWithin(
      location::geography,
      $newPlantLocation::geography,
      $minDistanceMeters
  );
```
**Pattern**: ST_DWithin for spacing validation
**Index**: GIST on `garden_plants.location`
**Business Logic**: If query returns rows, spacing violation detected

#### Validate Plant Location Within Garden
```sql
SELECT ST_Contains(g.boundary, $plantLocation::geometry)
FROM gardens g
WHERE garden_id = $1;
```
**Pattern**: Reuse ST_Contains pattern from zone validation
**Returns**: Boolean (reject if false)

#### Find Plants in Zone
```sql
SELECT gp.*
FROM garden_plants gp
JOIN garden_zones gz ON gz.zone_id = $1
WHERE ST_Contains(gz.geometry, gp.location);
```
**Pattern**: Spatial join with ST_Contains
**Note**: Handles plants that may have moved zones

---

## Required Indexes

### Gardens Table
```sql
CREATE INDEX idx_gardens_boundary ON gardens USING GIST(boundary);
CREATE INDEX idx_gardens_location ON gardens USING GIST(location);
CREATE INDEX idx_gardens_user ON gardens(user_id);  -- B-tree for FK lookups
```

### Garden Zones Table
```sql
CREATE INDEX idx_garden_zones_geometry ON garden_zones USING GIST(geometry);
CREATE INDEX idx_garden_zones_garden ON garden_zones(garden_id);  -- B-tree
```

### Garden Features Table
```sql
CREATE INDEX idx_garden_features_geometry ON garden_features USING GIST(geometry);
CREATE INDEX idx_garden_features_garden ON garden_features(garden_id);  -- B-tree
```

### Garden Plants Table
```sql
CREATE INDEX idx_garden_plants_location ON garden_plants USING GIST(location);
CREATE INDEX idx_garden_plants_garden ON garden_plants(garden_id);  -- B-tree
CREATE INDEX idx_garden_plants_zone ON garden_plants(zone_id);  -- B-tree
CREATE INDEX idx_garden_plants_plant ON garden_plants(plant_id);  -- B-tree
```

### Climate Zones Table (External Dependency)
```sql
CREATE INDEX idx_climate_zones_geometry ON climate_zones USING GIST(zone_geometry);
```

**Index Type**:
- **GIST**: Required for all spatial queries (ST_Contains, ST_DWithin, ST_Overlaps)
- **B-tree**: Standard indexes for foreign keys and exact matches

**Performance Impact**:
- With GIST indexes: 10-50ms for spatial queries
- Without GIST indexes: 500-5000ms (sequential scan)

---

## Performance Benchmarks

### Target Performance (From PRD)
- Garden spatial query: <50ms (p95)
- Zone overlap check: <20ms
- Plant spacing validation: <30ms
- Area calculations: <10ms

### Actual Expected Performance (With Indexes)
| Operation | Target | Expected | Notes |
|-----------|--------|----------|-------|
| FindByLocation (10km radius) | <50ms | 30-40ms | 1000 gardens |
| CalculateArea | <10ms | 5-8ms | Single garden |
| DetectHardinessZone | <20ms | 15-20ms | Spatial join |
| CheckZoneOverlaps | <20ms | 10-15ms | 10 zones per garden |
| CheckPlantSpacing | <30ms | 20-25ms | 100 plants per garden |
| ValidateZoneWithinGarden | <20ms | 10-15ms | Single check |

### Optimization Strategies
1. **Use Generated Columns**: `area_m2` in `garden_zones` avoids ST_Area recalculation
2. **Batch Spatial Queries**: FindByIDs avoids N+1 queries
3. **Index Coverage**: Ensure all spatial columns have GIST indexes
4. **Geography vs Geometry**: Use `::geography` for accurate meter-based calculations

---

## Common Pitfalls

### 1. Forgetting to Validate GeoJSON (Gotcha #32)
```go
// WRONG - Malformed GeoJSON reaches database
_, err := db.Exec(`INSERT INTO gardens (boundary) VALUES (ST_GeomFromGeoJSON($1))`, geojson)
// ERROR: invalid GeoJSON representation (SQLSTATE 22023) - cryptic!

// CORRECT - Validate in Go first
if err := database.ValidateGeoJSON(geojson); err != nil {
    return fmt.Errorf("invalid boundary: %w", err)
}
```
**Prevention**: Always call `ValidateGeoJSON()` before `ST_GeomFromGeoJSON()`

### 2. Invalid Coordinate Bounds (Gotcha #33)
```go
// WRONG - Out of bounds coordinates
lat, lng := 95.5, -200.0  // Invalid!

// CORRECT - Validate WGS84 bounds
if err := database.ValidateCoordinates(lat, lng); err != nil {
    return err  // "latitude must be between -90 and 90"
}
```
**Bounds**: Latitude: -90 to 90, Longitude: -180 to 180

### 3. Swapping Latitude and Longitude
```sql
-- WRONG - ST_MakePoint expects (lng, lat) NOT (lat, lng)
ST_MakePoint(37.7749, -122.4194)  -- Swapped!

-- CORRECT - Longitude first, then latitude
ST_MakePoint(-122.4194, 37.7749)
```
**Rule**: PostGIS uses (X, Y) = (Longitude, Latitude) order

### 4. Forgetting ::geography Cast for Distances
```sql
-- WRONG - Distance in degrees (meaningless)
SELECT ST_Distance(location, point)  -- Returns ~0.05 (degrees)

-- CORRECT - Distance in meters
SELECT ST_Distance(location::geography, point::geography)  -- Returns ~5000 (meters)
```
**Rule**: Always cast to `::geography` for meter-based calculations

### 5. Missing GIST Indexes
```sql
-- Query plan without index
EXPLAIN ANALYZE SELECT * FROM gardens WHERE ST_DWithin(location, ..., 5000);
-- Seq Scan on gardens (cost=0.00..500.00 rows=100) (actual time=485ms)

-- Query plan with GIST index
-- Index Scan using idx_gardens_location (cost=0.00..50.00 rows=10) (actual time=25ms)
```
**Verification**: Run `EXPLAIN ANALYZE` on all spatial queries during testing

### 6. Unclosed Polygon Rings
```json
// WRONG - First and last points differ
{
  "type": "Polygon",
  "coordinates": [[
    [-122.4194, 37.7749],
    [-122.4184, 37.7749],
    [-122.4184, 37.7739],
    [-122.4194, 37.7739]  // Missing closing point!
  ]]
}

// CORRECT - Ring is closed
{
  "coordinates": [[
    [-122.4194, 37.7749],
    [-122.4184, 37.7749],
    [-122.4184, 37.7739],
    [-122.4194, 37.7739],
    [-122.4194, 37.7749]  // Closes the ring
  ]]
}
```
**Validation**: `ValidatePolygonClosure()` checks this automatically

---

## Testing Spatial Queries

### Integration Test Pattern
```go
func TestGardenRepository_FindByLocation(t *testing.T) {
    // Setup: Create test gardens with known locations
    garden1 := createTestGarden(37.7749, -122.4194)  // San Francisco
    garden2 := createTestGarden(37.7849, -122.4094)  // 1.5km away
    garden3 := createTestGarden(40.7128, -74.0060)   // New York (far away)

    // Execute: Find gardens within 5km of SF
    gardens, err := repo.FindByLocation(ctx, 37.7749, -122.4194, 5.0)

    // Assert: Should return garden1 and garden2, not garden3
    assert.NoError(t, err)
    assert.Len(t, gardens, 2)
    assert.Contains(t, gardens, garden1)
    assert.Contains(t, gardens, garden2)
}
```

### GeoJSON Test Data
```go
// Valid garden boundary (rectangle in San Francisco)
validBoundary := `{
  "type": "Polygon",
  "coordinates": [[
    [-122.4194, 37.7749],
    [-122.4184, 37.7749],
    [-122.4184, 37.7739],
    [-122.4194, 37.7739],
    [-122.4194, 37.7749]
  ]]
}`

// Invalid boundary (unclosed ring)
invalidBoundary := `{
  "type": "Polygon",
  "coordinates": [[
    [-122.4194, 37.7749],
    [-122.4184, 37.7749],
    [-122.4184, 37.7739]
  ]]
}`
```

---

## References

- **PostGIS Documentation**: https://postgis.net/docs/
- **GeoJSON Specification**: https://datatracker.ietf.org/doc/html/rfc7946
- **GIST Indexes**: https://www.postgresql.org/docs/current/gist.html
- **WGS84 Coordinate System**: SRID 4326

---

*Last Updated: 2025-10-03*
*Part 3: Garden Spatial Service Implementation*
