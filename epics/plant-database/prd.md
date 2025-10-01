
# Plant Database & Garden Management System - PRD

## System Overview

Comprehensive plant database with spatial garden mapping capabilities. The system uses PostgreSQL with PostGIS for spatial operations, Go for backend services, and provides both GraphQL and REST APIs.

### Implementation Status

- ✅ **Part 1: Database & Core Infrastructure** - PostgreSQL 17 with PostGIS, migrations, Cloud SQL setup
- ✅ **Part 2: Plant Domain Service** - Complete DDD implementation with caching
- 🚧 **Part 3: Garden Spatial Service** - In progress
- 📋 **Part 4: Garden Analysis Engine** - Planned
- 📋 **Part 5: REST API Gateway** - Planned
- 📋 **Part 6: GraphQL Gateway** - Planned
- 📋 **Part 7: Auth & User Management** - Planned

---

## Part 2: Plant Domain Service (Implemented)

### Features Delivered

**Plant Management**
- Hierarchical taxonomy: Family → Genus → Species → Cultivar
- Full botanical name generation and validation
- Common names support (multiple per plant)
- Full-text search with relevance ranking
- Plant type classification (tree, shrub, perennial, etc.)

**Growing Conditions**
- Climate zones (USDA hardiness, heat zones)
- Sun/shade requirements (7 types: full sun to full shade)
- Water needs (very dry to aquatic)
- Soil requirements (types, drainage, pH ranges)
- Environmental tolerances (drought, salt, wind)
- Temporal data (flowering/fruiting months)
- Confidence levels for data quality

**Companion Planting**
- Beneficial, antagonistic, and neutral relationships
- Distance recommendations (optimal and maximum)
- Benefits tracking (pest control, nitrogen fixation, etc.)
- Compatibility validation between plant pairs

**Physical Characteristics**
- Mature height and spread ranges
- Growth rate classification
- Flexible JSONB traits (colors, textures, toxicity, wildlife value)

### API Endpoints (Service Layer)

```go
// Core operations
GetPlant(plantID, includeDetails) -> Plant
GetPlantWithConditions(plantID, countryID) -> Plant
SearchPlants(query, filters) -> SearchResult
FindByBotanicalName(botanicalName) -> Plant

// Taxonomy queries
FindPlantsByFamily(familyName) -> []Plant
FindPlantsByGenus(genusName) -> []Plant

// Companion planting
GetCompanionPlants(plantID, beneficialOnly) -> []Companion
GetBeneficialCompanions(plantID) -> []Companion
GetAntagonisticPlants(plantID) -> []Companion
ValidatePlantCompatibility(plantA, plantB) -> CompatibilityResult

// Recommendations
RecommendPlants(hardinessZone, sunRequirement) -> []Plant

// Management
CreatePlant(plant) -> error
```

### Performance Characteristics

- Plant retrieval: <10ms (cached), <50ms (uncached)
- Full-text search: <100ms (p95)
- Companion queries: <50ms
- Redis caching with smart TTLs (15min-2hr)
- Connection pooling (25 connections)

### Technical Implementation

**Architecture**: Domain-Driven Design with clean architecture layers
- `domain/entity/` - Business entities (Plant, Companion)
- `domain/repository/` - Data access interfaces
- `domain/service/` - Business logic
- `infrastructure/database/` - PostgreSQL implementation
- `infrastructure/cache/` - Redis caching
- `pkg/types/` - Shared value objects

**Database Integration**
- Full-text search using PostgreSQL `to_tsvector`
- Complex joins across taxonomy tables
- GIST indexes on search vectors
- Prepared statements for common queries

**Testing**
- 18 unit tests with mock repository
- 52% test coverage on service layer
- Testify/mock for isolation
- All tests passing in <3s

---

## Part 3: Garden Spatial Service (PostGIS Features)

Version 5.0 of the plant database includes full PostGIS support, enabling sophisticated spatial analysis for garden planning and management. This section covers implementation of spatial features.

## Core Spatial Tables

### 1. Gardens Table
- **boundary**: Full garden perimeter as polygon
- **location**: Center point for distance calculations
- Automatically detects climate zones based on location

### 2. Garden Zones
- Individual beds, paths, and areas within the garden
- Automatic area calculation using PostGIS
- Track sun exposure per zone

### 3. Garden Features
- Trees, buildings, structures that create shade
- Height tracking for shade calculations
- Deciduous flag for seasonal shade variation

### 4. Garden Plants
- Exact planting locations as points
- Links to zones for organization
- Spatial queries for spacing validation

## Key Spatial Features Implemented

### 1. Climate Zone Detection

```sql
-- Automatically detect hardiness zone for a garden
UPDATE gardens g
SET hardiness_zone = (
    SELECT zone_code 
    FROM climate_zones cz
    WHERE ST_Contains(cz.zone_geometry, g.location::geometry)
    AND zone_system = 'USDA'
    LIMIT 1
)
WHERE g.garden_id = 'your-garden-id';
```

