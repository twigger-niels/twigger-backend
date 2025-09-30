# Spatial Queries and PostGIS Usage

## Overview

The plant database leverages PostGIS extensively for spatial operations, garden mapping, climate zone detection, and plant placement optimization. This document covers all spatial functionality and provides practical query examples.

## PostGIS Extensions

### Required Extensions
```sql
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;
CREATE EXTENSION IF NOT EXISTS btree_gist;
```

### Extension Verification
```sql
-- Check PostGIS version
SELECT PostGIS_Version();

-- List all spatial functions
SELECT proname, prosrc
FROM pg_proc
WHERE proname LIKE 'ST_%'
ORDER BY proname;
```

## Spatial Data Types

### Geometry vs Geography

**Geometry**: Planar coordinate system, faster but less accurate for global calculations
```sql
-- Create a point in WGS84 (SRID 4326)
SELECT ST_GeomFromText('POINT(-71.06 42.36)', 4326);
```

**Geography**: Spherical coordinate system, accurate for global calculations
```sql
-- Create a geographic point
SELECT ST_GeogFromText('POINT(-71.06 42.36)');
```

### Common SRID (Spatial Reference System Identifier)
- **4326**: WGS84 (World Geodetic System 1984) - Global GPS coordinates
- **3857**: Web Mercator - Used by web mapping services

## Core Spatial Operations

### 1. Distance Calculations

**Geographic Distance (accurate for global calculations):**
```sql
-- Distance between two points in meters
SELECT ST_Distance(
    ST_Point(-71.06, 42.36)::geography,  -- Boston
    ST_Point(-74.00, 40.71)::geography   -- New York
) AS distance_meters;

-- Find gardens within 50km of a location
SELECT g.garden_name,
       ST_Distance(g.location, ST_Point(-71.06, 42.36)::geography) / 1000 AS distance_km
FROM gardens g
WHERE ST_DWithin(
    g.location,
    ST_Point(-71.06, 42.36)::geography,
    50000  -- 50km in meters
)
ORDER BY distance_km;
```

**Planar Distance (faster for local calculations):**
```sql
-- Distance between garden zones
SELECT
    z1.zone_name AS zone1,
    z2.zone_name AS zone2,
    ST_Distance(z1.geometry, z2.geometry) AS distance_degrees
FROM garden_zones z1, garden_zones z2
WHERE z1.garden_id = z2.garden_id
  AND z1.zone_id != z2.zone_id;
```

### 2. Containment and Intersection

**Point in Polygon:**
```sql
-- Check which climate zone contains a garden
SELECT cz.zone_system, cz.zone_code
FROM climate_zones cz, gardens g
WHERE ST_Contains(cz.zone_geometry, g.location::geometry)
  AND g.garden_id = 'your-garden-uuid';

-- Find all plants within a specific garden zone
SELECT gp.*, p.full_botanical_name
FROM garden_plants gp
JOIN plants p ON gp.plant_id = p.plant_id
JOIN garden_zones gz ON gp.zone_id = gz.zone_id
WHERE ST_Contains(gz.geometry, gp.location)
  AND gz.zone_id = 'your-zone-uuid';
```

**Polygon Intersection:**
```sql
-- Find overlapping garden zones (should not exist)
SELECT
    z1.zone_name AS zone1,
    z2.zone_name AS zone2,
    ST_Area(ST_Intersection(z1.geometry, z2.geometry)) AS overlap_area
FROM garden_zones z1, garden_zones z2
WHERE z1.garden_id = z2.garden_id
  AND z1.zone_id < z2.zone_id  -- Avoid duplicates
  AND ST_Intersects(z1.geometry, z2.geometry);
```

### 3. Area and Perimeter Calculations

**Garden and Zone Areas:**
```sql
-- Garden area with zone breakdown
SELECT
    g.garden_name,
    ST_Area(g.boundary::geography) / 10000 AS garden_area_hectares,
    COUNT(gz.zone_id) AS num_zones,
    SUM(gz.area_m2) AS total_zone_area_m2,
    ST_Area(g.boundary::geography) - SUM(gz.area_m2) AS unzoned_area_m2
FROM gardens g
LEFT JOIN garden_zones gz ON g.garden_id = gz.garden_id
GROUP BY g.garden_id, g.garden_name, g.boundary;
```

**Zone Utilization:**
```sql
-- Calculate zone utilization by plant coverage
SELECT
    gz.zone_name,
    gz.area_m2 AS zone_area,
    COUNT(gp.garden_plant_id) AS num_plants,
    SUM(ST_Area(gp.planting_area::geography)) AS planted_area_m2,
    ROUND(
        (SUM(ST_Area(gp.planting_area::geography)) / gz.area_m2) * 100, 2
    ) AS utilization_percent
FROM garden_zones gz
LEFT JOIN garden_plants gp ON gz.zone_id = gp.zone_id
WHERE gp.removed_date IS NULL
GROUP BY gz.zone_id, gz.zone_name, gz.area_m2;
```

### 4. Buffer Operations

**Plant Spacing Analysis:**
```sql
-- Create buffers around plants to check spacing
WITH plant_buffers AS (
    SELECT
        gp.garden_plant_id,
        gp.plant_id,
        ST_Buffer(gp.location::geography, 1.0) AS buffer_1m  -- 1 meter buffer
    FROM garden_plants gp
    WHERE gp.removed_date IS NULL
)
SELECT
    pb1.garden_plant_id AS plant1,
    pb2.garden_plant_id AS plant2,
    'Too close' AS spacing_issue
FROM plant_buffers pb1, plant_buffers pb2
WHERE pb1.garden_plant_id < pb2.garden_plant_id
  AND ST_Intersects(pb1.buffer_1m, pb2.buffer_1m);
```

**Shade Zone Calculation:**
```sql
-- Calculate shade areas from garden features
SELECT
    gf.feature_name,
    gf.feature_type,
    ST_Buffer(
        gf.geometry,
        COALESCE(gf.height_m, 0) * TAN(RADIANS(45))  -- 45-degree sun angle
    ) AS shade_area
FROM garden_features gf
WHERE gf.feature_type IN ('tree', 'building', 'fence')
  AND gf.height_m IS NOT NULL;
```

## Advanced Spatial Functions

### 1. Climate Zone Detection

```sql
-- Function to detect climate zone for a location
CREATE OR REPLACE FUNCTION detect_climate_zone(
    p_location GEOGRAPHY,
    p_country_id UUID
) RETURNS TABLE (
    zone_system VARCHAR(20),
    zone_code VARCHAR(10)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cz.zone_system,
        cz.zone_code
    FROM climate_zones cz
    WHERE cz.country_id = p_country_id
      AND ST_Contains(cz.zone_geometry, p_location::geometry);
END;
$$ LANGUAGE plpgsql;

-- Usage example
SELECT * FROM detect_climate_zone(
    ST_Point(-71.06, 42.36)::geography,  -- Boston coordinates
    (SELECT country_id FROM countries WHERE country_code = 'US')
);
```

### 2. Optimal Planting Location Finder