### 2. Shade Calculation

```sql
-- Calculate shade areas from trees and buildings
SELECT * FROM calculate_shade_zones('garden-id', 45);
-- Returns polygons representing shaded areas
-- Angle parameter represents sun angle (varies by season/time)
```

### 3. Plant Spacing Validation

```sql
-- Check if plants are properly spaced
SELECT * FROM check_plant_spacing('garden-id')
WHERE NOT spacing_ok;
-- Returns pairs of plants that are too close together
```

### 4. Optimal Planting Location

```sql
-- Find suitable spots for a new plant
SELECT 
    ST_AsGeoJSON(suitable_area) AS area_json,
    area_m2
FROM find_planting_spots('garden-id', 'plant-id', 1.0)
WHERE area_m2 > 2;  -- At least 2 square meters
```

### 5. Garden Zone Management

```sql
-- Create a new garden bed
INSERT INTO garden_zones (garden_id, zone_name, zone_type, geometry)
VALUES (
    'garden-id',
    'Tomato Bed',
    'bed',
    ST_GeomFromGeoJSON('{
        "type": "Polygon",
        "coordinates": [[[0,0],[0,5],[10,5],[10,0],[0,0]]]
    }')
);
```

## Data Input Requirements

### Minimum Required Data

1. **Garden Boundary**: Users draw their garden perimeter
2. **Plant Locations**: Click/tap to place plants on map
3. **Major Features**: Mark trees and buildings

### Optional Enhanced Data

1. **Feature Heights**: For accurate shade calculation
2. **Zone Boundaries**: Define individual beds
3. **Soil Amendments**: Track per-zone improvements

## Mobile App Integration

### React Native Example

```javascript
// Capture garden boundary
const captureGardenBoundary = (coordinates) => {
  const polygon = {
    type: 'Polygon',
    coordinates: [coordinates]
  };
  
  // Send to backend
  api.post('/gardens', {
    boundary: polygon,
    location: calculateCentroid(coordinates)
  });
};

// Add a plant to garden
const addPlant = (lat, lng, plantId) => {
  api.post('/garden-plants', {
    location: { type: 'Point', coordinates: [lng, lat] },
    plant_id: plantId,
    garden_id: currentGardenId
  });
};
```

### Web Interface Example

```javascript
// Using Leaflet.js for web mapping
map.on('click', (e) => {
  // Add plant at clicked location
  L.marker([e.latlng.lat, e.latlng.lng])
    .addTo(plantsLayer);
    
  savePlantLocation(e.latlng.lat, e.latlng.lng);
});

// Draw garden zones
const drawnItems = new L.FeatureGroup();
map.addLayer(drawnItems);

const drawControl = new L.Control.Draw({
  edit: { featureGroup: drawnItems },
  draw: {
    polygon: true,
    rectangle: true,
    circle: false,
    marker: true
  }
});
```

## Performance Considerations

### Spatial Indexes

All spatial columns have GIST indexes created automatically:
- `idx_gardens_boundary`
- `idx_garden_zones_geometry`
- `idx_garden_plants_location`

### Query Optimization

```sql
-- Use ST_DWithin for radius searches (uses spatial index)
-- GOOD: Uses spatial index
SELECT * FROM garden_plants
WHERE ST_DWithin(location::geography, point::geography, 100);

-- BAD: Doesn't use spatial index efficiently
SELECT * FROM garden_plants
WHERE ST_Distance(location::geography, point::geography) < 100;
```

### Caching Strategy

Use materialized views for expensive calculations:
```sql
-- Refresh garden statistics periodically
REFRESH MATERIALIZED VIEW CONCURRENTLY garden_stats;
```

## Common Spatial Queries

### Find Nearby Gardens

```sql
-- Find gardens within 5km
SELECT 
    g.garden_name,
    ST_Distance(g.location, my_location) / 1000 AS distance_km
FROM gardens g
WHERE ST_DWithin(
    g.location,
    ST_GeogFromText('POINT(-71.06 42.36)'),
    5000  -- meters
)
ORDER BY distance_km;
```

### Calculate Garden Metrics

```sql
-- Get comprehensive garden statistics
SELECT 
    ST_Area(boundary::geography) AS total_area_m2,
    ST_Perimeter(boundary::geography) AS perimeter_m,
    (SELECT COUNT(*) FROM garden_plants WHERE garden_id = g.garden_id) AS plant_count,
    (SELECT SUM(ST_Area(geometry::geography)) 
     FROM garden_zones 
     WHERE garden_id = g.garden_id AND zone_type = 'bed') AS cultivated_area_m2
FROM gardens g
WHERE garden_id = 'your-garden-id';
```