```sql
-- Function to find optimal planting spots
CREATE OR REPLACE FUNCTION find_planting_spots(
    p_garden_id UUID,
    p_plant_id UUID,
    p_min_spacing_m NUMERIC DEFAULT 1.0
) RETURNS TABLE (
    zone_id UUID,
    zone_name VARCHAR(100),
    suitable_area GEOMETRY,
    area_m2 NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    WITH existing_plants AS (
        SELECT ST_Union(ST_Buffer(gp.location::geography, p_min_spacing_m)) AS occupied_area
        FROM garden_plants gp
        WHERE gp.garden_id = p_garden_id
          AND gp.removed_date IS NULL
    )
    SELECT
        gz.zone_id,
        gz.zone_name,
        ST_Difference(
            gz.geometry,
            COALESCE(ep.occupied_area::geometry, ST_GeomFromText('GEOMETRYCOLLECTION EMPTY'))
        ) AS suitable_area,
        ST_Area(ST_Difference(
            gz.geometry,
            COALESCE(ep.occupied_area::geometry, ST_GeomFromText('GEOMETRYCOLLECTION EMPTY'))
        )::geography) AS area_m2
    FROM garden_zones gz
    LEFT JOIN existing_plants ep ON true
    WHERE gz.garden_id = p_garden_id
      AND gz.zone_type = 'bed';
END;
$$ LANGUAGE plpgsql;
```

### 3. Plant Spacing Verification

```sql
-- Function to check plant spacing compliance
CREATE OR REPLACE FUNCTION check_plant_spacing(
    p_garden_id UUID
) RETURNS TABLE (
    plant1_id UUID,
    plant2_id UUID,
    distance_m NUMERIC,
    recommended_spacing_m NUMERIC,
    spacing_ok BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    WITH plant_pairs AS (
        SELECT
            gp1.garden_plant_id AS plant1_id,
            gp2.garden_plant_id AS plant2_id,
            ST_Distance(gp1.location::geography, gp2.location::geography) AS distance_m,
            GREATEST(
                COALESCE((pc1.mature_spread).typical_m, 1.0) / 2 +
                COALESCE((pc2.mature_spread).typical_m, 1.0) / 2,
                1.0
            ) AS recommended_spacing_m
        FROM garden_plants gp1
        JOIN garden_plants gp2 ON gp1.garden_id = gp2.garden_id
        LEFT JOIN physical_characteristics pc1 ON gp1.plant_id = pc1.plant_id
        LEFT JOIN physical_characteristics pc2 ON gp2.plant_id = pc2.plant_id
        WHERE gp1.garden_id = p_garden_id
          AND gp1.garden_plant_id < gp2.garden_plant_id
          AND gp1.removed_date IS NULL
          AND gp2.removed_date IS NULL
    )
    SELECT
        pp.plant1_id,
        pp.plant2_id,
        pp.distance_m,
        pp.recommended_spacing_m,
        pp.distance_m >= pp.recommended_spacing_m AS spacing_ok
    FROM plant_pairs pp
    WHERE pp.distance_m < pp.recommended_spacing_m * 1.5; -- Only show potential issues
END;
$$ LANGUAGE plpgsql;
```

## Spatial Queries by Use Case

### 1. Garden Management

**Garden Overview:**
```sql
-- Complete garden spatial analysis
SELECT
    g.garden_name,
    ST_Area(g.boundary::geography) AS garden_area_m2,
    ST_Perimeter(g.boundary::geography) AS garden_perimeter_m,
    g.elevation_m,
    g.slope_degrees,
    g.aspect,
    COUNT(DISTINCT gz.zone_id) AS num_zones,
    COUNT(DISTINCT gp.garden_plant_id) AS total_plants,
    COUNT(DISTINCT gp.plant_id) AS unique_plant_species
FROM gardens g
LEFT JOIN garden_zones gz ON g.garden_id = gz.garden_id
LEFT JOIN garden_plants gp ON g.garden_id = gp.garden_id AND gp.removed_date IS NULL
WHERE g.garden_id = 'your-garden-uuid'
GROUP BY g.garden_id, g.garden_name, g.boundary, g.elevation_m, g.slope_degrees, g.aspect;
```

**Zone Microclimate Analysis:**
```sql
-- Analyze microclimates within zones
SELECT
    gz.zone_name,
    gz.sun_hours_summer,
    gz.sun_hours_winter,
    gz.soil_type,
    gz.irrigation_type,
    -- Calculate zone aspect from geometry centroid
    CASE
        WHEN ST_Azimuth(ST_Centroid(gz.geometry), ST_PointN(ST_ExteriorRing(gz.geometry), 1)) BETWEEN 0 AND PI()/4 THEN 'N'
        WHEN ST_Azimuth(ST_Centroid(gz.geometry), ST_PointN(ST_ExteriorRing(gz.geometry), 1)) BETWEEN PI()/4 AND PI()/2 THEN 'NE'
        -- ... additional cases
        ELSE 'Unknown'
    END AS calculated_aspect
FROM garden_zones gz
WHERE gz.garden_id = 'your-garden-uuid';
```

### 2. Plant Distribution Analysis

**Plant Density Mapping:**
```sql
-- Calculate plant density per zone
SELECT
    gz.zone_name,
    gz.area_m2,
    COUNT(gp.garden_plant_id) AS plant_count,
    ROUND(COUNT(gp.garden_plant_id) / (gz.area_m2 / 10000), 2) AS plants_per_hectare,
    -- Create density categories
    CASE
        WHEN COUNT(gp.garden_plant_id) / (gz.area_m2 / 10000) < 100 THEN 'Low Density'
        WHEN COUNT(gp.garden_plant_id) / (gz.area_m2 / 10000) < 500 THEN 'Medium Density'
        ELSE 'High Density'
    END AS density_category
FROM garden_zones gz
LEFT JOIN garden_plants gp ON gz.zone_id = gp.zone_id AND gp.removed_date IS NULL
GROUP BY gz.zone_id, gz.zone_name, gz.area_m2;
```

**Species Distribution:**
```sql
-- Map plant species distribution across the garden
SELECT
    p.full_botanical_name,
    COUNT(gp.garden_plant_id) AS total_planted,
    COUNT(DISTINCT gz.zone_id) AS zones_present,
    ST_Collect(gp.location) AS distribution_points,
    ST_ConvexHull(ST_Collect(gp.location)) AS distribution_area
FROM garden_plants gp
JOIN plants p ON gp.plant_id = p.plant_id
JOIN garden_zones gz ON gp.zone_id = gz.zone_id
WHERE gp.garden_id = 'your-garden-uuid'
  AND gp.removed_date IS NULL
GROUP BY p.plant_id, p.full_botanical_name
HAVING COUNT(gp.garden_plant_id) > 1;
```

### 3. Environmental Analysis

**Sun Exposure Analysis:**
```sql
-- Calculate daily sun patterns
WITH shade_analysis AS (
    SELECT
        gz.zone_id,
        gz.zone_name,
        -- Calculate potential shade from garden features
        ST_Union(
            ST_Buffer(
                gf.geometry,
                gf.height_m * TAN(RADIANS(30))  -- Low sun angle
            )
        ) AS potential_shade_area
    FROM garden_zones gz
    LEFT JOIN garden_features gf ON gz.garden_id = gf.garden_id
    WHERE gf.feature_type IN ('tree', 'building')
      AND gf.height_m > 2
    GROUP BY gz.zone_id, gz.zone_name
)
SELECT
    sa.zone_name,
    CASE
        WHEN sa.potential_shade_area IS NULL THEN 'Full Sun'
        WHEN ST_Area(ST_Intersection(gz.geometry, sa.potential_shade_area)) / gz.area_m2 < 0.25 THEN 'Mostly Sun'
        WHEN ST_Area(ST_Intersection(gz.geometry, sa.potential_shade_area)) / gz.area_m2 < 0.75 THEN 'Partial Shade'
        ELSE 'Mostly Shade'
    END AS sun_exposure_category
FROM shade_analysis sa
JOIN garden_zones gz ON sa.zone_id = gz.zone_id;
```