### Companion Planting Analysis

```sql
-- Find companion plants already in garden
WITH my_plants AS (
    SELECT DISTINCT plant_id, location
    FROM garden_plants
    WHERE garden_id = 'your-garden-id'
)
SELECT 
    p1.plant_id AS plant,
    p2.plant_id AS companion,
    cr.relationship_type,
    ST_Distance(p1.location::geography, p2.location::geography) AS distance_m
FROM my_plants p1
JOIN companion_relationships cr ON p1.plant_id = cr.plant_a_id
JOIN my_plants p2 ON cr.plant_b_id = p2.plant_id;
```

## Migration from Non-Spatial

If migrating from v4.0 (without PostGIS):

```sql
-- Enable PostGIS
CREATE EXTENSION IF NOT EXISTS postgis;

-- Add spatial columns to existing tables
ALTER TABLE gardens 
ADD COLUMN boundary GEOMETRY(Polygon, 4326),
ADD COLUMN location GEOGRAPHY(Point, 4326);

-- Convert lat/lng to spatial points
UPDATE gardens 
SET location = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

-- Create spatial indexes
CREATE INDEX idx_gardens_location ON gardens USING GIST(location);
```

## Hosting Requirements

### Database Size Estimates

- PostGIS extension: +50MB
- Spatial indexes: ~20% of data size
- Climate zone polygons: ~100MB for USA
- Per garden: ~1KB boundary + zones

### Recommended Specifications

**Development/Small Scale (< 1000 users)**
- 2 vCPU
- 4GB RAM
- 20GB storage
- Cost: ~$40/month

**Production (1000-10000 users)**
- 4 vCPU
- 8GB RAM
- 100GB storage
- Cost: ~$100/month

**Large Scale (10000+ users)**
- 8+ vCPU
- 16GB+ RAM
- 500GB+ storage
- Consider read replicas
- Cost: $300+/month

## Testing Spatial Features

### Sample Test Data

```sql
-- Create test garden
INSERT INTO gardens (user_id, garden_name, boundary, location)
VALUES (
    'test-user-id',
    'Test Garden',
    ST_GeomFromText('POLYGON((0 0, 0 10, 10 10, 10 0, 0 0))', 4326),
    ST_GeogFromText('POINT(5 5)')
);

-- Add test zones
INSERT INTO garden_zones (garden_id, zone_name, zone_type, geometry)
VALUES 
    ('garden-id', 'North Bed', 'bed', 
     ST_GeomFromText('POLYGON((1 6, 1 9, 4 9, 4 6, 1 6))', 4326)),
    ('garden-id', 'South Bed', 'bed',
     ST_GeomFromText('POLYGON((1 1, 1 4, 4 4, 4 1, 1 1))', 4326));

-- Add test plants
INSERT INTO garden_plants (garden_id, plant_id, location)
VALUES 
    ('garden-id', 'tomato-id', ST_GeomFromText('POINT(2 2)', 4326)),
    ('garden-id', 'basil-id', ST_GeomFromText('POINT(2.5 2)', 4326));
```

### Validation Queries

```sql
-- Verify spatial relationships
SELECT 
    'Garden contains all zones' AS test,
    bool_and(ST_Contains(g.boundary, z.geometry)) AS passes
FROM gardens g
JOIN garden_zones z ON g.garden_id = z.garden_id
GROUP BY g.garden_id;

-- Check spatial index usage
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM garden_plants
WHERE ST_DWithin(location::geography, ST_GeogFromText('POINT(0 0)'), 1000);
```

## Troubleshooting

### Common Issues

1. **Slow spatial queries**: Check for missing GIST indexes
2. **Incorrect distances**: Ensure using geography type for accurate measurements
3. **Invalid geometries**: Use ST_IsValid() to check, ST_MakeValid() to fix
4. **Projection issues**: Always use SRID 4326 for GPS coordinates

### Performance Monitoring

```sql
-- Find slow spatial queries
SELECT 
    query,
    calls,
    mean_exec_time,
    total_exec_time
FROM pg_stat_statements
WHERE query LIKE '%ST_%'
ORDER BY mean_exec_time DESC
LIMIT 10;
```

## Next Steps (Phase 3 Considerations)

While not implemented in v5.0, consider these future enhancements:

1. **3D Analysis**: Use PostGIS 3D features for slope/aspect
2. **Temporal Analysis**: Track garden changes over seasons
3. **Weather Integration**: Overlay weather data on garden maps
4. **Community Features**: Share garden designs spatially
5. **Augmented Reality**: Use spatial data for AR plant placement

## Conclusion

This PostGIS implementation provides powerful spatial capabilities while remaining manageable for small to medium-scale deployments. The schema balances functionality with practicality, avoiding the complexity trap that the scientific reviewer's full recommendations would create.