**Drainage Analysis:**
```sql
-- Identify potential drainage issues based on elevation
SELECT
    gz.zone_name,
    ST_Z(ST_PointN(ST_ExteriorRing(gz.geometry), 1)) AS min_elevation,
    ST_Z(ST_PointN(ST_ExteriorRing(gz.geometry), ST_NPoints(ST_ExteriorRing(gz.geometry)))) AS max_elevation,
    CASE
        WHEN ABS(ST_Z(ST_PointN(ST_ExteriorRing(gz.geometry), 1)) -
                 ST_Z(ST_PointN(ST_ExteriorRing(gz.geometry), ST_NPoints(ST_ExteriorRing(gz.geometry))))) < 0.5
        THEN 'Level - May have drainage issues'
        ELSE 'Sloped - Good drainage'
    END AS drainage_assessment
FROM garden_zones gz
WHERE gz.garden_id = 'your-garden-uuid';
```

## Performance Optimization

### Spatial Indexes

**Essential Spatial Indexes:**
```sql
-- Ensure all spatial columns have GIST indexes
CREATE INDEX CONCURRENTLY idx_countries_boundary ON countries USING GIST(country_boundary);
CREATE INDEX CONCURRENTLY idx_climate_zones_geometry ON climate_zones USING GIST(zone_geometry);
CREATE INDEX CONCURRENTLY idx_gardens_boundary ON gardens USING GIST(boundary);
CREATE INDEX CONCURRENTLY idx_gardens_location ON gardens USING GIST(location);
CREATE INDEX CONCURRENTLY idx_garden_zones_geometry ON garden_zones USING GIST(geometry);
CREATE INDEX CONCURRENTLY idx_garden_features_geometry ON garden_features USING GIST(geometry);
CREATE INDEX CONCURRENTLY idx_garden_plants_location ON garden_plants USING GIST(location);
```

### Query Optimization Tips

1. **Use appropriate SRID**: Always specify SRID for consistency
2. **Geography vs Geometry**: Use geography for global calculations, geometry for local
3. **Buffer operations**: Consider buffering smaller geometries, not larger ones
4. **Spatial joins**: Always have spatial indexes on both sides of spatial joins
5. **ST_DWithin vs ST_Distance**: Use ST_DWithin for distance filters (it's optimized)

### Performance Monitoring

```sql
-- Check spatial query performance
EXPLAIN (ANALYZE, BUFFERS)
SELECT COUNT(*)
FROM garden_plants gp
JOIN garden_zones gz ON ST_Contains(gz.geometry, gp.location)
WHERE gz.garden_id = 'your-garden-uuid';

-- Monitor index usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE indexname LIKE '%gist%'
ORDER BY idx_scan DESC;
```

## Integration with Application

### Go/PostGIS Integration

**Sample Go code for spatial queries:**
```go
// Example: Find plants within radius
func FindPlantsWithinRadius(ctx context.Context, db *pgxpool.Pool, centerLat, centerLon, radiusMeters float64) ([]PlantLocation, error) {
    query := `
        SELECT
            gp.garden_plant_id,
            p.full_botanical_name,
            ST_X(gp.location) as longitude,
            ST_Y(gp.location) as latitude,
            ST_Distance(gp.location::geography, ST_Point($2, $1)::geography) as distance_meters
        FROM garden_plants gp
        JOIN plants p ON gp.plant_id = p.plant_id
        WHERE ST_DWithin(
            gp.location::geography,
            ST_Point($2, $1)::geography,
            $3
        )
        AND gp.removed_date IS NULL
        ORDER BY distance_meters;
    `

    rows, err := db.Query(ctx, query, centerLat, centerLon, radiusMeters)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []PlantLocation
    for rows.Next() {
        var pl PlantLocation
        err := rows.Scan(&pl.ID, &pl.BotanicalName, &pl.Longitude, &pl.Latitude, &pl.Distance)
        if err != nil {
            return nil, err
        }
        results = append(results, pl)
    }

    return results, nil
}
```

This comprehensive spatial documentation provides the foundation for all garden mapping, climate analysis, and plant placement optimization features in the plant database system.